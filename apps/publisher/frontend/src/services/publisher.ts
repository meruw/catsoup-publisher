// Keep desktop transport out of UI components. A different adapter can replace it.
import { GetEpisode, ListDeliveries } from '../../bindings/catsoup/publisher/publisherservice';
export type { Entry, Episode, Video } from '../../bindings/catsoup/publisher/internal/delivery/models';
export const publisher = { list: ListDeliveries, get: GetEpisode };
export function mediaURL(episode: string, video: string): string {
  return '/media?' + new URLSearchParams({ episode, video }).toString();
}
