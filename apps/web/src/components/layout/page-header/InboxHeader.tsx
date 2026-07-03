import { IconInbox, IconCheck, IconArchive, IconFilter, IconMoreVertical } from './icons';

export function InboxHeader() {
  return (
    <>
      <div className="flex items-center gap-2 text-sm font-medium text-(--txt-secondary)">
        <span className="flex size-5 items-center justify-center text-(--txt-icon-tertiary)">
          <IconInbox />
        </span>
        Inbox
      </div>
      <div className="flex items-center gap-1">
        <button
          type="button"
          className="flex size-8 items-center justify-center rounded-md text-(--txt-icon-tertiary) hover:bg-(--bg-layer-2) hover:text-(--txt-icon-secondary)"
          aria-label="Mark as read"
        >
          <IconCheck />
        </button>
        <button
          type="button"
          className="flex size-8 items-center justify-center rounded-md text-(--txt-icon-tertiary) hover:bg-(--bg-layer-2) hover:text-(--txt-icon-secondary)"
          aria-label="Archive"
        >
          <IconArchive />
        </button>
        <button
          type="button"
          className="flex size-8 items-center justify-center rounded-md text-(--txt-icon-tertiary) hover:bg-(--bg-layer-2) hover:text-(--txt-icon-secondary)"
          aria-label="Filters"
        >
          <IconFilter />
        </button>
        <button
          type="button"
          className="flex size-8 items-center justify-center rounded-md text-(--txt-icon-tertiary) hover:bg-(--bg-layer-2) hover:text-(--txt-icon-secondary)"
          aria-label="More options"
        >
          <IconMoreVertical />
        </button>
      </div>
    </>
  );
}
