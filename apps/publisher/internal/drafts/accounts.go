package drafts

import (
	"context"
	"database/sql"
	"errors"
)

type Profile struct {
	Platform string `json:"platform"`
	Language string `json:"language"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	API      bool   `json:"api"`
}

func (s *Store) Profiles() ([]Profile, error) {
	rows, err := s.db.Query("SELECT platform,language,account_id,name,length(secret)>0 FROM accounts ORDER BY platform,language")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles := []Profile{}
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.Platform, &p.Language, &p.ID, &p.Name, &p.API); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}
func (s *Store) PutProfile(p Profile, secret []byte, expected string) error {
	// Replace only the connection the user saw when starting authorization.
	var res sql.Result
	var err error
	if expected == "" {
		res, err = s.db.Exec("INSERT OR IGNORE INTO accounts(platform,language,account_id,name,secret) VALUES(?,?,?,?,?)", p.Platform, p.Language, p.ID, p.Name, secret)
	} else {
		res, err = s.db.Exec("UPDATE OR IGNORE accounts SET account_id=?,name=?,secret=? WHERE platform=? AND language=? AND account_id=?", p.ID, p.Name, secret, p.Platform, p.Language, expected)
	}
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("account is already assigned to the other language or this connection changed; reload Accounts")
	}
	return nil
}
func (s *Store) RemoveProfile(platform, lang, id string) error {
	res, err := s.db.Exec("DELETE FROM accounts WHERE platform=? AND language=? AND account_id=?", platform, lang, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("account changed; reload Accounts")
	}
	return nil
}

// Serialize refresh-token rotation across Publisher instances only when necessary.
func (s *Store) Credential(platform, lang, id string) ([]byte, error) {
	var b []byte
	err := s.db.QueryRow("SELECT secret FROM accounts WHERE platform=? AND language=? AND account_id=?", platform, lang, id).Scan(&b)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("destination changed; check the account for this language")
	}
	return b, err
}
func (s *Store) RotateCredential(ctx context.Context, platform, lang, id string, rotate func([]byte) ([]byte, error)) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "UPDATE accounts SET secret=secret WHERE platform=? AND language=? AND account_id=?", platform, lang, id); err != nil {
		return err
	}
	var secret []byte
	if err = tx.QueryRowContext(ctx, "SELECT secret FROM accounts WHERE platform=? AND language=? AND account_id=?", platform, lang, id).Scan(&secret); err != nil {
		return err
	}
	replacement, err := rotate(secret)
	if err != nil {
		return err
	}
	if replacement != nil {
		if _, err = tx.ExecContext(ctx, "UPDATE accounts SET secret=? WHERE platform=? AND language=? AND account_id=?", replacement, platform, lang, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}
