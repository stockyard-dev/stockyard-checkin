package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stockyard-dev/stockyard-checkin/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits, dataDir: dataDir}
	s.loadPersonalConfig()
	s.mux.HandleFunc("GET /api/members", s.listMembers)
	s.mux.HandleFunc("POST /api/members", s.createMembers)
	s.mux.HandleFunc("GET /api/members/export.csv", s.exportMembers)
	s.mux.HandleFunc("GET /api/members/{id}", s.getMembers)
	s.mux.HandleFunc("PUT /api/members/{id}", s.updateMembers)
	s.mux.HandleFunc("DELETE /api/members/{id}", s.delMembers)
	s.mux.HandleFunc("GET /api/checkins", s.listCheckins)
	s.mux.HandleFunc("POST /api/checkins", s.createCheckins)
	s.mux.HandleFunc("GET /api/checkins/export.csv", s.exportCheckins)
	s.mux.HandleFunc("GET /api/checkins/{id}", s.getCheckins)
	s.mux.HandleFunc("PUT /api/checkins/{id}", s.updateCheckins)
	s.mux.HandleFunc("DELETE /api/checkins/{id}", s.delCheckins)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", s.tierHandler)
	s.mux.HandleFunc("GET /api/config", s.configHandler)
	s.mux.HandleFunc("GET /api/extras/{resource}", s.listExtras)
	s.mux.HandleFunc("GET /api/extras/{resource}/{id}", s.getExtras)
	s.mux.HandleFunc("PUT /api/extras/{resource}/{id}", s.putExtras)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(c); json.NewEncoder(w).Encode(v) }
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" { http.NotFound(w, r); return }; http.Redirect(w, r, "/ui", 302) }
func oe[T any](s []T) []T { if s == nil { return []T{} }; return s }
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if v := r.URL.Query().Get("membership_type"); v != "" { filters["membership_type"] = v }
	if v := r.URL.Query().Get("status"); v != "" { filters["status"] = v }
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"members": oe(s.db.SearchMembers(q, filters))}); return }
	wj(w, 200, map[string]any{"members": oe(s.db.ListMembers())})
}

func (s *Server) createMembers(w http.ResponseWriter, r *http.Request) {
	if s.limits.Tier == "none" { we(w, 402, "No license key. Start a 14-day trial at https://stockyard.dev/for/"); return }
	if s.limits.TrialExpired { we(w, 402, "Trial expired. Subscribe at https://stockyard.dev/pricing/"); return }
	var e store.Members
	json.NewDecoder(r.Body).Decode(&e)
	if e.Name == "" { we(w, 400, "name required"); return }
	s.db.CreateMembers(&e)
	wj(w, 201, s.db.GetMembers(e.ID))
}

func (s *Server) getMembers(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetMembers(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateMembers(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetMembers(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Members
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.Name == "" { patch.Name = existing.Name }
	if patch.Email == "" { patch.Email = existing.Email }
	if patch.Phone == "" { patch.Phone = existing.Phone }
	if patch.MemberId == "" { patch.MemberId = existing.MemberId }
	if patch.MembershipType == "" { patch.MembershipType = existing.MembershipType }
	if patch.Status == "" { patch.Status = existing.Status }
	if patch.Notes == "" { patch.Notes = existing.Notes }
	s.db.UpdateMembers(&patch)
	wj(w, 200, s.db.GetMembers(patch.ID))
}

func (s *Server) delMembers(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id"); s.db.DeleteMembers(id); s.db.DeleteExtras("members", id)
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportMembers(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListMembers()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=members.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "name", "email", "phone", "member_id", "membership_type", "status", "notes", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.Name), fmt.Sprintf("%v", e.Email), fmt.Sprintf("%v", e.Phone), fmt.Sprintf("%v", e.MemberId), fmt.Sprintf("%v", e.MembershipType), fmt.Sprintf("%v", e.Status), fmt.Sprintf("%v", e.Notes), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) listCheckins(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"checkins": oe(s.db.SearchCheckins(q, filters))}); return }
	wj(w, 200, map[string]any{"checkins": oe(s.db.ListCheckins())})
}

func (s *Server) createCheckins(w http.ResponseWriter, r *http.Request) {
	var e store.Checkins
	json.NewDecoder(r.Body).Decode(&e)
	if e.MemberId == "" { we(w, 400, "member_id required"); return }
	if e.CheckedInAt == "" { we(w, 400, "checked_in_at required"); return }
	s.db.CreateCheckins(&e)
	wj(w, 201, s.db.GetCheckins(e.ID))
}

func (s *Server) getCheckins(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetCheckins(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateCheckins(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetCheckins(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Checkins
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.MemberId == "" { patch.MemberId = existing.MemberId }
	if patch.MemberName == "" { patch.MemberName = existing.MemberName }
	if patch.CheckedInAt == "" { patch.CheckedInAt = existing.CheckedInAt }
	if patch.CheckedOutAt == "" { patch.CheckedOutAt = existing.CheckedOutAt }
	if patch.Location == "" { patch.Location = existing.Location }
	if patch.Notes == "" { patch.Notes = existing.Notes }
	s.db.UpdateCheckins(&patch)
	wj(w, 200, s.db.GetCheckins(patch.ID))
}

func (s *Server) delCheckins(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id"); s.db.DeleteCheckins(id); s.db.DeleteExtras("checkins", id)
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportCheckins(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListCheckins()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=checkins.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "member_id", "member_name", "checked_in_at", "checked_out_at", "location", "notes", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.MemberId), fmt.Sprintf("%v", e.MemberName), fmt.Sprintf("%v", e.CheckedInAt), fmt.Sprintf("%v", e.CheckedOutAt), fmt.Sprintf("%v", e.Location), fmt.Sprintf("%v", e.Notes), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{}
	m["members_total"] = s.db.CountMembers()
	m["checkins_total"] = s.db.CountCheckins()
	wj(w, 200, m)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{"status": "ok", "service": "checkin"}
	m["members"] = s.db.CountMembers()
	m["checkins"] = s.db.CountCheckins()
	wj(w, 200, m)
}

// loadPersonalConfig reads config.json from the data directory.
func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("warning: could not parse config.json: %v", err)
		return
	}
	s.pCfg = cfg
	log.Printf("loaded personalization from %s", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		wj(w, 200, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}

// listExtras returns all extras for a resource type as {record_id: {...fields...}}
func (s *Server) listExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	all := s.db.AllExtras(resource)
	out := make(map[string]json.RawMessage, len(all))
	for id, data := range all {
		out[id] = json.RawMessage(data)
	}
	wj(w, 200, out)
}

// getExtras returns the extras blob for a single record.
func (s *Server) getExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	data := s.db.GetExtras(resource, id)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

// putExtras stores the extras blob for a single record.
func (s *Server) putExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		we(w, 400, "read body")
		return
	}
	var probe map[string]any
	if err := json.Unmarshal(body, &probe); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if err := s.db.SetExtras(resource, id, string(body)); err != nil {
		we(w, 500, "save failed")
		return
	}
	wj(w, 200, map[string]string{"ok": "saved"})
}
