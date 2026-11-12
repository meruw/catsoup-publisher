package drafts

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type TikTokTracking struct {
	Status    string `json:"status"`
	Reason    string `json:"reason"`
	CheckedAt int64  `json:"checkedAt"`
	NextCheck int64  `json:"nextCheck"`
	Problem   string `json:"problem"`
}
type TikTokWatch struct {
	Account, Language, Reference string
	Started                      int64
	Failures                     int
}

// Reconstruct work from durable active AND archived receipts. Account joins
// prevent using a replacement account's token for an older account's receipts.
func (s *Store) DueTikTok(ctx context.Context, now int64) ([]TikTokWatch, error) {
	rows, err := s.db.QueryContext(ctx, `WITH receipts AS (
 SELECT account_id,reference,started FROM upload_attempts WHERE platform='tiktok'
 AND (state<>'sending' OR started<?)
 UNION ALL SELECT account_id,reference,started FROM inbox_receipt_history)
 SELECT r.account_id,a.language,r.reference,MIN(r.started),COALESCE(t.failures,0)
 FROM receipts r JOIN accounts a ON a.platform='tiktok' AND a.account_id=r.account_id
 LEFT JOIN tiktok_tracking t ON t.account_id=r.account_id AND t.reference=r.reference
 WHERE r.reference GLOB 'v_inbox_file~v2.[0-9]*'
 AND COALESCE(t.next_check,0)<=? AND COALESCE(t.status,'') NOT IN ('PUBLISH_COMPLETE','FAILED')
 GROUP BY r.account_id,a.language,r.reference ORDER BY COALESCE(t.next_check,0),MIN(r.started) LIMIT 8`, now-16*60, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []TikTokWatch
	for rows.Next() {
		var w TikTokWatch
		if err = rows.Scan(&w.Account, &w.Language, &w.Reference, &w.Started, &w.Failures); err != nil {
			return nil, err
		}
		list = append(list, w)
	}
	return list, rows.Err()
}
func (s *Store) TikTokTracking(account, reference string) (*TikTokTracking, error) {
	var t TikTokTracking
	err := s.db.QueryRow(`SELECT status,reason,checked_at,next_check,problem FROM tiktok_tracking WHERE account_id=? AND reference=?`, account, reference).Scan(&t.Status, &t.Reason, &t.CheckedAt, &t.NextCheck, &t.Problem)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &t, err
}
func (s *Store) SaveTikTokObservation(ctx context.Context, w TikTokWatch, status, reason, problem string, now int64) error {
	if problem == "" {
		switch status {
		case "PROCESSING_UPLOAD", "PROCESSING_DOWNLOAD", "SEND_TO_USER_INBOX", "PUBLISH_COMPLETE", "FAILED":
		default:
			return errors.New("invalid TikTok status")
		}
	}
	interval := int64(60)
	if status == "SEND_TO_USER_INBOX" {
		interval = 15 * 60
		if now-w.Started > 24*3600 {
			interval = 6 * 3600
		}
	}
	if status == "PROCESSING_UPLOAD" || status == "PROCESSING_DOWNLOAD" {
		if now-w.Started > 3600 {
			interval = 15 * 60
		}
	}
	failures := 0
	if problem != "" {
		failures = w.Failures + 1
		interval = int64(60) * int64(1<<min(failures, 8))
		problem = "Could not check TikTok. Retrying automatically; last confirmed status retained."
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO tiktok_tracking(account_id,reference,status,reason,checked_at,next_check,failures,problem)
 SELECT ?,?,?,?,?,?,?,? WHERE EXISTS(SELECT 1 FROM accounts WHERE platform='tiktok' AND language=? AND account_id=?)
 ON CONFLICT(account_id,reference) DO UPDATE SET
 status=CASE WHEN excluded.problem='' THEN excluded.status ELSE tiktok_tracking.status END,
 reason=CASE WHEN excluded.problem='' THEN excluded.reason ELSE tiktok_tracking.reason END,
 checked_at=excluded.checked_at,next_check=excluded.next_check,failures=excluded.failures,problem=excluded.problem`, w.Account, w.Reference, status, reason, now, now+interval, failures, problem, w.Language, w.Account)
	if err != nil {
		return err
	}
	// Reconcile only the same receipt. Archived results must never overwrite a
	// newer resend, and a recent in-flight sender owns its own completion write.
	state := ""
	switch status {
	case "SEND_TO_USER_INBOX":
		state = "inbox"
	case "PUBLISH_COMPLETE":
		state = "published"
	case "FAILED":
		state = "failed"
	}
	if problem == "" && state != "" {
		_, err = tx.ExecContext(ctx, `UPDATE upload_attempts SET state=? WHERE platform='tiktok' AND account_id=? AND reference=?
 AND (state IN ('check','inbox') OR (state='sending' AND started<?))
 AND EXISTS(SELECT 1 FROM accounts WHERE platform='tiktok' AND language=? AND account_id=?)`, state, w.Account, w.Reference, now-16*60, w.Language, w.Account)
		if err != nil {
			return err
		}
		if state == "inbox" || state == "published" {
			_, err = tx.ExecContext(ctx, `INSERT INTO receipt_times(platform,account_id,video_hash,state,reference,confirmed_at)
 SELECT platform,account_id,video_hash,state,reference,? FROM upload_attempts
 WHERE platform='tiktok' AND account_id=? AND reference=? AND state=?
 ON CONFLICT(platform,account_id,video_hash) DO UPDATE SET state=excluded.state,reference=excluded.reference,confirmed_at=excluded.confirmed_at,published_at=0
 WHERE receipt_times.state<>excluded.state OR receipt_times.reference<>excluded.reference`, now, w.Account, w.Reference, state)
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

// StartTikTokMonitor lives only while Publisher is open. The persisted schedule
// is resumed on startup; there is no background service or upload retry.
func StartTikTokMonitor(ctx context.Context, s *Store, fetch func(context.Context, TikTokWatch) (string, string, error), changed func()) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if ctx.Err() != nil {
				return
			}
			list, err := s.DueTikTok(ctx, time.Now().Unix())
			if err == nil {
				for _, w := range list {
					call, cancel := context.WithTimeout(ctx, 25*time.Second)
					status, reason, err := fetch(call, w)
					cancel()
					if ctx.Err() != nil {
						return
					}
					problem := ""
					if err != nil {
						problem = "unavailable"
					}
					if s.SaveTikTokObservation(ctx, w, status, reason, problem, time.Now().Unix()) == nil && changed != nil {
						changed()
					}
					// Bound all-account polling below TikTok's per-token status request limit.
					select {
					case <-ctx.Done():
						return
					case <-time.After(5 * time.Second):
					}
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(30 * time.Second):
			}
		}
	}()
	return done
}
