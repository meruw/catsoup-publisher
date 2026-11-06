package drafts

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type PublishingDefault struct {
	Platform  string `json:"platform"`
	Language  string `json:"language"`
	AccountID string `json:"accountId"`
	Mode      string `json:"mode"`
}

func (s *Store) PublishingDefaults(ctx context.Context) ([]PublishingDefault, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT a.platform,a.language,a.account_id,COALESCE(p.mode,'draft') FROM accounts a LEFT JOIN publishing_defaults p ON p.platform=a.platform AND p.language=a.language AND p.account_id=a.account_id ORDER BY a.language,a.platform`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []PublishingDefault{}
	for rows.Next() {
		var p PublishingDefault
		if err = rows.Scan(&p.Platform, &p.Language, &p.AccountID, &p.Mode); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
func (s *Store) SetPublishingDefault(ctx context.Context, p PublishingDefault) error {
	if err := ValidateKey(Key{"catsoup-settings", p.Language, p.Platform, "account:" + p.AccountID}); err != nil {
		return err
	}
	if p.AccountID == "" || (p.Mode != "draft" && p.Mode != "publish") || (p.Platform == "tiktok" && p.Mode != "draft") {
		return errors.New("unsupported publishing default")
	}
	res, err := s.db.ExecContext(ctx, `INSERT INTO publishing_defaults(platform,language,account_id,mode) SELECT platform,language,account_id,? FROM accounts WHERE platform=? AND language=? AND account_id=? ON CONFLICT(platform,language,account_id) DO UPDATE SET mode=excluded.mode`, p.Mode, p.Platform, p.Language, p.AccountID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("account changed; reload Settings")
	}
	return nil
}
func (s *Store) DefaultMode(ctx context.Context, k Key) (string, error) {
	if !strings.HasPrefix(k.Account, "account:") {
		return "draft", nil
	}
	var mode string
	err := s.db.QueryRowContext(ctx, `SELECT p.mode FROM publishing_defaults p JOIN accounts a ON a.platform=p.platform AND a.language=p.language AND a.account_id=p.account_id WHERE p.platform=? AND p.language=? AND p.account_id=?`, k.Platform, k.Language, strings.TrimPrefix(k.Account, "account:")).Scan(&mode)
	if errors.Is(err, sql.ErrNoRows) {
		return "draft", nil
	}
	return mode, err
}
