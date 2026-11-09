package drafts

import (
	"context"
	"errors"
	"time"
)

func isoTime(value int64) string {
	if value <= 0 {
		return ""
	}
	return time.Unix(value, 0).UTC().Format(time.RFC3339)
}

// Only ownership-checked provider metadata for an existing published receipt may
// enrich its date. No state, approval, remote reference or upload lock changes.
func (s *Store) RecordPublicationTime(ctx context.Context, platform, account, reference, raw string) error {
	at, err := time.Parse(time.RFC3339, raw)
	if err != nil || at.Unix() <= 0 || at.After(time.Now().Add(time.Minute)) {
		return errors.New("invalid publication date")
	}
	if platform != "youtube" && platform != "instagram" && platform != "facebook" {
		return errors.New("unsupported publication source")
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO receipt_times(platform,account_id,video_hash,state,reference,published_at)
 SELECT platform,account_id,video_hash,state,reference,? FROM upload_attempts WHERE platform=? AND account_id=? AND reference=? AND state='published'
 ON CONFLICT(platform,account_id,video_hash) DO UPDATE SET published_at=excluded.published_at,state=excluded.state,reference=excluded.reference,
 confirmed_at=CASE WHEN receipt_times.reference=excluded.reference AND receipt_times.state=excluded.state THEN receipt_times.confirmed_at ELSE 0 END`, at.Unix(), platform, account, reference)
	return err
}
