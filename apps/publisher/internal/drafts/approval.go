package drafts

import (
	"context"
	"errors"
	"strings"
)

// SetPostApproval records an explicit decision on exactly one saved revision.
func (s *Store) SetPostApproval(ctx context.Context, key Key, revision int64, approved bool) (Draft, error) {
	if err := ValidateKey(key); err != nil {
		return Draft{}, err
	}
	if revision < 1 || !strings.HasPrefix(key.Account, "account:") {
		return Draft{}, errors.New("save a draft for a connected account first")
	}
	value := int64(0)
	if approved {
		value = revision
	}
	result, err := s.db.ExecContext(ctx, `UPDATE drafts SET approved_revision=? WHERE episode=? AND language=? AND platform=? AND account=? AND revision=? AND EXISTS(SELECT 1 FROM accounts a WHERE a.platform=drafts.platform AND a.language=drafts.language AND 'account:'||a.account_id=drafts.account)`, value, key.Episode, key.Language, key.Platform, key.Account, revision)
	if err != nil {
		return Draft{}, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return Draft{}, err
	}
	if n != 1 {
		return Draft{}, errors.New("draft or account changed; reload before approving")
	}
	return s.Get(ctx, key)
}

// Used inside upload reservation statements to check approval atomically with
// the saved revision. Legacy label-keyed drafts retain their original review gate.
const approvalCondition = `((substr(drafts.account,1,8)='account:' AND approved_revision=revision AND revision>0) OR (substr(drafts.account,1,8)<>'account:' AND EXISTS(SELECT 1 FROM reviews r WHERE r.episode=drafts.episode AND r.language=drafts.language AND r.video_id=drafts.video_id AND r.video_hash=drafts.video_hash AND r.fingerprint=drafts.delivery_fingerprint AND r.state='reviewed')))`
