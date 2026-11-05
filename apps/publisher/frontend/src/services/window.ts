import { Events, Window } from '@wailsio/runtime';

// Narrow native adapter: no additional service, polling or window manager.
export const desktopWindow = {
  minimise: () => Window.Minimise(),
  toggleMaximise: () => Window.ToggleMaximise(),
  close: () => Window.Close(),
  isMaximised: () => Window.IsMaximised(),
  onMaximisedChange(callback: (maximised: boolean) => void) {
    const stopMax = Events.On(Events.Types.Common.WindowMaximise, () => callback(true));
    const stopRestore = Events.On(Events.Types.Common.WindowUnMaximise, () => callback(false));
    return () => { stopMax(); stopRestore(); };
  },
};
