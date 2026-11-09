// Package drafts owns local metadata; saving a draft never changes a delivery.
package drafts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	_ "modernc.org/sqlite"
)

var ErrConflict = errors.New("draft changed in another window; reload the saved draft before editing")
var identifier = regexp.MustCompile(`^catsoup(?:-[a-z0-9]+)+$`)

// Bound drafts retain the provider's identity when a connected account changes.
// Unbound legacy text remains readable for explicit recovery in the editor.
func CheckDestination(k Key, id string) error {
	if strings.HasPrefix(k.Account, "account:") && k.Account != "account:"+id {
		return errors.New("this draft belongs to a different account")
	}
	if k.Account == "unassigned" {
		return errors.New("connect an account and save its draft first")
	}
	return nil
}

type Key struct {
	Episode  string `json:"episode"`
	Language string `json:"language"`
	Platform string `json:"platform"`
	Account  string `json:"account"`
}

type Draft struct {
	Key                 Key    `json:"key"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	Hashtags            string `json:"hashtags"`
	DeliveryMode        string `json:"deliveryMode"`
	ApprovedRevision    int64  `json:"approvedRevision"`
	VideoID             string `json:"videoId"`
	VideoHash           string `json:"videoHash"`
	DeliveryFingerprint string `json:"deliveryFingerprint"`
	Revision            int64  `json:"revision"`
	UpdatedAt           string `json:"updatedAt"`
}

func ValidateKey(k Key) error {
	if !identifier.MatchString(k.Episode) || len(k.Episode) > 160 {
		return errors.New("invalid episode identifier")
	}
	if k.Language != "es" && k.Language != "en" {
		return errors.New("choose Spanish or English")
	}
	switch k.Platform {
	case "youtube", "tiktok", "instagram", "facebook", "x":
	default:
		return errors.New("unsupported platform")
	}
	if strings.TrimSpace(k.Account) != k.Account || k.Account == "" || !validText(k.Account, 100) {
		return errors.New("account label must contain 1–100 characters without surrounding spaces")
	}
	return nil
}
func validText(s string, limit int) bool {
	return utf8.ValidString(s) && utf8.RuneCountInString(s) <= limit && !strings.ContainsRune(s, 0)
}
func Validate(d Draft) error {
	if err := ValidateKey(d.Key); err != nil {
		return err
	}
	if d.DeliveryMode != "" && d.DeliveryMode != "draft" && d.DeliveryMode != "publish" {
		return errors.New("invalid delivery mode")
	}
	if d.Revision < 0 || !validText(d.Title, 500) || !validText(d.Description, 10000) || !validText(d.Hashtags, 2000) {
		return errors.New("draft exceeds the local text limits or contains invalid text")
	}
	if d.VideoID == "" || !validText(d.VideoID, 160) || !hash(d.VideoHash) || !hash(d.DeliveryFingerprint) {
		return errors.New("draft must reference a validated video")
	}
	return nil
}
func hash(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

type Store struct{ db *sql.DB }

// Open migrates atomically, rejects newer schemas, and never resets a damaged DB.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	p := filepath.ToSlash(absolute)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	u := url.URL{Scheme: "file", Path: p}
	q := url.Values{}
	q.Add("_pragma", "busy_timeout(3000)")
	q.Add("_pragma", "synchronous(FULL)")
	q.Add("_pragma", "foreign_keys(ON)")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err = s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("open draft database (no files reset): %w", err)
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Acquire the write lock before reading the version to serialize first-run setup.
	if _, err = tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_version (id INTEGER PRIMARY KEY CHECK(id=1), version INTEGER NOT NULL)"); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO schema_version VALUES(1,0)"); err != nil {
		return err
	}
	var version int
	if err = tx.QueryRowContext(ctx, "SELECT version FROM schema_version WHERE id=1").Scan(&version); err != nil {
		return err
	}
	if version > 9 || version < 0 {
		return errors.New("unsupported schema version; use a compatible Publisher build")
	}
	if version == 0 {
		_, err = tx.ExecContext(ctx, `CREATE TABLE drafts (
   episode TEXT NOT NULL, language TEXT NOT NULL, platform TEXT NOT NULL, account TEXT NOT NULL,
   title TEXT NOT NULL, description TEXT NOT NULL, hashtags TEXT NOT NULL,
   video_id TEXT NOT NULL, video_hash TEXT NOT NULL, delivery_fingerprint TEXT NOT NULL,
   revision INTEGER NOT NULL CHECK(revision>0), updated_at TEXT NOT NULL,
   PRIMARY KEY(episode,language,platform,account));
   UPDATE schema_version SET version=1 WHERE id=1;`)
		if err != nil {
			return err
		}
	}
	if version < 2 {
		_, err = tx.ExecContext(ctx, `CREATE TABLE upload_attempts (
 platform TEXT NOT NULL, account_id TEXT NOT NULL, video_hash TEXT NOT NULL,
 state TEXT NOT NULL, reference TEXT NOT NULL, started INTEGER NOT NULL,
 PRIMARY KEY(platform,account_id,video_hash));
 UPDATE schema_version SET version=2 WHERE id=1;`)
		if err != nil {
			return err
		}
	}
	if version < 3 {
		_, err = tx.ExecContext(ctx, `CREATE TABLE accounts(platform TEXT NOT NULL, language TEXT NOT NULL,
 account_id TEXT NOT NULL, name TEXT NOT NULL, secret BLOB NOT NULL,
 PRIMARY KEY(platform,language), UNIQUE(platform,account_id));
 UPDATE schema_version SET version=3 WHERE id=1`)
		if err != nil {
			return err
		}
	}
	if version < 4 {
		_, err = tx.ExecContext(ctx, `CREATE TABLE reviews (
 episode TEXT NOT NULL, language TEXT NOT NULL, video_id TEXT NOT NULL,
 fingerprint TEXT NOT NULL, video_hash TEXT NOT NULL, state TEXT NOT NULL,
 notes TEXT NOT NULL, revision INTEGER NOT NULL, updated_at TEXT NOT NULL,
 PRIMARY KEY(episode,language,video_id));
 UPDATE schema_version SET version=4 WHERE id=1;`)
		if err != nil {
			return err
		}
	}
	if version < 5 {
		if _, err = tx.ExecContext(ctx, `ALTER TABLE drafts ADD COLUMN delivery_mode TEXT NOT NULL DEFAULT 'draft'; UPDATE schema_version SET version=5 WHERE id=1;`); err != nil {
			return err
		}
	}
	if version < 6 {
		if _, err = tx.ExecContext(ctx, `ALTER TABLE drafts ADD COLUMN approved_revision INTEGER NOT NULL DEFAULT 0; UPDATE schema_version SET version=6 WHERE id=1;`); err != nil {
			return err
		}
	}
	if version < 7 {
		if _, err = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS publishing_defaults(platform TEXT NOT NULL,language TEXT NOT NULL,account_id TEXT NOT NULL,mode TEXT NOT NULL CHECK(mode IN ('draft','publish')),PRIMARY KEY(platform,language,account_id)); UPDATE schema_version SET version=7 WHERE id=1;`); err != nil {
			return err
		}
	}
	if version < 8 {
		if _, err = tx.ExecContext(ctx, `CREATE TABLE statistics_samples (
 platform TEXT NOT NULL, account TEXT NOT NULL, reference TEXT NOT NULL,
 bucket INTEGER NOT NULL, observed INTEGER NOT NULL, items TEXT NOT NULL,
 PRIMARY KEY(platform,account,reference,bucket));
 CREATE INDEX statistics_expiry ON statistics_samples(observed);
 UPDATE schema_version SET version=8 WHERE id=1;`); err != nil {
			return err
		}
	}
	if version < 9 {
		if _, err = tx.ExecContext(ctx, `ALTER TABLE upload_attempts ADD COLUMN upload_hash TEXT NOT NULL DEFAULT ''; ALTER TABLE upload_attempts ADD COLUMN preparation TEXT NOT NULL DEFAULT ''; UPDATE schema_version SET version=9 WHERE id=1;`); err != nil {
			return err
		}
	}
	return tx.Commit()
}

const columns = "episode,language,platform,account,title,description,hashtags,video_id,video_hash,delivery_fingerprint,revision,updated_at,delivery_mode,approved_revision"

func scan(row interface{ Scan(...any) error }) (d Draft, err error) {
	err = row.Scan(&d.Key.Episode, &d.Key.Language, &d.Key.Platform, &d.Key.Account, &d.Title, &d.Description, &d.Hashtags, &d.VideoID, &d.VideoHash, &d.DeliveryFingerprint, &d.Revision, &d.UpdatedAt, &d.DeliveryMode, &d.ApprovedRevision)
	return
}
func (s *Store) Get(ctx context.Context, k Key) (Draft, error) {
	if err := ValidateKey(k); err != nil {
		return Draft{}, err
	}
	d, err := scan(s.db.QueryRowContext(ctx, "SELECT "+columns+" FROM drafts WHERE episode=? AND language=? AND platform=? AND account=?", k.Episode, k.Language, k.Platform, k.Account))
	if errors.Is(err, sql.ErrNoRows) {
		mode, modeErr := s.DefaultMode(ctx, k)
		return Draft{Key: k, DeliveryMode: mode}, modeErr
	}
	return d, err
}
func (s *Store) List(ctx context.Context, episode, language string) ([]Draft, error) {
	if err := ValidateKey(Key{episode, language, "youtube", "check"}); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, "SELECT "+columns+" FROM drafts WHERE episode=? AND language=? ORDER BY updated_at DESC,platform,account", episode, language)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Draft{}
	for rows.Next() {
		d, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}
func (s *Store) Save(ctx context.Context, d Draft) (Draft, error) {
	if d.DeliveryMode == "" {
		d.DeliveryMode = "draft"
	}
	if err := Validate(d); err != nil {
		return Draft{}, err
	}
	previous := d.Revision
	// Text, mode or file changes revoke this account's approval.
	d.ApprovedRevision = 0
	d.Revision++
	d.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	var result sql.Result
	var err error
	if previous == 0 {
		result, err = s.db.ExecContext(ctx, "INSERT INTO drafts ("+columns+") VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING", d.Key.Episode, d.Key.Language, d.Key.Platform, d.Key.Account, d.Title, d.Description, d.Hashtags, d.VideoID, d.VideoHash, d.DeliveryFingerprint, d.Revision, d.UpdatedAt, d.DeliveryMode, d.ApprovedRevision)
	} else {
		result, err = s.db.ExecContext(ctx, `UPDATE drafts SET title=?,description=?,hashtags=?,video_id=?,video_hash=?,delivery_fingerprint=?,revision=?,updated_at=?,delivery_mode=?,approved_revision=0 WHERE episode=? AND language=? AND platform=? AND account=? AND revision=?`, d.Title, d.Description, d.Hashtags, d.VideoID, d.VideoHash, d.DeliveryFingerprint, d.Revision, d.UpdatedAt, d.DeliveryMode, d.Key.Episode, d.Key.Language, d.Key.Platform, d.Key.Account, previous)
	}
	if err != nil {
		return Draft{}, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return Draft{}, err
	}
	if count != 1 {
		return Draft{}, ErrConflict
	}
	return d, nil
}
