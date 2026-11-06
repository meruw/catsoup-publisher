package drafts

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"
)

type PostFilter struct {
	EpisodeMatches []string `json:"episodeMatches"`
	Episode        string   `json:"episode"`
	Search         string   `json:"search"`
	Platform       string   `json:"platform"`
	Language       string   `json:"language"`
	State          string   `json:"state"`
	Offset         int      `json:"offset"`
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
	if err := validatePostFilter(f); err != nil {
		return out, err
	}
	rows, err := s.db.QueryContext(ctx, postSnapshotSQL+` SELECT episode,language,platform,account,title,video_id,delivery_mode,account_name,connected,status,reference,updated FROM posts
 WHERE (?='' OR episode=?) AND (?='' OR platform=?) AND (?='' OR language=?) AND (?='' OR status=?)
 AND (?='' OR instr(lower(title||' '||episode||' '||account_name||' '||account),lower(?))>0)
 ORDER BY updated DESC,episode,language,platform,account LIMIT 51 OFFSET ?`, time.Now().Add(-30*time.Minute).Unix(), f.Episode, f.Episode, f.Platform, f.Platform, f.Language, f.Language, f.State, f.State, f.Search, f.Search, f.Offset)
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

func validatePostFilter(f PostFilter) error {
	if len(f.EpisodeMatches) > 5000 {
		return errors.New("too many catalog matches")
	}
	for _, id := range f.EpisodeMatches {
		if !identifier.MatchString(id) || len(id) > 160 {
			return errors.New("invalid catalog match")
		}
	}
	if len(f.Search) > 200 || len(f.Episode) > 160 || f.Offset < 0 || f.Offset > 1000000 {
		return errors.New("invalid post filter")
	}
	if f.Platform != "" && !slices.Contains([]string{"youtube", "tiktok", "instagram", "x"}, f.Platform) {
		return errors.New("invalid platform")
	}
	if f.Language != "" && f.Language != "es" && f.Language != "en" {
		return errors.New("invalid language")
	}
	if f.State != "" && !slices.Contains([]string{"draft", "approved", "published", "private", "restricted", "inbox", "sending", "check", "failed"}, f.State) {
		return errors.New("invalid status")
	}

	return nil
}

const postSnapshotSQL = `WITH posts AS (
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
 )`

type PostGroup struct {
	Episode   string         `json:"episode"`
	Title     string         `json:"title"`
	Languages string         `json:"languages"`
	Platforms int            `json:"platforms"`
	Total     int            `json:"total"`
	Counts    map[string]int `json:"counts"`
	UpdatedAt string         `json:"updatedAt"`
}
type PostGroupPage struct {
	Items []PostGroup `json:"items"`
	More  bool        `json:"more"`
}

// Paginate videos, not destinations. Filters select matching videos; each summary
// includes all destinations so a filtered page cannot imply complete publication.
func (s *Store) PostGroups(ctx context.Context, f PostFilter) (PostGroupPage, error) {
	out := PostGroupPage{Items: []PostGroup{}}
	if err := validatePostFilter(f); err != nil {
		return out, err
	}
	matched, _ := json.Marshal(f.EpisodeMatches)
	rows, err := s.db.QueryContext(ctx, postSnapshotSQL+`, matching AS (
 SELECT episode FROM posts
 WHERE (?='' OR episode=?) AND (?='' OR platform=?) AND (?='' OR language=?) AND (?='' OR status=?)
 AND (?='' OR instr(lower(title||' '||episode||' '||account_name||' '||account),lower(?))>0 OR episode IN (SELECT value FROM json_each(?)))
 ), selected AS (SELECT episode FROM posts WHERE episode IN (SELECT episode FROM matching)
 GROUP BY episode ORDER BY MAX(updated) DESC,episode LIMIT 21 OFFSET ?
 ) SELECT episode,MIN(title),GROUP_CONCAT(DISTINCT language),COUNT(DISTINCT platform),COUNT(*),
 SUM(status='draft'),SUM(status='approved'),SUM(status='published'),SUM(status='private'),SUM(status='restricted'),SUM(status='inbox'),SUM(status='sending'),SUM(status='check'),SUM(status='failed'),MAX(updated)
 FROM posts WHERE episode IN (SELECT episode FROM selected) GROUP BY episode ORDER BY MAX(updated) DESC,episode`,
		time.Now().Add(-30*time.Minute).Unix(), f.Episode, f.Episode, f.Platform, f.Platform, f.Language, f.Language, f.State, f.State, f.Search, f.Search, string(matched), f.Offset)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var p PostGroup
		var counts [9]int
		var updated int64
		if err = rows.Scan(&p.Episode, &p.Title, &p.Languages, &p.Platforms, &p.Total, &counts[0], &counts[1], &counts[2], &counts[3], &counts[4], &counts[5], &counts[6], &counts[7], &counts[8], &updated); err != nil {
			return out, err
		}
		p.Counts = map[string]int{}
		for i, k := range []string{"draft", "approved", "published", "private", "restricted", "inbox", "sending", "check", "failed"} {
			if counts[i] > 0 {
				p.Counts[k] = counts[i]
			}
		}
		p.UpdatedAt = time.Unix(updated, 0).UTC().Format(time.RFC3339)
		out.Items = append(out.Items, p)
	}
	if len(out.Items) > 20 {
		out.More = true
		out.Items = out.Items[:20]
	}
	return out, rows.Err()
}
