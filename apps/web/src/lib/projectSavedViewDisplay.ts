export type SavedViewDisplayPropertyId =
  | 'id'
  | 'assignee'
  | 'start_date'
  | 'due_date'
  | 'labels'
  | 'priority'
  | 'state'
  | 'sub_work_count'
  | 'attachment_count'
  | 'link'
  | 'estimate'
  | 'module'
  | 'cycle';

export type SavedViewGroupBy =
  | 'states'
  | 'priority'
  | 'cycle'
  | 'module'
  | 'labels'
  | 'assignees'
  | 'created_by'
  | 'none';

export type SavedViewOrderBy =
  | 'manual'
  | 'last_created'
  | 'last_updated'
  | 'start_date'
  | 'due_date'
  | 'priority';

export type SavedViewOrderDirection = 'asc' | 'desc';

export interface SavedViewDisplaySettings {
  displayProperties: Set<SavedViewDisplayPropertyId>;
  groupBy: SavedViewGroupBy;
  orderBy: SavedViewOrderBy;
  orderDirection: SavedViewOrderDirection;
  showSubWorkItems: boolean;
}

export const ALL_SAVED_VIEW_DISPLAY_PROPERTIES: SavedViewDisplayPropertyId[] = [
  'id',
  'assignee',
  'start_date',
  'due_date',
  'labels',
  'priority',
  'state',
  'sub_work_count',
  'attachment_count',
  'link',
  'estimate',
  'module',
  'cycle',
];

export const DEFAULT_SAVED_VIEW_DISPLAY: SavedViewDisplaySettings = {
  displayProperties: new Set(ALL_SAVED_VIEW_DISPLAY_PROPERTIES),
  groupBy: 'states',
  orderBy: 'manual',
  orderDirection: 'asc',
  showSubWorkItems: false,
};

export function cloneDefaultSettings(): SavedViewDisplaySettings {
  return {
    displayProperties: new Set(DEFAULT_SAVED_VIEW_DISPLAY.displayProperties),
    groupBy: DEFAULT_SAVED_VIEW_DISPLAY.groupBy,
    orderBy: DEFAULT_SAVED_VIEW_DISPLAY.orderBy,
    orderDirection: DEFAULT_SAVED_VIEW_DISPLAY.orderDirection,
    showSubWorkItems: DEFAULT_SAVED_VIEW_DISPLAY.showSubWorkItems,
  };
}

function isValidPropertyId(x: string): x is SavedViewDisplayPropertyId {
  return (ALL_SAVED_VIEW_DISPLAY_PROPERTIES as string[]).includes(x);
}

export interface PersistedSavedViewDisplay {
  displayProperties: string[];
  groupBy: string;
  orderBy: string;
  orderDirection: string;
  showSubWorkItems: boolean;
}

const ORDER_DIRECTIONS: SavedViewOrderDirection[] = ['asc', 'desc'];

const GROUP_OPTIONS: SavedViewGroupBy[] = [
  'states',
  'priority',
  'cycle',
  'module',
  'labels',
  'assignees',
  'created_by',
  'none',
];

const ORDER_OPTIONS: SavedViewOrderBy[] = [
  'manual',
  'last_created',
  'last_updated',
  'start_date',
  'due_date',
  'priority',
];

export function parsePersistedSavedViewDisplay(
  raw: string | null,
): SavedViewDisplaySettings | null {
  if (!raw) return null;
  try {
    const p = JSON.parse(raw) as PersistedSavedViewDisplay;
    const props = new Set<SavedViewDisplayPropertyId>();
    if (Array.isArray(p.displayProperties)) {
      for (const id of p.displayProperties) {
        if (typeof id === 'string' && isValidPropertyId(id)) props.add(id);
      }
    }
    const groupBy = GROUP_OPTIONS.includes(p.groupBy as SavedViewGroupBy)
      ? (p.groupBy as SavedViewGroupBy)
      : 'states';
    const orderBy = ORDER_OPTIONS.includes(p.orderBy as SavedViewOrderBy)
      ? (p.orderBy as SavedViewOrderBy)
      : 'manual';
    const orderDirection = ORDER_DIRECTIONS.includes(p.orderDirection as SavedViewOrderDirection)
      ? (p.orderDirection as SavedViewOrderDirection)
      : 'asc';
    return {
      displayProperties: props.size > 0 ? props : new Set(ALL_SAVED_VIEW_DISPLAY_PROPERTIES),
      groupBy,
      orderBy,
      orderDirection,
      showSubWorkItems: Boolean(p.showSubWorkItems),
    };
  } catch {
    return null;
  }
}

export function serializeSettings(s: SavedViewDisplaySettings): string {
  return JSON.stringify({
    displayProperties: [...s.displayProperties],
    groupBy: s.groupBy,
    orderBy: s.orderBy,
    orderDirection: s.orderDirection,
    showSubWorkItems: s.showSubWorkItems,
  });
}

export const SAVED_VIEW_DISPLAY_PROPERTY_LABELS: Record<SavedViewDisplayPropertyId, string> = {
  id: 'ID',
  assignee: 'Assignee',
  start_date: 'Start date',
  due_date: 'Due date',
  labels: 'Labels',
  priority: 'Priority',
  state: 'State',
  sub_work_count: 'Sub-work item count',
  attachment_count: 'Attachment count',
  link: 'Link',
  estimate: 'Estimate',
  module: 'Module',
  cycle: 'Cycle',
};
