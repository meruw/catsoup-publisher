import { Events } from '@wailsio/runtime';
import { ListDrafts, GetDraft, SaveDraft, CopyDraft, CopyUploadPath, OpenVideoFolder, EnableCloseGuard, ConfirmClose } from '../../bindings/catsoup/publisher/draftservice';
export type { Draft, Key as DraftKey } from '../../bindings/catsoup/publisher/internal/drafts/models';
export const drafts = {
  list: ListDrafts, get: GetDraft, save: SaveDraft, copy: CopyDraft,
  reveal: OpenVideoFolder, copyUploadPath: CopyUploadPath, enableCloseGuard: EnableCloseGuard, confirmClose: ConfirmClose,
  onCloseRequested: (callback: () => void) => Events.On('publisher:close-requested', callback),
};
