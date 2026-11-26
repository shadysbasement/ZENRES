package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/shadysbasement/zevra/services/auth-api/internal/pgp"
	"github.com/shadysbasement/zevra/services/auth-api/internal/store"
	"github.com/shadysbasement/zevra/services/internal/platform"
)

func TestSupportCustomerIsolation(t *testing.T) {
	dsn := os.Getenv("SUPPORT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set SUPPORT_TEST_DATABASE_URL for disposable migrated PostgreSQL")
	}
	if !strings.Contains(dsn, "test") {
		t.Fatal("refusing non-test database")
	}
	ctx := context.Background()
	pg, err := store.NewPostgres(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pg.DB.Close()
	s := &Server{Cfg: Config{InternalSecret: "internal"}, Users: pg, SupportDB: pg.DB, Limiter: platform.NewMemLimiter(), Now: time.Now}
	h := s.Routes()
	makeUser := func(name string) (uuid.UUID, string) {
		u := &store.User{Email: uuid.NewString() + "@example.test", FullName: name, AuthOrigin: "manual", EmailVerified: true, Status: "active"}
		if err := pg.CreateUser(ctx, u); err != nil {
			t.Fatal(err)
		}
		token := platform.RandomToken(32)
		if err := pg.CreateSession(ctx, &store.Session{UserID: u.ID, TokenHash: platform.SHA256(token), ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
		return u.ID, token
	}
	_, a := makeUser("A")
	_, b := makeUser("B")
	do := func(method, path, token string, body any) *httptest.ResponseRecorder {
		var raw []byte
		if body != nil {
			raw, _ = json.Marshal(body)
		}
		r := httptest.NewRequest(method, path, bytes.NewReader(raw))
		r.AddCookie(&http.Cookie{Name: SessionCookie, Value: token})
		r.AddCookie(&http.Cookie{Name: CSRFCookie, Value: "csrf"})
		r.Header.Set("X-CSRF-Token", "csrf")
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	create := func(token string) string {
		w := do("POST", "/api/v1/support/tickets", token, map[string]any{"kind": "question", "category": "platform", "subject": "Need help with a workflow", "message": "Please explain this workflow."})
		if w.Code != 201 {
			t.Fatalf("create status %d: %s", w.Code, w.Body.String())
		}
		var out struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out.ID
	}
	aID := create(a)
	bID := create(b)
	list := do("GET", "/api/v1/support/tickets", a, nil)
	if list.Code != 200 || !strings.Contains(list.Body.String(), aID) || strings.Contains(list.Body.String(), bID) {
		t.Fatalf("list isolation: %d %s", list.Code, list.Body.String())
	}
	for _, path := range []string{"/api/v1/support/tickets/" + bID} {
		if w := do("GET", path, a, nil); w.Code != 404 {
			t.Fatalf("cross customer detail %d", w.Code)
		}
	}
	if w := do("POST", "/api/v1/support/tickets/"+bID+"/messages", a, map[string]any{"body": "Not my ticket"}); w.Code != 404 {
		t.Fatalf("cross customer reply %d", w.Code)
	}
	if w := do("POST", "/api/v1/support/tickets/"+aID+"/feedback", a, map[string]any{"rating": 5}); w.Code != 409 {
		t.Fatalf("premature feedback %d", w.Code)
	}
	if _, err := pg.DB.Exec(ctx, `UPDATE support_tickets SET status='resolved',resolved_at=NOW() WHERE id=$1`, aID); err != nil {
		t.Fatal(err)
	}
	if w := do("POST", "/api/v1/support/tickets/"+aID+"/feedback", a, map[string]any{"rating": 5, "comment": "Helpful"}); w.Code != 200 {
		t.Fatalf("feedback %d: %s", w.Code, w.Body.String())
	}
	if w := do("POST", "/api/v1/support/tickets/"+aID+"/messages", a, map[string]any{"body": "Too late"}); w.Code != 409 {
		t.Fatalf("resolved reply %d", w.Code)
	}
	if w := do("POST", "/api/v1/support/tickets", a, map[string]any{"kind": "callback", "category": "platform", "subject": "Please call back", "message": "A call would help me understand.", "callback_at": time.Now().Add(3 * time.Hour).UTC().Format(time.RFC3339), "callback_timezone": "Asia/Kolkata"}); w.Code != 201 {
		t.Fatalf("callback %d: %s", w.Code, w.Body.String())
	}
	if w := do("GET", "/internal/support/tickets", a, nil); w.Code != 404 {
		t.Fatalf("internal route without secret %d", w.Code)
	}
	// PGP-session RBAC: managers see the queue; agents see only assigned tickets.
	mem := store.NewMemory()
	vault, err := pgp.New(mem, "zenres.example")
	if err != nil {
		t.Fatal(err)
	}
	s.Vault = vault
	makeAdmin := func(handle, role string) (uuid.UUID, string) {
		key := &store.AdminKey{ID: uuid.New(), Handle: handle, Fingerprint: strings.Repeat("A", 32) + uuid.NewString()[:8], ArmoredPublicKey: "test-public-key", Role: role, Status: "active"}
		if err := mem.InsertAdminKey(ctx, key); err != nil {
			t.Fatal(err)
		}
		if _, err := pg.DB.Exec(ctx, `INSERT INTO admin_pgp_keys(id,handle,fingerprint,armored_public_key,role) VALUES($1,$2,$3,$4,$5)`, key.ID, key.Handle, key.Fingerprint, key.ArmoredPublicKey, key.Role); err != nil {
			t.Fatal(err)
		}
		token := platform.RandomToken(32)
		if err := mem.CreateAdminSession(ctx, &store.AdminSession{AdminKeyID: key.ID, TokenHash: platform.SHA256(token), IPUAHash: pgp.IPUAHash("127.0.0.1", "support-test"), ExpiresAt: time.Now().Add(time.Hour), LastSeenAt: time.Now()}); err != nil {
			t.Fatal(err)
		}
		return key.ID, token
	}
	_, manager := makeAdmin("support-manager", "support_admin")
	agentID, agent := makeAdmin("support-agent", "support_agent")
	doAdmin := func(method, path, token string, body any, ip string) *httptest.ResponseRecorder {
		var raw []byte
		if body != nil {
			raw, _ = json.Marshal(body)
		}
		r := httptest.NewRequest(method, path, bytes.NewReader(raw))
		r.Header.Set("X-Internal-Secret", "internal")
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("X-Client-IP", ip)
		r.Header.Set("X-Client-UA", "support-test")
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := doAdmin("GET", "/internal/support/tickets", manager, nil, "127.0.0.1"); w.Code != 200 || !strings.Contains(w.Body.String(), aID) || !strings.Contains(w.Body.String(), bID) {
		t.Fatalf("manager list %d: %s", w.Code, w.Body.String())
	}
	if w := doAdmin("POST", "/internal/support/tickets/"+aID, manager, map[string]any{"assignee_id": agentID.String(), "message": "We are reviewing this.", "internal_note": true}, "127.0.0.1"); w.Code != 200 {
		t.Fatalf("manager assignment %d: %s", w.Code, w.Body.String())
	}
	if w := doAdmin("GET", "/internal/support/tickets", agent, nil, "127.0.0.1"); w.Code != 200 || !strings.Contains(w.Body.String(), aID) || strings.Contains(w.Body.String(), bID) {
		t.Fatalf("agent scope %d: %s", w.Code, w.Body.String())
	}
	if w := doAdmin("GET", "/internal/support/tickets/"+bID, agent, nil, "127.0.0.1"); w.Code != 404 {
		t.Fatalf("agent cross-ticket read %d", w.Code)
	}
	if w := doAdmin("POST", "/internal/support/tickets/"+aID, manager, map[string]any{"work_minutes": 15, "work_cost": "1.25", "work_currency": "USD", "work_note": "Reviewed issue"}, "127.0.0.1"); w.Code != 200 {
		t.Fatalf("cost log %d: %s", w.Code, w.Body.String())
	}
	if w := doAdmin("GET", "/internal/support/costs", manager, nil, "127.0.0.1"); w.Code != 200 || !strings.Contains(w.Body.String(), "1.25") {
		t.Fatalf("cost summary %d: %s", w.Code, w.Body.String())
	}
	if w := doAdmin("GET", "/internal/support/costs", agent, nil, "127.0.0.1"); w.Code != 403 {
		t.Fatalf("agent cost access %d", w.Code)
	}
	if w := doAdmin("POST", "/internal/support/tickets/"+bID, agent, map[string]any{"message": "Unauthorized reply"}, "127.0.0.1"); w.Code != 404 {
		t.Fatalf("agent cross-ticket write %d", w.Code)
	}
	if w := doAdmin("GET", "/internal/support/tickets", agent, nil, "203.0.113.10"); w.Code != 404 {
		t.Fatalf("IP-bound session %d", w.Code)
	}
	if w := do("GET", "/api/v1/support/tickets/"+aID, a, nil); strings.Contains(w.Body.String(), "We are reviewing this.") {
		t.Fatal("internal note leaked to customer")
	}
	if w := doAdmin("POST", "/internal/support/announcements", manager, map[string]any{"title": "Planned maintenance", "body": "A scheduled check is in progress.", "state": "maintenance"}, "127.0.0.1"); w.Code != 201 {
		t.Fatalf("announcement %d: %s", w.Code, w.Body.String())
	}
	if w := do("GET", "/api/v1/support/announcements", a, nil); w.Code != 200 || !strings.Contains(w.Body.String(), "Planned maintenance") {
		t.Fatalf("public notice %d: %s", w.Code, w.Body.String())
	}
}
