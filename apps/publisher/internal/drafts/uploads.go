package drafts

import (
	"database/sql"
	"errors"
	"time"
)

type Receipt struct {
	State        string `json:"state"`
	Reference    string `json:"reference"`
	RetryAllowed bool   `json:"retryAllowed"`
}

func (s *Store) Receipt(platform, account, hash string) (Receipt, error) {
	var r Receipt
	var started int64
	err := s.db.QueryRow("SELECT state,reference,started FROM upload_attempts WHERE platform=? AND account_id=? AND video_hash=?", platform, account, hash).Scan(&r.State, &r.Reference, &started)
	if errors.Is(err, sql.ErrNoRows) {
		return Receipt{}, nil
	}
	r.RetryAllowed = r.State == "check" || (r.State == "sending" && time.Since(time.Unix(started, 0)) > 30*time.Minute)
	return r, err
}
func (s *Store) Reserve(d Draft, account string) error {
	res, err := s.db.Exec(`INSERT OR IGNORE INTO upload_attempts
 SELECT platform,?,video_hash,'sending','',? FROM drafts
 WHERE episode=? AND language=? AND platform=? AND account=? AND revision=? AND video_hash=?`, account, time.Now().Unix(), d.Key.Episode, d.Key.Language, d.Key.Platform, d.Key.Account, d.Revision, d.VideoHash)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("draft changed or an upload already exists; check its receipt before retrying")
	}
	return nil
}
func (s *Store) SaveReceipt(d Draft, account, state, reference string) error {
	_, err := s.db.Exec("UPDATE upload_attempts SET state=?,reference=CASE WHEN ?='' THEN reference ELSE ? END WHERE platform=? AND account_id=? AND video_hash=?", state, reference, reference, d.Key.Platform, account, d.VideoHash)
	return err
}
func (s *Store) ResetAttempt(d Draft, account string) error {
	res, err := s.db.Exec("DELETE FROM upload_attempts WHERE platform=? AND account_id=? AND video_hash=? AND (state='check' OR (state='sending' AND started<?))", d.Key.Platform, account, d.VideoHash, time.Now().Add(-30*time.Minute).Unix())
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("only an unresolved upload may be cleared after checking the platform")
	}
	return nil
}
