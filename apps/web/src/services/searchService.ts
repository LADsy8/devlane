import { apiClient } from '../api/client';

export interface SearchHit {
  id: string;
  name: string;
  project_id?: string;
  project_identifier?: string;
  sequence_id?: number;
}

/** Entity-grouped search results. Keys mirror the backend response. */
export interface SearchResults {
  issue: SearchHit[];
  epic: SearchHit[];
  cycle: SearchHit[];
  module: SearchHit[];
  view: SearchHit[];
  page: SearchHit[];
  project: SearchHit[];
}

export function emptySearchResults(): SearchResults {
  return { issue: [], epic: [], cycle: [], module: [], view: [], page: [], project: [] };
}

export const searchService = {
  /**
   * Cross-entity workspace search. Pass `projectId` to scope the project-owned
   * entities to a single project; pass an `AbortSignal` to cancel in-flight
   * requests when the query changes.
   */
  async search(
    workspaceSlug: string,
    query: string,
    opts?: { projectId?: string; signal?: AbortSignal },
  ): Promise<SearchResults> {
    const params = new URLSearchParams({ query });
    if (opts?.projectId) params.set('project_id', opts.projectId);
    const { data } = await apiClient.get<{ results: SearchResults }>(
      `/api/workspaces/${encodeURIComponent(workspaceSlug)}/search/?${params.toString()}`,
      { signal: opts?.signal },
    );
    return { ...emptySearchResults(), ...(data?.results ?? {}) };
  },
};
