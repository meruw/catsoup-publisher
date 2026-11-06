import { Events } from '@wailsio/runtime';
import { SetPostApproval, ListDrafts, GetDraft, SaveDraft, GetReview, SaveReview, CopyDraft, CopyUploadPath, OpenVideoFolder, EnableCloseGuard, ConfirmClose } from '../../bindings/catsoup/publisher/draftservice';
export type { Review } from '../../bindings/catsoup/publisher/internal/drafts/models';
export type { Draft, Key as DraftKey } from '../../bindings/catsoup/publisher/internal/drafts/models';
export const drafts = {
  approve: SetPostApproval, review: GetReview, saveReview: SaveReview,
  list: ListDrafts, get: GetDraft, save: SaveDraft, copy: CopyDraft,
  reveal: OpenVideoFolder, copyUploadPath: CopyUploadPath, enableCloseGuard: EnableCloseGuard, confirmClose: ConfirmClose,
  onCloseRequested: (callback: () => void) => Events.On('publisher:close-requested', callback),
};
