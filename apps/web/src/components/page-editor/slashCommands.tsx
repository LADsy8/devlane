import { Extension, type Range } from '@tiptap/core';
import Suggestion from '@tiptap/suggestion';
import type { Editor } from '@tiptap/react';
import {
  Code2,
  Heading1,
  Heading2,
  Heading3,
  Image as ImageIcon,
  List,
  ListOrdered,
  ListTodo,
  Minus,
  Table as TableIcon,
  TextQuote,
  Type,
  type LucideIcon,
} from 'lucide-react';
import { cn } from '../../lib/utils';
import { createSuggestionRenderer, type SuggestionMenuProps } from './suggestionPopup';

interface SlashItem {
  title: string;
  subtitle: string;
  icon: LucideIcon;
  keywords: string[];
  run: (editor: Editor, range: Range) => void;
}

const ITEMS: SlashItem[] = [
  {
    title: 'Text',
    subtitle: 'Plain paragraph',
    icon: Type,
    keywords: ['paragraph', 'body'],
    run: (editor, range) => editor.chain().focus().deleteRange(range).setParagraph().run(),
  },
  {
    title: 'Heading 1',
    subtitle: 'Large section heading',
    icon: Heading1,
    keywords: ['h1', 'title'],
    run: (editor, range) =>
      editor.chain().focus().deleteRange(range).setNode('heading', { level: 1 }).run(),
  },
  {
    title: 'Heading 2',
    subtitle: 'Medium section heading',
    icon: Heading2,
    keywords: ['h2', 'subtitle'],
    run: (editor, range) =>
      editor.chain().focus().deleteRange(range).setNode('heading', { level: 2 }).run(),
  },
  {
    title: 'Heading 3',
    subtitle: 'Small section heading',
    icon: Heading3,
    keywords: ['h3'],
    run: (editor, range) =>
      editor.chain().focus().deleteRange(range).setNode('heading', { level: 3 }).run(),
  },
  {
    title: 'Bulleted list',
    subtitle: 'Unordered list',
    icon: List,
    keywords: ['ul', 'unordered', 'bullet'],
    run: (editor, range) => editor.chain().focus().deleteRange(range).toggleBulletList().run(),
  },
  {
    title: 'Numbered list',
    subtitle: 'Ordered list',
    icon: ListOrdered,
    keywords: ['ol', 'ordered', 'number'],
    run: (editor, range) => editor.chain().focus().deleteRange(range).toggleOrderedList().run(),
  },
  {
    title: 'To-do list',
    subtitle: 'Checklist',
    icon: ListTodo,
    keywords: ['task', 'checkbox', 'todo'],
    run: (editor, range) => editor.chain().focus().deleteRange(range).toggleTaskList().run(),
  },
  {
    title: 'Quote',
    subtitle: 'Block quote',
    icon: TextQuote,
    keywords: ['blockquote', 'citation'],
    run: (editor, range) => editor.chain().focus().deleteRange(range).toggleBlockquote().run(),
  },
  {
    title: 'Code block',
    subtitle: 'Formatted code',
    icon: Code2,
    keywords: ['pre', 'snippet'],
    run: (editor, range) => editor.chain().focus().deleteRange(range).toggleCodeBlock().run(),
  },
  {
    title: 'Table',
    subtitle: '3x3 table',
    icon: TableIcon,
    keywords: ['grid'],
    run: (editor, range) =>
      editor
        .chain()
        .focus()
        .deleteRange(range)
        .insertTable({ rows: 3, cols: 3, withHeaderRow: true })
        .run(),
  },
  {
    title: 'Image',
    subtitle: 'Embed by URL',
    icon: ImageIcon,
    keywords: ['picture', 'photo', 'embed'],
    run: (editor, range) => {
      const url = window.prompt('Image URL');
      const chain = editor.chain().focus().deleteRange(range);
      if (url) chain.setImage({ src: url }).run();
      else chain.run();
    },
  },
  {
    title: 'Divider',
    subtitle: 'Horizontal rule',
    icon: Minus,
    keywords: ['hr', 'separator', 'line'],
    run: (editor, range) => editor.chain().focus().deleteRange(range).setHorizontalRule().run(),
  },
];

function filterItems(query: string): SlashItem[] {
  const q = query.trim().toLowerCase();
  if (!q) return ITEMS;
  return ITEMS.filter(
    (i) => i.title.toLowerCase().includes(q) || i.keywords.some((k) => k.includes(q)),
  );
}

function SlashMenu({ items, selectedIndex, onSelect }: SuggestionMenuProps<SlashItem>) {
  if (items.length === 0) {
    return (
      <div className="w-64 rounded-md border border-(--border-subtle) bg-(--bg-surface-1) px-3 py-2 text-sm text-(--txt-tertiary) shadow-(--shadow-overlay)">
        No matching blocks
      </div>
    );
  }
  return (
    <div className="max-h-72 w-64 overflow-y-auto rounded-md border border-(--border-subtle) bg-(--bg-surface-1) py-1 shadow-(--shadow-overlay)">
      {items.map((item, i) => (
        <button
          key={item.title}
          type="button"
          onMouseDown={(e) => {
            e.preventDefault();
            onSelect(i);
          }}
          className={cn(
            'flex w-full items-center gap-2 px-2 py-1.5 text-left',
            i === selectedIndex ? 'bg-(--bg-layer-1-hover)' : 'hover:bg-(--bg-layer-1-hover)',
          )}
        >
          <span className="grid size-7 shrink-0 place-items-center rounded border border-(--border-subtle) text-(--txt-secondary)">
            <item.icon size={15} />
          </span>
          <span className="min-w-0">
            <span className="block truncate text-sm text-(--txt-primary)">{item.title}</span>
            <span className="block truncate text-xs text-(--txt-tertiary)">{item.subtitle}</span>
          </span>
        </button>
      ))}
    </div>
  );
}

/**
 * Slash-command menu: type "/" to insert headings, lists, quotes, code, tables,
 * images, and dividers.
 */
export const SlashCommand = Extension.create({
  name: 'slashCommand',
  addProseMirrorPlugins() {
    return [
      Suggestion<SlashItem>({
        editor: this.editor,
        char: '/',
        allowSpaces: false,
        startOfLine: false,
        items: ({ query }) => filterItems(query),
        command: ({ editor, range, props }) => props.run(editor, range),
        render: createSuggestionRenderer(SlashMenu),
      }),
    ];
  },
});
