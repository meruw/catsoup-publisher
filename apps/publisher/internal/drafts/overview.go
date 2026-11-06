package drafts

import (
	"context"
	"errors"
	"slices"
	"time"
)

type PostFilter struct {
	Search   string `json:"search"`
	Platform string `json:"platform"`
	Language string `json:"language"`
	State    string `json:"state"`
	Offset   int    `json:"offset"`
}
type PostStatus struct {
	Key         Key    `json:"key"`
	Title       string `json:"title"`
	VideoID     string `json:"videoId"`
	AccountName string `json:"accountName"`
	Connected   bool   `json:"connected"`
	Mode        string `json:"mode"`
	State       string `json:"state"`
	Reference   string `json:"reference"`
	UpdatedAt   string `json:"updatedAt"`
}
type PostPage struct {
	Items []PostStatus `json:"items"`
	More  bool         `json:"more"`
}

// Local snapshot only. No credentials or provider requests are exposed by this query.
func (s *Store) PostStatuses(ctx context.Context, f PostFilter) (PostPage, error) {
	out := PostPage{Items: []PostStatus{}}
	if len(f.Search) > 200 || f.Offset < 0 || f.Offset > 1000000 {
		return out, errors.New("invalid post filter")
	}
	if f.Platform != "" && !slices.Contains([]string{"youtube", "tiktok", "instagram", "x"}, f.Platform) {
		return out, errors.New("invalid platform")
	}
	if f.Language != "" && f.Language != "es" && f.Language != "en" {
		return out, errors.New("invalid language")
	}
	if f.State != "" && !slices.Contains([]string{"draft", "approved", "published", "private", "restricted", "inbox", "sending", "check", "failed"}, f.State) {
		return out, errors.New("invalid status")
	}
	rows, err := s.db.QueryContext(ctx, `WITH posts AS (
 SELECT d.episode,d.language,d.platform,d.account,d.title,d.video_id,d.delivery_mode,
 COALESCE(a.name,d.account) AS account_name,a.account_id IS NOT NULL AS connected,
 CASE WHEN u.state='sent' THEN CASE d.platform WHEN 'youtube' THEN 'private' WHEN 'tiktok' THEN 'inbox' ELSE 'published' END
 WHEN u.state IN ('sending','publishing') THEN CASE WHEN u.started<? THEN 'check' ELSE 'sending' END
 WHEN u.state IN ('check','check_publish') THEN 'check'
 WHEN u.state IS NOT NULL THEN u.state
 WHEN d.approved_revision=d.revision THEN 'approved' ELSE 'draft' END AS status,
 COALESCE(u.reference,'') AS reference,MAX(COALESCE(unixepoch(d.updated_at),0),COALESCE(u.started,0)) AS updated
 FROM drafts d LEFT JOIN accounts a ON a.platform=d.platform AND a.language=d.language AND 'account:'||a.account_id=d.account
 LEFT JOIN upload_attempts u ON u.platform=d.platform AND 'account:'||u.account_id=d.account AND u.video_hash=d.video_hash
 ) SELECT episode,language,platform,account,title,video_id,delivery_mode,account_name,connected,status,reference,updated FROM posts
 WHERE (?='' OR platform=?) AND (?='' OR language=?) AND (?='' OR status=?)
 AND (?='' OR instr(lower(title||' '||episode||' '||account_name||' '||account),lower(?))>0)
 ORDER BY updated DESC,episode,language,platform,account LIMIT 51 OFFSET ?`, time.Now().Add(-30*time.Minute).Unix(), f.Platform, f.Platform, f.Language, f.Language, f.State, f.State, f.Search, f.Search, f.Offset)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var p PostStatus
		var updated int64
		if err = rows.Scan(&p.Key.Episode, &p.Key.Language, &p.Key.Platform, &p.Key.Account, &p.Title, &p.VideoID, &p.Mode, &p.AccountName, &p.Connected, &p.State, &p.Reference, &updated); err != nil {
			return out, err
		}
		p.UpdatedAt = time.Unix(updated, 0).UTC().Format(time.RFC3339)
		out.Items = append(out.Items, p)
	}
	if len(out.Items) > 50 {
		out.More = true
		out.Items = out.Items[:50]
	}
	return out, rows.Err()
}
