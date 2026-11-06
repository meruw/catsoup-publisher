package drafts

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Review is an operator's local assessment. It never grants delivery approval.
type Review struct {
	Episode     string `json:"episode"`
	Language    string `json:"language"`
	VideoID     string `json:"videoId"`
	Fingerprint string `json:"fingerprint"`
	VideoHash   string `json:"videoHash"`
	State       string `json:"state"`
	Notes       string `json:"notes"`
	Revision    int64  `json:"revision"`
	UpdatedAt   string `json:"updatedAt"`
}

func (s *Store) GetReview(ctx context.Context, episode, language, video string) (Review, error) {
	r := Review{Episode: episode, Language: language, VideoID: video, State: "pending"}
	if err := ValidateKey(Key{episode, language, "youtube", "review"}); err != nil {
		return r, err
	}
	if video == "" || !validText(video, 160) {
		return r, errors.New("invalid video identifier")
	}
	err := s.db.QueryRowContext(ctx, `SELECT fingerprint,video_hash,state,notes,revision,updated_at FROM reviews WHERE episode=? AND language=? AND video_id=?`, episode, language, video).Scan(&r.Fingerprint, &r.VideoHash, &r.State, &r.Notes, &r.Revision, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return r, err
}

func (s *Store) SaveReview(ctx context.Context, r Review) (Review, error) {
	if _, err := s.GetReview(ctx, r.Episode, r.Language, r.VideoID); err != nil {
		return Review{}, err
	}
	if !hash(r.Fingerprint) || !hash(r.VideoHash) || !validText(r.Notes, 4000) || r.Revision < 0 {
		return Review{}, errors.New("invalid review")
	}
	if r.State != "pending" && r.State != "reviewed" && r.State != "changes_requested" {
		return Review{}, errors.New("invalid review decision")
	}
	previous := r.Revision
	r.Revision++
	r.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	var result sql.Result
	var err error
	if previous == 0 {
		result, err = s.db.ExecContext(ctx, `INSERT INTO reviews VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`, r.Episode, r.Language, r.VideoID, r.Fingerprint, r.VideoHash, r.State, r.Notes, r.Revision, r.UpdatedAt)
	} else {
		result, err = s.db.ExecContext(ctx, `UPDATE reviews SET fingerprint=?,video_hash=?,state=?,notes=?,revision=?,updated_at=? WHERE episode=? AND language=? AND video_id=? AND revision=?`, r.Fingerprint, r.VideoHash, r.State, r.Notes, r.Revision, r.UpdatedAt, r.Episode, r.Language, r.VideoID, previous)
	}
	if err != nil {
		return Review{}, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return Review{}, err
	}
	if n != 1 {
		return Review{}, errors.New("review changed; reload before saving")
	}
	return r, nil
}

func (s *Store) RequireReview(ctx context.Context, d Draft) error {
	r, err := s.GetReview(ctx, d.Key.Episode, d.Key.Language, d.VideoID)
	if err != nil {
		return err
	}
	if r.State != "reviewed" || r.Fingerprint != d.DeliveryFingerprint || r.VideoHash != d.VideoHash {
		return errors.New("review the current video in Review before sending")
	}
	return nil
}
