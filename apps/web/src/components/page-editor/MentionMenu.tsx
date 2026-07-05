import { Avatar } from '../ui';
import { getImageUrl } from '../../lib/utils';
import type { SuggestionMenuProps } from './suggestionPopup';
import type { MentionItem } from './mentionTypes';

export function MentionMenu({ items, selectedIndex, onSelect }: SuggestionMenuProps<MentionItem>) {
  if (items.length === 0) {
    return (
      <div className="w-60 rounded-md border border-(--border-subtle) bg-(--bg-surface-1) px-3 py-2 text-sm text-(--txt-tertiary) shadow-(--shadow-overlay)">
        No members found
      </div>
    );
  }
  return (
    <div className="max-h-72 w-60 overflow-y-auto rounded-md border border-(--border-subtle) bg-(--bg-surface-1) py-1 shadow-(--shadow-overlay)">
      {items.map((item, i) => (
        <button
          key={item.id}
          type="button"
          onMouseDown={(e) => {
            e.preventDefault();
            onSelect(i);
          }}
          className={
            'flex w-full items-center gap-2 px-2 py-1.5 text-left text-sm text-(--txt-primary) ' +
            (i === selectedIndex ? 'bg-(--bg-layer-1-hover)' : 'hover:bg-(--bg-layer-1-hover)')
          }
        >
          <Avatar name={item.label} src={getImageUrl(item.avatarUrl) ?? undefined} size="sm" />
          <span className="truncate">{item.label}</span>
        </button>
      ))}
    </div>
  );
}
