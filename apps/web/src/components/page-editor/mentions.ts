import Mention from '@tiptap/extension-mention';
import { createSuggestionRenderer } from './suggestionPopup';
import { MentionMenu } from './MentionMenu';
import type { MentionItem } from './mentionTypes';

/**
 * @mention of workspace members. `getItems` is read at suggestion time so the
 * editor picks up members loaded after it mounts.
 */
export const createMention = (getItems: () => MentionItem[]) =>
  Mention.configure({
    HTMLAttributes: { class: 'page-mention' },
    suggestion: {
      char: '@',
      items: ({ query }) => {
        const q = query.trim().toLowerCase();
        const list = getItems();
        const matched = q ? list.filter((i) => i.label.toLowerCase().includes(q)) : list;
        return matched.slice(0, 8);
      },
      render: createSuggestionRenderer(MentionMenu),
    },
  });
