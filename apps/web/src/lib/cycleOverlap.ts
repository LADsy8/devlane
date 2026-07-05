import type { CycleApiResponse } from '../api/types';

/** True when two closed date ranges overlap. Parses to timestamps so it is
 *  safe across date-only and full-timestamp string formats. */
function rangesOverlap(aStart: string, aEnd: string, bStart: string, bEnd: string): boolean {
  const as = Date.parse(aStart);
  const ae = Date.parse(aEnd);
  const bs = Date.parse(bStart);
  const be = Date.parse(bEnd);
  if ([as, ae, bs, be].some(Number.isNaN)) return false;
  return as <= be && bs <= ae;
}

/**
 * Cycles whose date range overlaps [start, end], excluding the cycle with id
 * `excludeId`. Cycles missing either date are ignored. Used to warn (not block)
 * about overlapping cycles in the same project.
 */
export function overlappingCycles(
  cycles: CycleApiResponse[],
  start: string | null,
  end: string | null,
  excludeId?: string,
): CycleApiResponse[] {
  if (!start || !end) return [];
  return cycles.filter(
    (c) =>
      c.id !== excludeId &&
      !!c.start_date &&
      !!c.end_date &&
      rangesOverlap(start, end, c.start_date, c.end_date),
  );
}
