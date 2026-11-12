import { Events } from '@wailsio/runtime';
import { ImportCover, CoverPreview, CompletedLanguages, PostGroups, PostStatuses, GetPublishingDefaults, SetPublishingDefault, SetPostApproval, ListDrafts, GetDraft, SaveDraft, GetReview, SaveReview, CopyDraft, CopyUploadPath, OpenVideoFolder, EnableCloseGuard, ConfirmClose } from '../../bindings/catsoup/publisher/draftservice';
export type { Review } from '../../bindings/catsoup/publisher/internal/drafts/models';
export type { Draft, Key as DraftKey } from '../../bindings/catsoup/publisher/internal/drafts/models';
export const drafts = {
 importCover:ImportCover,coverPreview:CoverPreview,
  completedLanguages: CompletedLanguages, groups: PostGroups, statuses: PostStatuses, publishingDefaults: GetPublishingDefaults, setPublishingDefault: SetPublishingDefault,
  approve: SetPostApproval, review: GetReview, saveReview: SaveReview,
  list: ListDrafts, get: GetDraft, save: SaveDraft, copy: CopyDraft,
  reveal: OpenVideoFolder, copyUploadPath: CopyUploadPath, enableCloseGuard: EnableCloseGuard, confirmClose: ConfirmClose,
  onCloseRequested: (callback: () => void) => Events.On('publisher:close-requested', callback),
};

export type { PublishingDefault } from "../../bindings/catsoup/publisher/internal/drafts/models";

export type { PostStatus, PostPage, PostGroup } from "../../bindings/catsoup/publisher/internal/drafts/models";

export type {Completion} from "../../bindings/catsoup/publisher/internal/drafts/models";
