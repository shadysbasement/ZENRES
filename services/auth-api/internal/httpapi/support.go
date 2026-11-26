package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/shadysbasement/zevra/services/auth-api/internal/store"
	"github.com/shadysbasement/zevra/services/internal/platform"
)

type supportTicket struct {
	ID               string     `json:"id"`
	Number           int64      `json:"number"`
	Requester        string     `json:"requester,omitempty"`
	Plan             string     `json:"plan"`
	Kind             string     `json:"kind"`
	Category         string     `json:"category"`
	Priority         string     `json:"priority"`
	Status           string     `json:"status"`
	Subject          string     `json:"subject"`
	AssignedTo       string     `json:"assigned_to,omitempty"`
	CallbackAt       *time.Time `json:"callback_at,omitempty"`
	CallbackTimezone string     `json:"callback_timezone,omitempty"`
	FirstResponseAt  *time.Time `json:"first_response_at,omitempty"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Rating           *int       `json:"rating,omitempty"`
}

const ticketColumns = `t.id::text,t.ticket_no,u.email::text,t.plan_at_creation,t.kind,t.category,t.priority,t.status,t.subject,COALESCE(a.handle,''),t.callback_at,COALESCE(t.callback_timezone,''),t.first_response_at,t.resolved_at,t.created_at,t.updated_at,f.rating`
const ticketJoins = ` FROM support_tickets t JOIN users u ON u.id=t.requester_user_id LEFT JOIN admin_pgp_keys a ON a.id=t.assigned_admin_key_id LEFT JOIN support_feedback f ON f.ticket_id=t.id `

func scanTicket(row pgx.Row) (supportTicket, error) {
	var t supportTicket
	err := row.Scan(&t.ID, &t.Number, &t.Requester, &t.Plan, &t.Kind, &t.Category, &t.Priority, &t.Status, &t.Subject, &t.AssignedTo, &t.CallbackAt, &t.CallbackTimezone, &t.FirstResponseAt, &t.ResolvedAt, &t.CreatedAt, &t.UpdatedAt, &t.Rating)
	return t, err
}

func (s *Server) supportReady(w http.ResponseWriter) bool {
	if s.SupportDB == nil {
		platform.Error(w, 503, "support_unavailable")
		return false
	}
	return true
}
func (s *Server) supportUser(w http.ResponseWriter, r *http.Request) *store.User {
	u, _, err := s.currentUser(r)
	if err != nil || u.Status != "active" || !u.EmailVerified {
		platform.Error(w, 401, "unauthenticated")
		return nil
	}
	return u
}
func supportID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		platform.Error(w, 404, "not_found")
		return uuid.Nil, false
	}
	return id, true
}
func (s *Server) supportTickets(w http.ResponseWriter, r *http.Request) {
	if !s.supportReady(w) {
		return
	}
	u := s.supportUser(w, r)
	if u == nil {
		return
	}
	rows, err := s.SupportDB.Query(r.Context(), `SELECT `+ticketColumns+ticketJoins+` WHERE t.requester_user_id=$1 ORDER BY t.updated_at DESC LIMIT 100`, u.ID)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	out := []supportTicket{}
	for rows.Next() {
		t, e := scanTicket(rows)
		if e != nil {
			s.fail(w, e)
			return
		}
		t.Requester = ""
		out = append(out, t)
	}
	if rows.Err() != nil {
		s.fail(w, rows.Err())
		return
	}
	platform.JSON(w, 200, out)
}
func (s *Server) supportCreateTicket(w http.ResponseWriter, r *http.Request) {
	if !s.supportReady(w) {
		return
	}
	u := s.supportUser(w, r)
	if u == nil {
		return
	}
	if !s.Limiter.Allow("support-create:"+u.ID.String(), 10, time.Hour) {
		platform.Error(w, 429, "rate_limited")
		return
	}
	var in struct {
		Kind             string `json:"kind"`
		Category         string `json:"category"`
		Subject          string `json:"subject"`
		Message          string `json:"message"`
		CallbackAt       string `json:"callback_at"`
		CallbackTimezone string `json:"callback_timezone"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	in.Subject = strings.TrimSpace(in.Subject)
	in.Message = strings.TrimSpace(in.Message)
	if !oneOf(in.Kind, "question", "live_chat", "callback", "bug", "feedback") || !oneOf(in.Category, "platform", "tally", "cbam", "trace", "freight", "marketplace", "billing", "security", "other") || len([]rune(in.Subject)) < 5 || len([]rune(in.Subject)) > 160 || len([]rune(in.Message)) < 10 || len([]rune(in.Message)) > 4000 {
		platform.Error(w, 400, "invalid_input")
		return
	}
	var callback *time.Time
	if in.Kind == "callback" {
		when, e := time.Parse(time.RFC3339, in.CallbackAt)
		if e != nil || when.Before(s.Now().Add(2*time.Hour)) || when.After(s.Now().Add(60*24*time.Hour)) || len(in.CallbackTimezone) > 80 || strings.TrimSpace(in.CallbackTimezone) == "" {
			platform.Error(w, 400, "invalid_callback_time")
			return
		}
		callback = &when
	}
	var orgID *uuid.UUID
	plan := "unassigned"
	if org, e := s.Users.OrgForUser(r.Context(), u.ID); e == nil {
		orgID = &org.ID
		plan = org.Tier
	}
	if s.Cfg.SupportPlanEnforce && (plan == "starter" || plan == "unassigned") && oneOf(in.Kind, "live_chat", "callback") {
		platform.Error(w, 403, "plan_upgrade_required")
		return
	}
	priority := "normal"
	if in.Kind == "bug" || in.Category == "security" {
		priority = "high"
	}
	tx, err := s.SupportDB.Begin(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	var id uuid.UUID
	var number int64
	err = tx.QueryRow(r.Context(), `INSERT INTO support_tickets (requester_user_id,org_id,plan_at_creation,kind,category,priority,subject,callback_at,callback_timezone) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id,ticket_no`, u.ID, orgID, plan, in.Kind, in.Category, priority, in.Subject, callback, strings.TrimSpace(in.CallbackTimezone)).Scan(&id, &number)
	if err == nil {
		_, err = tx.Exec(r.Context(), `INSERT INTO support_messages(ticket_id,author_user_id,body) VALUES($1,$2,$3)`, id, u.ID, in.Message)
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `INSERT INTO support_audit_logs(ticket_id,actor_user_id,action) VALUES($1,$2,'ticket_created')`, id, u.ID)
	}
	if err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	platform.JSON(w, 201, map[string]any{"id": id, "number": number, "status": "open"})
}
func oneOf(value string, options ...string) bool {
	for _, o := range options {
		if value == o {
			return true
		}
	}
	return false
}

func (s *Server) customerTicket(r *http.Request, u *store.User, id uuid.UUID) (supportTicket, error) {
	t, err := scanTicket(s.SupportDB.QueryRow(r.Context(), `SELECT `+ticketColumns+ticketJoins+` WHERE t.id=$1 AND t.requester_user_id=$2`, id, u.ID))
	t.Requester = ""
	return t, err
}

type supportMessage struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	Internal  bool      `json:"internal"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Server) ticketMessages(ctx context.Context, id uuid.UUID, includeInternal bool) ([]supportMessage, error) {
	query := `SELECT m.id::text,CASE WHEN m.author_admin_key_id IS NULL THEN 'customer' ELSE 'support' END,m.body,m.internal_note,m.created_at FROM support_messages m WHERE m.ticket_id=$1`
	if !includeInternal {
		query += ` AND NOT m.internal_note`
	}
	query += ` ORDER BY m.created_at,m.id LIMIT 500`
	rows, err := s.SupportDB.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []supportMessage{}
	for rows.Next() {
		var m supportMessage
		if err := rows.Scan(&m.ID, &m.Author, &m.Body, &m.Internal, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (s *Server) supportTicket(w http.ResponseWriter, r *http.Request) {
	if !s.supportReady(w) {
		return
	}
	u := s.supportUser(w, r)
	if u == nil {
		return
	}
	id, ok := supportID(w, r)
	if !ok {
		return
	}
	t, err := s.customerTicket(r, u, id)
	if err != nil {
		platform.Error(w, 404, "not_found")
		return
	}
	m, err := s.ticketMessages(r.Context(), id, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	var rating *int
	_ = s.SupportDB.QueryRow(r.Context(), `SELECT rating FROM support_feedback WHERE ticket_id=$1 AND user_id=$2`, id, u.ID).Scan(&rating)
	w.Header().Set("Cache-Control", "no-store")
	platform.JSON(w, 200, map[string]any{"ticket": t, "messages": m, "rating": rating})
}
func (s *Server) supportCustomerMessage(w http.ResponseWriter, r *http.Request) {
	if !s.supportReady(w) {
		return
	}
	u := s.supportUser(w, r)
	if u == nil {
		return
	}
	id, ok := supportID(w, r)
	if !ok {
		return
	}
	if !s.Limiter.Allow("support-message:"+u.ID.String(), 60, time.Hour) {
		platform.Error(w, 429, "rate_limited")
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	in.Body = strings.TrimSpace(in.Body)
	if len([]rune(in.Body)) < 1 || len([]rune(in.Body)) > 4000 {
		platform.Error(w, 400, "invalid_input")
		return
	}
	tx, err := s.SupportDB.Begin(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	var status string
	err = tx.QueryRow(r.Context(), `SELECT status FROM support_tickets WHERE id=$1 AND requester_user_id=$2 FOR UPDATE`, id, u.ID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Error(w, 404, "not_found")
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	if status == "resolved" {
		platform.Error(w, 409, "ticket_resolved")
		return
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO support_messages(ticket_id,author_user_id,body) VALUES($1,$2,$3)`, id, u.ID, in.Body)
	if err == nil {
		_, err = tx.Exec(r.Context(), `UPDATE support_tickets SET status='open',updated_at=NOW() WHERE id=$1`, id)
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `INSERT INTO support_audit_logs(ticket_id,actor_user_id,action) VALUES($1,$2,'customer_reply')`, id, u.ID)
	}
	if err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]string{"status": "sent"})
}
func (s *Server) supportFeedback(w http.ResponseWriter, r *http.Request) {
	if !s.supportReady(w) {
		return
	}
	u := s.supportUser(w, r)
	if u == nil {
		return
	}
	id, ok := supportID(w, r)
	if !ok {
		return
	}
	var in struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if platform.DecodeLoose(r, &in) != nil || in.Rating < 1 || in.Rating > 5 || len([]rune(in.Comment)) > 1000 {
		platform.Error(w, 400, "invalid_input")
		return
	}
	var status string
	err := s.SupportDB.QueryRow(r.Context(), `SELECT status FROM support_tickets WHERE id=$1 AND requester_user_id=$2`, id, u.ID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Error(w, 404, "not_found")
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	if status != "resolved" {
		platform.Error(w, 409, "not_resolved")
		return
	}
	_, err = s.SupportDB.Exec(r.Context(), `INSERT INTO support_feedback(ticket_id,user_id,rating,comment) VALUES($1,$2,$3,$4) ON CONFLICT(ticket_id) DO UPDATE SET rating=$3,comment=$4,created_at=NOW() WHERE support_feedback.user_id=$2`, id, u.ID, in.Rating, strings.TrimSpace(in.Comment))
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]string{"status": "saved"})
}

type supportAdminContextKey struct{}

func (s *Server) supportAdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.supportReady(w) {
			return
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		ip := r.Header.Get("X-Client-IP")
		ua := r.Header.Get("X-Client-UA")
		if token == "" || ip == "" || ua == "" {
			notFound(w)
			return
		}
		k, err := s.Vault.ValidateSession(r.Context(), token, ip, ua)
		if err != nil || !oneOf(k.Role, "super_admin", "support_admin", "support_agent") {
			notFound(w)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), supportAdminContextKey{}, k)))
	})
}
func supportAdmin(r *http.Request) *store.AdminKey {
	return r.Context().Value(supportAdminContextKey{}).(*store.AdminKey)
}
func canManageSupport(k *store.AdminKey) bool {
	return k.Role == "super_admin" || k.Role == "support_admin"
}
func (s *Server) supportAdminTickets(w http.ResponseWriter, r *http.Request) {
	k := supportAdmin(r)
	query := `SELECT ` + ticketColumns + ticketJoins
	args := []any{}
	if !canManageSupport(k) {
		query += ` WHERE t.assigned_admin_key_id=$1`
		args = append(args, k.ID)
	}
	query += ` ORDER BY CASE t.status WHEN 'escalated' THEN 0 WHEN 'open' THEN 1 WHEN 'pending' THEN 2 ELSE 3 END,CASE t.priority WHEN 'urgent' THEN 0 WHEN 'high' THEN 1 WHEN 'normal' THEN 2 ELSE 3 END,CASE t.plan_at_creation WHEN 'importer_hub' THEN 0 WHEN 'growth' THEN 1 ELSE 2 END,t.created_at LIMIT 200`
	rows, err := s.SupportDB.Query(r.Context(), query, args...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	out := []supportTicket{}
	for rows.Next() {
		t, e := scanTicket(rows)
		if e != nil {
			s.fail(w, e)
			return
		}
		out = append(out, t)
	}
	if rows.Err() != nil {
		s.fail(w, rows.Err())
		return
	}
	platform.JSON(w, 200, out)
}

func (s *Server) supportAdminAgents(w http.ResponseWriter, r *http.Request) {
	if !canManageSupport(supportAdmin(r)) {
		platform.Error(w, 403, "forbidden")
		return
	}
	rows, err := s.SupportDB.Query(r.Context(), `SELECT id::text,handle,role FROM admin_pgp_keys WHERE status='active' AND revoked_at IS NULL AND role IN ('super_admin','support_admin','support_agent') ORDER BY handle`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	type agent struct {
		ID     string `json:"id"`
		Handle string `json:"handle"`
		Role   string `json:"role"`
	}
	out := []agent{}
	for rows.Next() {
		var a agent
		if err := rows.Scan(&a.ID, &a.Handle, &a.Role); err != nil {
			s.fail(w, err)
			return
		}
		out = append(out, a)
	}
	if rows.Err() != nil {
		s.fail(w, rows.Err())
		return
	}
	platform.JSON(w, 200, out)
}

func (s *Server) supportAdminCosts(w http.ResponseWriter, r *http.Request) {
	if !canManageSupport(supportAdmin(r)) {
		platform.Error(w, 403, "forbidden")
		return
	}
	rows, err := s.SupportDB.Query(r.Context(), `SELECT u.email::text,l.cost_currency::text,COUNT(DISTINCT t.id),SUM(l.minutes)::bigint,SUM(l.cost_amount)::text FROM support_work_logs l JOIN support_tickets t ON t.id=l.ticket_id JOIN users u ON u.id=t.requester_user_id WHERE l.cost_amount IS NOT NULL GROUP BY u.email,l.cost_currency ORDER BY u.email,l.cost_currency LIMIT 200`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	type cost struct {
		Customer string `json:"customer"`
		Currency string `json:"currency"`
		Tickets  int64  `json:"tickets"`
		Minutes  int64  `json:"minutes"`
		Amount   string `json:"amount"`
	}
	out := []cost{}
	for rows.Next() {
		var c cost
		if err := rows.Scan(&c.Customer, &c.Currency, &c.Tickets, &c.Minutes, &c.Amount); err != nil {
			s.fail(w, err)
			return
		}
		out = append(out, c)
	}
	if rows.Err() != nil {
		s.fail(w, rows.Err())
		return
	}
	platform.JSON(w, 200, out)
}
func (s *Server) adminTicket(r *http.Request, id uuid.UUID, k *store.AdminKey) (supportTicket, error) {
	query := `SELECT ` + ticketColumns + ticketJoins + ` WHERE t.id=$1`
	args := []any{id}
	if !canManageSupport(k) {
		query += ` AND t.assigned_admin_key_id=$2`
		args = append(args, k.ID)
	}
	return scanTicket(s.SupportDB.QueryRow(r.Context(), query, args...))
}
func (s *Server) supportAdminTicket(w http.ResponseWriter, r *http.Request) {
	id, ok := supportID(w, r)
	if !ok {
		return
	}
	k := supportAdmin(r)
	t, err := s.adminTicket(r, id, k)
	if err != nil {
		platform.Error(w, 404, "not_found")
		return
	}
	m, err := s.ticketMessages(r.Context(), id, true)
	if err != nil {
		s.fail(w, err)
		return
	}
	var rating *int
	_ = s.SupportDB.QueryRow(r.Context(), `SELECT rating FROM support_feedback WHERE ticket_id=$1`, id).Scan(&rating)
	platform.JSON(w, 200, map[string]any{"ticket": t, "messages": m, "rating": rating})
}
func (s *Server) supportAdminUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := supportID(w, r)
	if !ok {
		return
	}
	k := supportAdmin(r)
	var in struct {
		Status       string `json:"status"`
		Priority     string `json:"priority"`
		AssigneeID   string `json:"assignee_id"`
		Message      string `json:"message"`
		InternalNote bool   `json:"internal_note"`
		WorkMinutes  int    `json:"work_minutes"`
		WorkCost     string `json:"work_cost"`
		WorkCurrency string `json:"work_currency"`
		WorkNote     string `json:"work_note"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	in.Message = strings.TrimSpace(in.Message)
	if (in.Status != "" && !oneOf(in.Status, "open", "pending", "escalated", "resolved")) || (in.Priority != "" && !oneOf(in.Priority, "low", "normal", "high", "urgent")) || len([]rune(in.Message)) > 4000 || (in.Status == "" && in.Priority == "" && in.AssigneeID == "" && in.Message == "" && in.WorkMinutes == 0) || in.WorkMinutes < 0 || in.WorkMinutes > 480 || len([]rune(in.WorkNote)) > 500 || (in.WorkCost != "" && (in.WorkMinutes == 0 || !regexp.MustCompile(`^[0-9]{1,9}(\.[0-9]{1,2})?$`).MatchString(in.WorkCost) || !oneOf(in.WorkCurrency, "USD", "INR", "EUR"))) || (in.WorkCost == "" && in.WorkCurrency != "") {
		platform.Error(w, 400, "invalid_input")
		return
	}
	if !canManageSupport(k) && (in.Priority != "" || in.AssigneeID != "" || in.WorkCost != "") {
		platform.Error(w, 403, "forbidden")
		return
	}
	var assignee *uuid.UUID
	if in.AssigneeID != "" {
		parsed, e := uuid.Parse(in.AssigneeID)
		if e != nil {
			platform.Error(w, 400, "invalid_assignee")
			return
		}
		var role string
		var status string
		e = s.SupportDB.QueryRow(r.Context(), `SELECT role,status FROM admin_pgp_keys WHERE id=$1 AND revoked_at IS NULL`, parsed).Scan(&role, &status)
		if e != nil || status != "active" || !oneOf(role, "super_admin", "support_admin", "support_agent") {
			platform.Error(w, 400, "invalid_assignee")
			return
		}
		assignee = &parsed
	}
	tx, err := s.SupportDB.Begin(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	var currentID uuid.UUID
	if canManageSupport(k) {
		err = tx.QueryRow(r.Context(), `SELECT id FROM support_tickets WHERE id=$1 FOR UPDATE`, id).Scan(&currentID)
	} else {
		err = tx.QueryRow(r.Context(), `SELECT id FROM support_tickets WHERE id=$1 AND assigned_admin_key_id=$2 FOR UPDATE`, id, k.ID).Scan(&currentID)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Error(w, 404, "not_found")
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	_, err = tx.Exec(r.Context(), `UPDATE support_tickets SET status=COALESCE(NULLIF($2,''),status),priority=COALESCE(NULLIF($3,''),priority),assigned_admin_key_id=COALESCE($4,assigned_admin_key_id),first_response_at=CASE WHEN $5::boolean AND first_response_at IS NULL THEN NOW() ELSE first_response_at END,resolved_at=CASE WHEN $2='resolved' THEN NOW() WHEN $2<>'' THEN NULL ELSE resolved_at END,updated_at=NOW() WHERE id=$1`, id, in.Status, in.Priority, assignee, in.Message != "" && !in.InternalNote)
	if err == nil && in.Message != "" {
		_, err = tx.Exec(r.Context(), `INSERT INTO support_messages(ticket_id,author_admin_key_id,body,internal_note) VALUES($1,$2,$3,$4)`, id, k.ID, in.Message, in.InternalNote)
	}
	if err == nil && in.WorkMinutes > 0 {
		_, err = tx.Exec(r.Context(), `INSERT INTO support_work_logs(ticket_id,admin_key_id,minutes,cost_amount,cost_currency,note) VALUES($1,$2,$3,NULLIF($4,'')::numeric,NULLIF($5,''),$6)`, id, k.ID, in.WorkMinutes, in.WorkCost, in.WorkCurrency, strings.TrimSpace(in.WorkNote))
	}
	if err == nil {
		details, _ := json.Marshal(map[string]any{"status": in.Status, "priority": in.Priority, "assignee_id": in.AssigneeID, "message": in.Message != "", "internal": in.InternalNote, "work_minutes": in.WorkMinutes, "work_cost": in.WorkCost, "work_currency": in.WorkCurrency})
		_, err = tx.Exec(r.Context(), `INSERT INTO support_audit_logs(ticket_id,actor_admin_key_id,action,details) VALUES($1,$2,'admin_update',$3)`, id, k.ID, details)
	}
	if err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	if s.Cfg.SupportNotify && in.Message != "" && !in.InternalNote {
		var email string
		var number int64
		if e := s.SupportDB.QueryRow(r.Context(), `SELECT u.email::text,t.ticket_no FROM support_tickets t JOIN users u ON u.id=t.requester_user_id WHERE t.id=$1`, id).Scan(&email, &number); e == nil {
			if e := s.Mail.Send(r.Context(), email, fmt.Sprintf("ZENRES support update · ZEN-%d", number), fmt.Sprintf("Your support ticket ZEN-%d has a new reply. Sign in at %s/support to read it.\n\n— ZENRES Support", number, s.Cfg.PublicURL), nil); e != nil {
				slog.Warn("support notification failed", "err", e)
			}
		}
	}
	platform.JSON(w, 200, map[string]string{"status": "updated"})
}
func (s *Server) supportAnnouncements(w http.ResponseWriter, r *http.Request) {
	if !s.supportReady(w) {
		return
	}
	rows, err := s.SupportDB.Query(r.Context(), `SELECT id::text,title,body,state,published_at,ends_at FROM support_announcements WHERE published_at<=NOW() AND (ends_at IS NULL OR ends_at>NOW()) ORDER BY published_at DESC LIMIT 20`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	type announcement struct {
		ID          string     `json:"id"`
		Title       string     `json:"title"`
		Body        string     `json:"body"`
		State       string     `json:"state"`
		PublishedAt time.Time  `json:"published_at"`
		EndsAt      *time.Time `json:"ends_at,omitempty"`
	}
	out := []announcement{}
	for rows.Next() {
		var a announcement
		if err := rows.Scan(&a.ID, &a.Title, &a.Body, &a.State, &a.PublishedAt, &a.EndsAt); err != nil {
			s.fail(w, err)
			return
		}
		out = append(out, a)
	}
	if rows.Err() != nil {
		s.fail(w, rows.Err())
		return
	}
	platform.JSON(w, 200, out)
}
func (s *Server) supportPublishAnnouncement(w http.ResponseWriter, r *http.Request) {
	k := supportAdmin(r)
	if !canManageSupport(k) {
		platform.Error(w, 403, "forbidden")
		return
	}
	var in struct {
		Title  string `json:"title"`
		Body   string `json:"body"`
		State  string `json:"state"`
		EndsAt string `json:"ends_at"`
	}
	if platform.DecodeLoose(r, &in) != nil || len([]rune(strings.TrimSpace(in.Title))) < 5 || len([]rune(in.Title)) > 160 || len([]rune(strings.TrimSpace(in.Body))) < 5 || len([]rune(in.Body)) > 2000 || !oneOf(in.State, "maintenance", "incident", "resolved") {
		platform.Error(w, 400, "invalid_input")
		return
	}
	var ends *time.Time
	if in.EndsAt != "" {
		v, e := time.Parse(time.RFC3339, in.EndsAt)
		if e != nil || v.Before(s.Now()) {
			platform.Error(w, 400, "invalid_time")
			return
		}
		ends = &v
	}
	var id uuid.UUID
	err := s.SupportDB.QueryRow(r.Context(), `INSERT INTO support_announcements(title,body,state,ends_at,created_by) VALUES($1,$2,$3,$4,$5) RETURNING id`, strings.TrimSpace(in.Title), strings.TrimSpace(in.Body), in.State, ends, k.ID).Scan(&id)
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]any{"id": id})
}
