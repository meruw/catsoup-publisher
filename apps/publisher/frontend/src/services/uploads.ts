import { Accounts, LinkAccount, UnlinkAccount, ConfigureClient, MetaCallbackURL, SetMetaCallbackURL, CredentialStatuses, ClientReady, OpenGoogleSetup, ClearUncertain, OpenXComposer, Receipt, Send } from '../../bindings/catsoup/publisher/uploadservice';
export type SetupPlatform = 'youtube'|'tiktok'|'instagram';
export type { Profile } from '../../bindings/catsoup/publisher/internal/drafts/models';
export const uploads = { metaCallback: MetaCallbackURL, setMetaCallback: SetMetaCallbackURL, credentialStatuses: CredentialStatuses, accounts: Accounts, link: LinkAccount, unlink: UnlinkAccount, configure: ConfigureClient, clientReady: ClientReady, googleSetup: OpenGoogleSetup, openX: OpenXComposer, receipt: Receipt, send: Send, clear: ClearUncertain };
