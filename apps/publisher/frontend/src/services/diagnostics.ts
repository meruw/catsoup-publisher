import { Clipboard, Events } from '@wailsio/runtime';
import { Query, Listen, CopyEvent, Export, OpenFolder } from '../../bindings/catsoup/publisher/diagnosticsservice';
export type { Event as DiagnosticEvent, Query as DiagnosticQuery } from '../../bindings/catsoup/publisher/internal/diagnostics/models';

export const diagnostics = {
  query: Query,
  listen: Listen,
  copyEvent: CopyEvent,
  export: Export,
  openFolder: OpenFolder,
  pasteText: Clipboard.Text,
  onChanged: (callback: () => void) => Events.On('publisher:diagnostics-changed', callback),
};
