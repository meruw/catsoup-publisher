import { Accounts, LinkAccount, UnlinkAccount, ConfigureClient, GoogleClientReady, OpenGoogleSetup, ClearUncertain, OpenXComposer, Receipt, Send } from '../../bindings/catsoup/publisher/uploadservice';
export type { Profile } from '../../bindings/catsoup/publisher/internal/drafts/models';
export const uploads = { accounts: Accounts, link: LinkAccount, unlink: UnlinkAccount, configure: ConfigureClient, googleReady: GoogleClientReady, googleSetup: OpenGoogleSetup, openX: OpenXComposer, receipt: Receipt, send: Send, clear: ClearUncertain };
