import { IconUser } from './icons';

export function YourWorkHeader() {
  return (
    <div className="flex items-center gap-2 text-sm font-medium text-(--txt-secondary)">
      <span className="flex size-5 items-center justify-center text-(--txt-icon-tertiary)">
        <IconUser />
      </span>
      Your work
    </div>
  );
}
