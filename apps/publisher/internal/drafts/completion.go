package drafts

import "context"

// Completion is a local outcome for one delivery version and language. A caller
// must match every language and fingerprint of the current validated delivery.
type Completion struct {
	Episode     string `json:"episode"`
	Fingerprint string `json:"fingerprint"`
	Language    string `json:"language"`
}

// CompletedLanguages includes only languages with at least one connected non-X
// destination and a successful current-file receipt for every such destination.
// It does not change receipts, approvals, drafts or provider state.
func (s *Store) CompletedLanguages(ctx context.Context) ([]Completion, error) {
	rows, err := s.db.QueryContext(ctx, `WITH candidates AS (
 SELECT DISTINCT episode,delivery_fingerprint,language FROM drafts WHERE platform<>'x'
 ) SELECT c.episode,c.delivery_fingerprint,c.language FROM candidates c
 WHERE EXISTS(SELECT 1 FROM accounts a WHERE a.language=c.language AND a.platform<>'x')
 AND NOT EXISTS(
 SELECT 1 FROM accounts a WHERE a.language=c.language AND a.platform<>'x'
 AND NOT EXISTS(
 SELECT 1 FROM drafts d JOIN upload_attempts u ON u.platform=d.platform AND u.account_id=a.account_id AND u.video_hash=d.video_hash
 WHERE d.episode=c.episode AND d.delivery_fingerprint=c.delivery_fingerprint AND d.language=c.language
 AND d.platform=a.platform AND d.account='account:'||a.account_id
 AND (u.state='published' OR (a.platform='tiktok' AND u.state='inbox'))
 )) ORDER BY c.episode,c.language`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Completion{}
	for rows.Next() {
		var c Completion
		if err = rows.Scan(&c.Episode, &c.Fingerprint, &c.Language); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
