import { ReactRenderer } from '@tiptap/react';
import type { SuggestionOptions, SuggestionProps } from '@tiptap/suggestion';
import type { ComponentType } from 'react';

export interface SuggestionMenuProps<T> {
  items: T[];
  selectedIndex: number;
  onSelect: (index: number) => void;
}

/**
 * Builds a TipTap suggestion `render` that shows `Menu` in a body-level popup
 * positioned under the caret, with arrow-key navigation and Enter to pick. All
 * mutable state lives in the returned closure so the menu component stays pure.
 */
export function createSuggestionRenderer<T>(
  Menu: ComponentType<SuggestionMenuProps<T>>,
): NonNullable<SuggestionOptions<T>['render']> {
  return () => {
    let renderer: ReactRenderer | null = null;
    let popup: HTMLDivElement | null = null;
    let items: T[] = [];
    let selectedIndex = 0;
    let choose: (item: T) => void = () => {};

    const paint = () => {
      renderer?.updateProps({ items, selectedIndex, onSelect: (i: number) => choose(items[i]) });
    };
    const place = (rect?: DOMRect | null) => {
      if (!popup || !rect) return;
      popup.style.top = `${rect.bottom + window.scrollY + 4}px`;
      popup.style.left = `${rect.left + window.scrollX}px`;
    };

    return {
      onStart: (props: SuggestionProps<T>) => {
        items = props.items;
        selectedIndex = 0;
        choose = (item) => props.command(item);
        renderer = new ReactRenderer(Menu, {
          props: { items, selectedIndex, onSelect: (i: number) => choose(items[i]) },
          editor: props.editor,
        });
        popup = document.createElement('div');
        popup.style.position = 'absolute';
        popup.style.zIndex = '10200';
        popup.appendChild(renderer.element);
        document.body.appendChild(popup);
        place(props.clientRect?.());
      },
      onUpdate: (props: SuggestionProps<T>) => {
        items = props.items;
        selectedIndex = 0;
        choose = (item) => props.command(item);
        paint();
        place(props.clientRect?.());
      },
      onKeyDown: (props: { event: KeyboardEvent }) => {
        if (props.event.key === 'Escape') return false;
        if (items.length === 0) return false;
        if (props.event.key === 'ArrowUp') {
          selectedIndex = (selectedIndex + items.length - 1) % items.length;
          paint();
          return true;
        }
        if (props.event.key === 'ArrowDown') {
          selectedIndex = (selectedIndex + 1) % items.length;
          paint();
          return true;
        }
        if (props.event.key === 'Enter') {
          choose(items[selectedIndex]);
          return true;
        }
        return false;
      },
      onExit: () => {
        popup?.remove();
        renderer?.destroy();
        popup = null;
        renderer = null;
      },
    };
  };
}
