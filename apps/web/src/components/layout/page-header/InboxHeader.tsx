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
        {/* Not implemented yet — disabled rather than rendered as inert-but-clickable. */}
        <button
          type="button"
          disabled
          className="flex size-8 items-center justify-center rounded-md text-(--txt-icon-tertiary) opacity-40 cursor-not-allowed"
          aria-label="Mark as read"
          title="Coming soon"
        >
          <IconCheck />
        </button>
        <button
          type="button"
          disabled
          className="flex size-8 items-center justify-center rounded-md text-(--txt-icon-tertiary) opacity-40 cursor-not-allowed"
          aria-label="Archive"
          title="Coming soon"
        >
          <IconArchive />
        </button>
        <button
          type="button"
          disabled
          className="flex size-8 items-center justify-center rounded-md text-(--txt-icon-tertiary) opacity-40 cursor-not-allowed"
          aria-label="Filters"
          title="Coming soon"
        >
          <IconFilter />
        </button>
        <button
          type="button"
          disabled
          className="flex size-8 items-center justify-center rounded-md text-(--txt-icon-tertiary) opacity-40 cursor-not-allowed"
          aria-label="More options"
          title="Coming soon"
        >
          <IconMoreVertical />
        </button>
      </div>
    </>
  );
}
