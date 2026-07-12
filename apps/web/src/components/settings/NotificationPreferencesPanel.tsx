import { useEffect, useState } from 'react';
import type { NotificationPreferencesResponse } from '../../api/types';

type Prefs = NotificationPreferencesResponse;
type InAppKey = 'property_change' | 'state_change' | 'issue_completed' | 'comment' | 'mention';
type EmailKey = keyof Prefs & `email_${string}`;

const ROWS: { id: string; label: string; desc: string; inApp: InAppKey; email: EmailKey }[] = [
  {
    id: 'property',
    label: 'Property changes',
    desc: "Notify me when work items' properties like assignees, priority, or estimates change.",
    inApp: 'property_change',
    email: 'email_property_change',
  },
  {
    id: 'state',
    label: 'State change',
    desc: 'Notify me when a work item moves to a different state.',
    inApp: 'state_change',
    email: 'email_state_change',
  },
  {
    id: 'completed',
    label: 'Work item completed',
    desc: 'Notify me when a work item is completed.',
    inApp: 'issue_completed',
    email: 'email_issue_completed',
  },
  {
    id: 'comments',
    label: 'Comments',
    desc: 'Notify me when someone comments on a work item.',
    inApp: 'comment',
    email: 'email_comment',
  },
  {
    id: 'mentions',
    label: 'Mentions',
    desc: 'Notify me when someone mentions me in a comment or description.',
    inApp: 'mention',
    email: 'email_mention',
  },
];

function Toggle({
  checked,
  disabled,
  label,
  onToggle,
}: {
  checked: boolean;
  disabled: boolean;
  label: string;
  onToggle: () => void;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      disabled={disabled}
      onClick={onToggle}
      className={`relative h-6 w-10 shrink-0 rounded-full transition-colors ${checked ? 'bg-(--brand-default)' : 'bg-(--neutral-400)'}`}
    >
      <span
        className={`absolute top-1 left-1 h-4 w-4 rounded-full bg-white shadow transition-transform ${checked ? 'translate-x-4' : 'translate-x-0'}`}
      />
    </button>
  );
}

/**
 * Renders the per-type notification toggles with independent In-app and Email
 * columns. `load`/`save` abstract the scope (account, workspace, or project) so
 * the same panel serves all three.
 */
export function NotificationPreferencesPanel({
  load,
  save,
  title = 'Notifications',
  description = 'Choose which updates reach you in-app and by email.',
}: {
  load: () => Promise<Prefs>;
  save: (partial: Partial<Prefs>) => Promise<Prefs>;
  title?: string;
  description?: string;
}) {
  const [prefs, setPrefs] = useState<Prefs | null>(null);

  useEffect(() => {
    let cancelled = false;
    load()
      .then((p) => {
        if (!cancelled) setPrefs(p);
      })
      .catch(() => {
        if (!cancelled) setPrefs(null);
      });
    return () => {
      cancelled = true;
    };
  }, [load]);

  const toggle = async (key: InAppKey | EmailKey) => {
    if (!prefs) return;
    const next = !prefs[key];
    // Update only this key via a functional update so a concurrent toggle of a
    // different switch is never stomped by this one's success or failure.
    setPrefs((cur) => (cur ? { ...cur, [key]: next } : cur));
    try {
      const saved = await save({ [key]: next });
      setPrefs((cur) => (cur ? { ...cur, [key]: saved[key] } : cur));
    } catch {
      setPrefs((cur) => (cur ? { ...cur, [key]: !next } : cur));
    }
  };

  const loaded = prefs !== null;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-base font-semibold text-(--txt-primary)">{title}</h2>
        <p className="mt-0.5 text-sm text-(--txt-secondary)">{description}</p>
      </div>
      <div className="overflow-hidden rounded-(--radius-md) border border-(--border-subtle)">
        <div className="flex items-center gap-4 border-b border-(--border-subtle) bg-(--bg-layer-1) px-4 py-2 text-xs font-medium text-(--txt-secondary)">
          <span className="flex-1">Type</span>
          <span className="w-14 text-center">In-app</span>
          <span className="w-14 text-center">Email</span>
        </div>
        {ROWS.map(({ id, label, desc, inApp, email }) => (
          <div
            key={id}
            className="flex items-start gap-4 border-b border-(--border-subtle) px-4 py-3 last:border-b-0"
          >
            <div className="flex-1">
              <p className="text-sm font-medium text-(--txt-primary)">{label}</p>
              <p className="mt-0.5 text-sm text-(--txt-secondary)">{desc}</p>
            </div>
            <div className="flex w-14 justify-center">
              <Toggle
                checked={!!prefs?.[inApp]}
                disabled={!loaded}
                label={`${label} in-app`}
                onToggle={() => void toggle(inApp)}
              />
            </div>
            <div className="flex w-14 justify-center">
              <Toggle
                checked={!!prefs?.[email]}
                disabled={!loaded}
                label={`${label} email`}
                onToggle={() => void toggle(email)}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
