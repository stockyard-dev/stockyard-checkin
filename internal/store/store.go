package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Members struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	MemberId string `json:"member_id"`
	MembershipType string `json:"membership_type"`
	Status string `json:"status"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

type Checkins struct {
	ID string `json:"id"`
	MemberId string `json:"member_id"`
	MemberName string `json:"member_name"`
	CheckedInAt string `json:"checked_in_at"`
	CheckedOutAt string `json:"checked_out_at"`
	Location string `json:"location"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil { return nil, err }
	db, err := sql.Open("sqlite", filepath.Join(d, "checkin.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil { return nil, err }
	db.SetMaxOpenConns(1)
	db.Exec(`CREATE TABLE IF NOT EXISTS members(id TEXT PRIMARY KEY, name TEXT NOT NULL, email TEXT DEFAULT '', phone TEXT DEFAULT '', member_id TEXT DEFAULT '', membership_type TEXT DEFAULT '', status TEXT DEFAULT '', notes TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS checkins(id TEXT PRIMARY KEY, member_id TEXT NOT NULL, member_name TEXT DEFAULT '', checked_in_at TEXT NOT NULL, checked_out_at TEXT DEFAULT '', location TEXT DEFAULT '', notes TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) CreateMembers(e *Members) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO members(id, name, email, phone, member_id, membership_type, status, notes, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.Name, e.Email, e.Phone, e.MemberId, e.MembershipType, e.Status, e.Notes, e.CreatedAt)
	return err
}

func (d *DB) GetMembers(id string) *Members {
	var e Members
	if d.db.QueryRow(`SELECT id, name, email, phone, member_id, membership_type, status, notes, created_at FROM members WHERE id=?`, id).Scan(&e.ID, &e.Name, &e.Email, &e.Phone, &e.MemberId, &e.MembershipType, &e.Status, &e.Notes, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListMembers() []Members {
	rows, _ := d.db.Query(`SELECT id, name, email, phone, member_id, membership_type, status, notes, created_at FROM members ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Members
	for rows.Next() { var e Members; rows.Scan(&e.ID, &e.Name, &e.Email, &e.Phone, &e.MemberId, &e.MembershipType, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateMembers(e *Members) error {
	_, err := d.db.Exec(`UPDATE members SET name=?, email=?, phone=?, member_id=?, membership_type=?, status=?, notes=? WHERE id=?`, e.Name, e.Email, e.Phone, e.MemberId, e.MembershipType, e.Status, e.Notes, e.ID)
	return err
}

func (d *DB) DeleteMembers(id string) error {
	_, err := d.db.Exec(`DELETE FROM members WHERE id=?`, id)
	return err
}

func (d *DB) CountMembers() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM members`).Scan(&n); return n
}

func (d *DB) SearchMembers(q string, filters map[string]string) []Members {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (name LIKE ? OR email LIKE ? OR phone LIKE ? OR member_id LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["membership_type"]; ok && v != "" { where += " AND membership_type=?"; args = append(args, v) }
	if v, ok := filters["status"]; ok && v != "" { where += " AND status=?"; args = append(args, v) }
	rows, _ := d.db.Query(`SELECT id, name, email, phone, member_id, membership_type, status, notes, created_at FROM members WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Members
	for rows.Next() { var e Members; rows.Scan(&e.ID, &e.Name, &e.Email, &e.Phone, &e.MemberId, &e.MembershipType, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) CreateCheckins(e *Checkins) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO checkins(id, member_id, member_name, checked_in_at, checked_out_at, location, notes, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.MemberId, e.MemberName, e.CheckedInAt, e.CheckedOutAt, e.Location, e.Notes, e.CreatedAt)
	return err
}

func (d *DB) GetCheckins(id string) *Checkins {
	var e Checkins
	if d.db.QueryRow(`SELECT id, member_id, member_name, checked_in_at, checked_out_at, location, notes, created_at FROM checkins WHERE id=?`, id).Scan(&e.ID, &e.MemberId, &e.MemberName, &e.CheckedInAt, &e.CheckedOutAt, &e.Location, &e.Notes, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListCheckins() []Checkins {
	rows, _ := d.db.Query(`SELECT id, member_id, member_name, checked_in_at, checked_out_at, location, notes, created_at FROM checkins ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Checkins
	for rows.Next() { var e Checkins; rows.Scan(&e.ID, &e.MemberId, &e.MemberName, &e.CheckedInAt, &e.CheckedOutAt, &e.Location, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateCheckins(e *Checkins) error {
	_, err := d.db.Exec(`UPDATE checkins SET member_id=?, member_name=?, checked_in_at=?, checked_out_at=?, location=?, notes=? WHERE id=?`, e.MemberId, e.MemberName, e.CheckedInAt, e.CheckedOutAt, e.Location, e.Notes, e.ID)
	return err
}

func (d *DB) DeleteCheckins(id string) error {
	_, err := d.db.Exec(`DELETE FROM checkins WHERE id=?`, id)
	return err
}

func (d *DB) CountCheckins() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM checkins`).Scan(&n); return n
}

func (d *DB) SearchCheckins(q string, filters map[string]string) []Checkins {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (member_id LIKE ? OR member_name LIKE ? OR location LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	rows, _ := d.db.Query(`SELECT id, member_id, member_name, checked_in_at, checked_out_at, location, notes, created_at FROM checkins WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Checkins
	for rows.Next() { var e Checkins; rows.Scan(&e.ID, &e.MemberId, &e.MemberName, &e.CheckedInAt, &e.CheckedOutAt, &e.Location, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}
