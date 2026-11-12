package drafts

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
)

// Cover contains only a content address; list views never load image bytes.
type Cover struct {
	Kind   string `json:"kind"`
	Hash   string `json:"hash"`
	TimeMS int64  `json:"timeMs"`
}

func ValidateCover(platform string, c *Cover) error {
	if c == nil {
		return nil
	}
	if !hash(c.Hash) || c.TimeMS < 0 || c.TimeMS > 24*60*60*1000 {
		return errors.New("invalid cover")
	}
	if c.Kind != "image" && c.Kind != "frame" {
		return errors.New("choose an image or video frame")
	}
	if platform != "youtube" && ((platform != "instagram" && platform != "tiktok") || c.Kind != "frame") {
		return errors.New("this connector does not support that cover option")
	}
	return nil
}
func (s *Store) ImportCover(ctx context.Context, b []byte) (string, error) {
	if len(b) == 0 || len(b) > 2<<20 {
		return "", errors.New("choose a PNG or JPEG up to 2 MiB")
	}
	cfg, kind, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil || (kind != "png" && kind != "jpeg") || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 4096 || cfg.Height > 4096 || int64(cfg.Width)*int64(cfg.Height) > 16_000_000 {
		return "", errors.New("invalid PNG/JPEG or image larger than 4096 pixels / 16 megapixels")
	}
	if _, _, err = image.Decode(bytes.NewReader(b)); err != nil {
		return "", errors.New("could not decode the complete image")
	}
	sum := sha256.Sum256(b)
	id := hex.EncodeToString(sum[:])
	_, err = s.db.ExecContext(ctx, "INSERT OR IGNORE INTO cover_assets(hash,mime,image) VALUES(?,?,?)", id, "image/"+kind, b)
	return id, err
}
func (s *Store) CoverImage(ctx context.Context, id string) ([]byte, string, error) {
	if !hash(id) {
		return nil, "", errors.New("invalid cover identifier")
	}
	var b []byte
	var mime string
	err := s.db.QueryRowContext(ctx, "SELECT image,mime FROM cover_assets WHERE hash=?", id).Scan(&b, &mime)
	if err != nil {
		return nil, "", errors.New("saved cover unavailable; choose it again")
	}
	sum := sha256.Sum256(b)
	if len(b) > 2<<20 || hex.EncodeToString(sum[:]) != id {
		return nil, "", errors.New("saved cover failed integrity check")
	}
	return b, mime, nil
}
func (s *Store) SaveCoverResult(ctx context.Context, d Draft, state string) error {
	if d.Cover == nil {
		return errors.New("no cover selected")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO cover_results(platform,account,video_hash,cover_hash,state) VALUES(?,?,?,?,?) ON CONFLICT(platform,account,video_hash) DO UPDATE SET cover_hash=excluded.cover_hash,state=excluded.state`, d.Key.Platform, d.Key.Account, d.VideoHash, d.Cover.Hash, state)
	return err
}
func (s *Store) CoverResult(ctx context.Context, d Draft) string {
	if d.Cover == nil {
		return ""
	}
	var state string
	_ = s.db.QueryRowContext(ctx, "SELECT state FROM cover_results WHERE platform=? AND account=? AND video_hash=? AND cover_hash=?", d.Key.Platform, d.Key.Account, d.VideoHash, d.Cover.Hash).Scan(&state)
	return state
}
