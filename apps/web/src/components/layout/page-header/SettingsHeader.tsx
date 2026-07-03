import { IconSettings } from './icons';

export function SettingsHeader() {
  return (
    <>
      <div className="flex items-center gap-2 text-sm font-medium text-(--txt-secondary)">
        <span className="flex size-5 items-center justify-center text-(--txt-icon-tertiary)">
          <IconSettings />
        </span>
        Settings
      </div>
      <div className="flex items-center gap-2" />
    </>
  );
}
