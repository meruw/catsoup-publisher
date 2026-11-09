package drafts

import (
	"database/sql"
	"encoding/hex"
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
	if err := CheckDestination(d.Key, account); err != nil {
		return err
	}
	res, err := s.db.Exec(`INSERT OR IGNORE INTO upload_attempts(platform,account_id,video_hash,state,reference,started)
 SELECT platform,?,video_hash,'sending','',? FROM drafts
 WHERE episode=? AND language=? AND platform=? AND account=? AND revision=? AND video_hash=? AND EXISTS(SELECT 1 FROM accounts a WHERE a.platform=drafts.platform AND a.language=drafts.language AND a.account_id=?)
 AND `+approvalCondition+``, account, time.Now().Unix(), d.Key.Episode, d.Key.Language, d.Key.Platform, d.Key.Account, d.Revision, d.VideoHash, account)
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

// Associate the provider container with both the approved input and exact
// prepared transport bytes. Neither hash substitutes for the post's approval.
func (s *Store) SaveInstagramCheckpoint(d Draft, account, reference, uploadHash string) error {
	if d.Key.Platform != "instagram" || reference == "" || len(uploadHash) != 64 {
		return errors.New("invalid Instagram checkpoint")
	}
	if _, err := hex.DecodeString(uploadHash); err != nil {
		return errors.New("invalid upload hash")
	}
	res, err := s.db.Exec(`UPDATE upload_attempts SET reference=?,upload_hash=?,preparation='instagram-aac128-v1' WHERE platform='instagram' AND account_id=? AND video_hash=? AND state='sending'`, reference, uploadHash, account, d.VideoHash)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("upload reservation changed")
	}
	return nil
}

// The service must confirm the exact remote container failed before calling.
// Keep its reference until the replacement checkpoint is durably saved.
func (s *Store) ReserveInstagramRetry(d Draft, account, reference string) error {
	return s.reserveInstagram(d, account, reference, false)
}

func (s *Store) ReserveInstagramResume(d Draft, account, reference string) error {
	return s.reserveInstagram(d, account, reference, true)
}

// Keep the original reference and require byte-identical preparation before a
// manual retransmission. The update also prevents concurrent recovery attempts.
func (s *Store) ReserveInstagramRetransfer(d Draft, account, reference, uploadHash string) error {
	if d.Key.Platform != "instagram" || reference == "" || len(uploadHash) != 64 {
		return errors.New("invalid Instagram recovery")
	}
	if err := CheckDestination(d.Key, account); err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE upload_attempts SET state='sending',started=?
 WHERE platform='instagram' AND account_id=? AND video_hash=? AND state IN ('failed','check') AND reference=? AND upload_hash=? AND preparation='instagram-aac128-v1'
 AND EXISTS(SELECT 1 FROM drafts WHERE episode=? AND language=? AND platform='instagram' AND account=? AND revision=? AND video_hash=? AND `+approvalCondition+`)
 AND EXISTS(SELECT 1 FROM accounts WHERE platform='instagram' AND language=? AND account_id=?)`, time.Now().Unix(), account, d.VideoHash, reference, uploadHash, d.Key.Episode, d.Key.Language, d.Key.Account, d.Revision, d.VideoHash, d.Key.Language, account)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("Instagram recovery requires the original prepared video, current approval and unchanged receipt")
	}
	return nil
}

func (s *Store) reserveInstagram(d Draft, account, reference string, resume bool) error {
	if d.Key.Platform != "instagram" || reference == "" {
		return errors.New("select a failed Instagram upload")
	}
	if err := CheckDestination(d.Key, account); err != nil {
		return err
	}
	state := "sending"
	if resume {
		state = "publishing"
	}
	res, err := s.db.Exec(`UPDATE upload_attempts SET state=?,started=?
 WHERE platform='instagram' AND account_id=? AND video_hash=? AND state IN ('failed','check') AND reference=?
 AND EXISTS(SELECT 1 FROM drafts WHERE episode=? AND language=? AND platform='instagram' AND account=? AND revision=? AND video_hash=? AND `+approvalCondition+`)
 AND EXISTS(SELECT 1 FROM accounts WHERE platform='instagram' AND language=? AND account_id=?)`, state, time.Now().Unix(), account, d.VideoHash, reference, d.Key.Episode, d.Key.Language, d.Key.Account, d.Revision, d.VideoHash, d.Key.Language, account)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("failed upload, approval or account changed; refresh its receipt")
	}
	return nil
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
	if err := CheckDestination(d.Key, account); err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE upload_attempts SET state='publishing',started=?
 WHERE platform='youtube' AND account_id=? AND video_hash=? AND reference=? AND (state IN ('private','sent','check_publish') OR (state='publishing' AND started<?))
 AND EXISTS(SELECT 1 FROM drafts WHERE episode=? AND language=? AND platform='youtube' AND account=? AND revision=? AND video_hash=?
 AND `+approvalCondition+`)
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
