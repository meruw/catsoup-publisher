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
	if r.State == "publishing" && time.Since(time.Unix(started, 0)) > 30*time.Minute {
		r.State = "check_publish"
	}
	return r, err
}
func (s *Store) Reserve(d Draft, account string) error {
	res, err := s.db.Exec(`INSERT OR IGNORE INTO upload_attempts
 SELECT platform,?,video_hash,'sending','',? FROM drafts
 WHERE episode=? AND language=? AND platform=? AND account=? AND revision=? AND video_hash=? AND EXISTS(SELECT 1 FROM accounts a WHERE a.platform=drafts.platform AND a.language=drafts.language AND a.account_id=?)
 AND EXISTS(SELECT 1 FROM reviews r WHERE r.episode=drafts.episode AND r.language=drafts.language AND r.video_id=drafts.video_id AND r.video_hash=drafts.video_hash AND r.fingerprint=drafts.delivery_fingerprint AND r.state='reviewed')`, account, time.Now().Unix(), d.Key.Episode, d.Key.Language, d.Key.Platform, d.Key.Account, d.Revision, d.VideoHash, account)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("draft or destination changed, or an upload already exists; check its receipt before retrying")
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

// Promotion has its own lock and recovery state. It must never unlock a re-upload.
func (s *Store) ReservePromotion(d Draft, account, reference string) error {
	res, err := s.db.Exec(`UPDATE upload_attempts SET state='publishing',started=?
 WHERE platform='youtube' AND account_id=? AND video_hash=? AND reference=? AND (state IN ('private','sent','check_publish') OR (state='publishing' AND started<?))
 AND EXISTS(SELECT 1 FROM drafts WHERE episode=? AND language=? AND platform='youtube' AND account=? AND revision=? AND video_hash=?
 AND EXISTS(SELECT 1 FROM reviews r WHERE r.episode=drafts.episode AND r.language=drafts.language AND r.video_id=drafts.video_id AND r.video_hash=drafts.video_hash AND r.fingerprint=drafts.delivery_fingerprint AND r.state='reviewed'))
 AND EXISTS(SELECT 1 FROM accounts WHERE platform='youtube' AND language=? AND account_id=?)`, time.Now().Unix(), account, d.VideoHash, reference, time.Now().Add(-30*time.Minute).Unix(), d.Key.Episode, d.Key.Language, d.Key.Account, d.Revision, d.VideoHash, d.Key.Language, account)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("private upload or draft changed; refresh its receipt")
	}
	return nil
}
