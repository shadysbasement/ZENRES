package twin

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBroadcastReachesAllTabs(t *testing.T) {
	b := NewBroker()
	org := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { b.Serve(w, r, org) }))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var readers []*bufio.Reader
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL, nil)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		rd := bufio.NewReader(res.Body)
		rd.ReadString('\n')
		rd.ReadString('\n')
		readers = append(readers, rd)
	}
	for b.Subscribers(org) < 2 {
		time.Sleep(10 * time.Millisecond)
	}
	b.Publish(org, map[string]any{"type": "zone", "id": "PROD-A", "patch": map[string]int{"active_workers": 12}})
	for _, rd := range readers {
		line, _ := rd.ReadString('\n')
		if !strings.Contains(line, `"active_workers":12`) {
			t.Fatalf("tab did not receive update: %q", line)
		}
	}
	b.Publish(uuid.New(), map[string]string{"type": "other-tenant"}) // must not leak across orgs
}
