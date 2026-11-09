package drafts

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"catsoup/publisher/internal/statistics"
)

// RecordStatistics stores observations, never inferred daily activity. One latest
// sample per five-minute bucket, at most 2,000 per remote post, retained for a year.
// The provider account and remote post ID isolate reposts and language accounts.
func (s *Store) RecordStatistics(ctx context.Context, platform, account, reference string, sample statistics.Sample) ([]statistics.Sample, error) {
	out := []statistics.Sample{}
	if (platform != "youtube" && platform != "instagram" && platform != "facebook") || !validText(account, 100) || account == "" || !validText(reference, 160) || reference == "" {
		return out, errors.New("invalid statistics destination")
	}
	at, err := time.Parse(time.RFC3339, sample.UpdatedAt)
	if err != nil || at.After(time.Now().Add(time.Minute)) || len(sample.Items) == 0 || len(sample.Items) > 3 {
		return out, errors.New("invalid statistics sample")
	}
	seen := map[string]bool{}
	for _, m := range sample.Items {
		if (m.Label != "Views" && m.Label != "Likes" && m.Label != "Comments") || seen[m.Label] {
			return out, errors.New("invalid statistics metric")
		}
		if _, err := strconv.ParseUint(m.Value, 10, 64); err != nil {
			return out, errors.New("invalid statistics count")
		}
		seen[m.Label] = true
	}
	encoded, err := json.Marshal(sample.Items)
	if err != nil {
		return out, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO statistics_samples VALUES(?,?,?,?,?,?)
 ON CONFLICT(platform,account,reference,bucket) DO UPDATE SET observed=excluded.observed,items=excluded.items
 WHERE excluded.observed>=statistics_samples.observed`, platform, account, reference, at.Unix()/300, at.Unix(), string(encoded)); err != nil {
		return out, err
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM statistics_samples WHERE observed<?", time.Now().AddDate(-1, 0, 0).Unix()); err != nil {
		return out, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM statistics_samples WHERE platform=? AND account=? AND reference=? AND bucket NOT IN
 (SELECT bucket FROM statistics_samples WHERE platform=? AND account=? AND reference=? ORDER BY bucket DESC LIMIT 2000)`, platform, account, reference, platform, account, reference); err != nil {
		return out, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT observed,items FROM statistics_samples WHERE platform=? AND account=? AND reference=? ORDER BY observed`, platform, account, reference)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var stamp int64
		var raw string
		var point statistics.Sample
		if err = rows.Scan(&stamp, &raw); err != nil {
			rows.Close()
			return out, err
		}
		if err = json.Unmarshal([]byte(raw), &point.Items); err != nil {
			rows.Close()
			return out, err
		}
		point.UpdatedAt = time.Unix(stamp, 0).UTC().Format(time.RFC3339)
		out = append(out, point)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return out, err
	}
	return out, tx.Commit()
}
