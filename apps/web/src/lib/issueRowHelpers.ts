import type { WorkspaceMemberApiResponse } from '../api/types';

export interface MemberLite {
  id: string;
  name: string;
  avatarUrl?: string | null;
}

/** Translate workspace members + assignee ids into a list of MemberLite. */
export function membersFromAssigneeIds(
  members: WorkspaceMemberApiResponse[],
  assigneeIds?: string[],
): MemberLite[] {
  if (!assigneeIds?.length || !members.length) return [];
  const byId = new Map(members.map((m) => [m.member_id, m]));
  const out: MemberLite[] = [];
  for (const id of assigneeIds) {
    const m = byId.get(id);
    if (!m) continue;
    out.push({
      id,
      name: m.member_display_name || (m.member_email ?? 'Unknown'),
      avatarUrl: m.member_avatar ?? null,
    });
  }
  return out;
}

/**
 * True when `targetDate` is in the past (older than ~1 day) and the issue isn't
 * already completed/cancelled. Shared so editable date cells reuse the same
 * overdue cue the read-only DueDateCell shows. `now` is passed in for purity.
 */
export function isOverdue(
  targetDate: string | null | undefined,
  stateGroup: string | undefined,
  now: number,
): boolean {
  if (!targetDate) return false;
  const t = Date.parse(targetDate);
  if (Number.isNaN(t)) return false;
  if (stateGroup === 'completed' || stateGroup === 'cancelled') return false;
  return t < now - 24 * 3600 * 1000;
}
