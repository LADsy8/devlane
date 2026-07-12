import { useCallback, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Badge } from '../components/ui';
import { workspaceService } from '../services/workspaceService';
import { projectService } from '../services/projectService';
import { issueService } from '../services/issueService';
import { intakeService } from '../services/intakeService';
import { emitIntakeUpdated } from '../lib/intakeEvents';
import { useDocumentTitle } from '../hooks/useDocumentTitle';
import type {
  IntakeItemApiResponse,
  IssueApiResponse,
  ProjectApiResponse,
  WorkspaceApiResponse,
} from '../api/types';
import type { Priority } from '../types';

const priorityVariant: Record<Priority, 'danger' | 'warning' | 'default' | 'neutral'> = {
  urgent: 'danger',
  high: 'danger',
  medium: 'warning',
  low: 'default',
  none: 'neutral',
};

type TabKey = 'pending' | 'snoozed' | 'accepted' | 'declined' | 'duplicate';
const TABS: { key: TabKey; label: string }[] = [
  { key: 'pending', label: 'Pending' },
  { key: 'snoozed', label: 'Snoozed' },
  { key: 'accepted', label: 'Accepted' },
  { key: 'declined', label: 'Declined' },
  { key: 'duplicate', label: 'Duplicate' },
];

const SNOOZE_PRESETS: { label: string; days: number }[] = [
  { label: '1 day', days: 1 },
  { label: '3 days', days: 3 },
  { label: '1 week', days: 7 },
];

function sourceLabel(source: string): string {
  return source === 'IN_APP' ? 'In-app' : source.replace(/_/g, ' ').toLowerCase();
}

export function IntakePage() {
  const { workspaceSlug, projectId } = useParams<{ workspaceSlug: string; projectId: string }>();
  const [loading, setLoading] = useState(true);
  const [workspace, setWorkspace] = useState<WorkspaceApiResponse | null>(null);
  const [project, setProject] = useState<ProjectApiResponse | null>(null);
  const [tab, setTab] = useState<TabKey>('pending');
  const [items, setItems] = useState<IntakeItemApiResponse[]>([]);
  const [pendingCount, setPendingCount] = useState(0);
  const [activeIssues, setActiveIssues] = useState<IssueApiResponse[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  useDocumentTitle('Intake');

  const reloadPending = useCallback(() => {
    if (!workspaceSlug || !projectId) return;
    intakeService
      .pendingCount(workspaceSlug, projectId)
      .then(setPendingCount)
      .catch(() => setPendingCount(0));
  }, [workspaceSlug, projectId]);

  const loadItems = useCallback(
    (which: TabKey) => {
      if (!workspaceSlug || !projectId) return;
      intakeService
        .list(workspaceSlug, projectId, which)
        .then(setItems)
        .catch(() => setItems([]));
    },
    [workspaceSlug, projectId],
  );

  useEffect(() => {
    if (!workspaceSlug || !projectId) return;
    let cancelled = false;
    queueMicrotask(() => {
      if (!cancelled) setLoading(true);
    });
    Promise.all([
      workspaceService.getBySlug(workspaceSlug),
      projectService.get(workspaceSlug, projectId),
      intakeService.list(workspaceSlug, projectId, 'pending'),
      intakeService.pendingCount(workspaceSlug, projectId),
      issueService.list(workspaceSlug, projectId, { limit: 500 }),
    ])
      .then(([w, p, pending, count, issues]) => {
        if (cancelled) return;
        setWorkspace(w ?? null);
        setProject(p ?? null);
        setItems(pending);
        setPendingCount(count);
        setActiveIssues(issues ?? []);
      })
      .catch(() => {
        if (!cancelled) {
          setWorkspace(null);
          setProject(null);
          setItems([]);
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [workspaceSlug, projectId]);

  const selectTab = (next: TabKey) => {
    setTab(next);
    setError(null);
    loadItems(next);
  };

  const runAction = async (itemId: string, action: () => Promise<void>) => {
    if (!workspaceSlug || !projectId) return;
    setBusy(itemId);
    setError(null);
    try {
      await action();
      loadItems(tab);
      reloadPending();
      emitIntakeUpdated(projectId);
    } catch {
      setError('That action could not be completed. Please try again.');
    } finally {
      setBusy(null);
    }
  };

  if (loading) return <div className="p-6 text-sm text-(--txt-tertiary)">Loading intake…</div>;
  if (!workspace || !project)
    return <div className="p-6 text-sm text-(--txt-secondary)">Project not found.</div>;

  const baseUrl = `/${workspace.slug}/projects/${project.id}`;
  const identifier = project.identifier ?? project.id.slice(0, 6);
  const canTriage = tab === 'pending' || tab === 'snoozed';
  const activeById = new Map(activeIssues.map((i) => [i.id, i]));

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-(--txt-primary)">Intake</h1>
          <p className="text-sm text-(--txt-tertiary)">
            Triage incoming work: accept it into the project, decline, snooze, or mark a duplicate.
          </p>
        </div>
        <Link
          to={`${baseUrl}/issues`}
          className="rounded-(--radius-md) px-3 py-1.5 text-sm text-(--txt-secondary) hover:bg-(--bg-layer-1-hover)"
        >
          Active issues →
        </Link>
      </div>

      <div className="flex items-center gap-1 border-b border-(--border-subtle)">
        {TABS.map((t) => (
          <button
            key={t.key}
            type="button"
            onClick={() => selectTab(t.key)}
            className={`flex items-center gap-1.5 border-b-2 px-3 py-2 text-sm transition-colors ${
              tab === t.key
                ? 'border-(--brand-default) text-(--txt-primary)'
                : 'border-transparent text-(--txt-secondary) hover:text-(--txt-primary)'
            }`}
          >
            {t.label}
            {t.key === 'pending' && pendingCount > 0 && (
              <span className="rounded-full bg-(--brand-default) px-1.5 text-xs text-white">
                {pendingCount}
              </span>
            )}
          </button>
        ))}
      </div>

      {error && (
        <p className="rounded-md bg-(--bg-danger-subtle) px-3 py-2 text-sm text-(--txt-danger-primary)">
          {error}
        </p>
      )}

      {items.length === 0 ? (
        <div className="rounded-md border border-(--border-subtle) bg-(--bg-surface-1) p-8 text-center text-sm text-(--txt-tertiary)">
          Nothing here. This triage queue is empty.
        </div>
      ) : (
        <div className="overflow-x-auto rounded-md border border-(--border-subtle) bg-(--bg-surface-1)">
          <table className="w-full text-left text-sm">
            <thead className="bg-(--bg-layer-1)">
              <tr>
                <th className="px-4 py-2 font-medium text-(--txt-secondary)">Work item</th>
                <th className="px-4 py-2 font-medium text-(--txt-secondary)">Source</th>
                <th className="px-4 py-2 font-medium text-(--txt-secondary)">Priority</th>
                <th className="px-4 py-2 font-medium text-(--txt-secondary)">
                  {tab === 'snoozed'
                    ? 'Snoozed until'
                    : tab === 'duplicate'
                      ? 'Duplicate of'
                      : 'Created'}
                </th>
                <th className="px-4 py-2 font-medium text-(--txt-secondary)">Actions</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => {
                const issue = item.issue;
                const dupOf = item.duplicate_to_id ? activeById.get(item.duplicate_to_id) : null;
                return (
                  <tr
                    key={item.id}
                    className="border-t border-(--border-subtle) hover:bg-(--bg-layer-1)"
                  >
                    <td className="px-4 py-2">
                      <Link
                        to={`${baseUrl}/issues/${issue.id}`}
                        className="font-medium text-(--txt-primary) hover:text-(--txt-accent-primary)"
                      >
                        {identifier}-{issue.sequence_id} {issue.name}
                      </Link>
                    </td>
                    <td className="px-4 py-2 text-xs text-(--txt-tertiary)">
                      {sourceLabel(item.source)}
                    </td>
                    <td className="px-4 py-2">
                      {issue.priority && issue.priority !== 'none' ? (
                        <Badge variant={priorityVariant[(issue.priority as Priority) ?? 'none']}>
                          {issue.priority}
                        </Badge>
                      ) : (
                        <span className="text-xs text-(--txt-tertiary)">—</span>
                      )}
                    </td>
                    <td className="px-4 py-2 text-xs text-(--txt-tertiary)">
                      {tab === 'snoozed' && item.snoozed_till
                        ? new Date(item.snoozed_till).toLocaleDateString()
                        : tab === 'duplicate'
                          ? dupOf
                            ? `${identifier}-${dupOf.sequence_id}`
                            : '—'
                          : new Date(issue.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-2">
                      {canTriage ? (
                        <div className="flex flex-wrap items-center gap-2">
                          <button
                            type="button"
                            disabled={busy === item.id}
                            onClick={() =>
                              runAction(item.id, () =>
                                intakeService.accept(workspace.slug, project.id, item.id),
                              )
                            }
                            className="rounded-(--radius-md) bg-(--bg-success-subtle) px-2 py-1 text-xs font-medium text-(--txt-success-primary) hover:opacity-90 disabled:opacity-50"
                          >
                            Accept
                          </button>
                          <button
                            type="button"
                            disabled={busy === item.id}
                            onClick={() =>
                              runAction(item.id, () =>
                                intakeService.decline(workspace.slug, project.id, item.id),
                              )
                            }
                            className="rounded-(--radius-md) bg-(--bg-danger-subtle) px-2 py-1 text-xs font-medium text-(--txt-danger-primary) hover:opacity-90 disabled:opacity-50"
                          >
                            Decline
                          </button>
                          <select
                            aria-label="Snooze"
                            disabled={busy === item.id}
                            value=""
                            onChange={(e) => {
                              const days = Number(e.target.value);
                              e.target.value = '';
                              if (!days) return;
                              const till = new Date(Date.now() + days * 86400000).toISOString();
                              void runAction(item.id, () =>
                                intakeService.snooze(workspace.slug, project.id, item.id, till),
                              );
                            }}
                            className="rounded-(--radius-md) border border-(--border-subtle) bg-(--bg-surface-1) px-1.5 py-1 text-xs text-(--txt-secondary)"
                          >
                            <option value="">Snooze…</option>
                            {SNOOZE_PRESETS.map((p) => (
                              <option key={p.days} value={p.days}>
                                {p.label}
                              </option>
                            ))}
                          </select>
                          <select
                            aria-label="Mark duplicate"
                            disabled={busy === item.id}
                            value=""
                            onChange={(e) => {
                              const dupId = e.target.value;
                              e.target.value = '';
                              if (!dupId) return;
                              void runAction(item.id, () =>
                                intakeService.markDuplicate(
                                  workspace.slug,
                                  project.id,
                                  item.id,
                                  dupId,
                                ),
                              );
                            }}
                            className="max-w-[9rem] rounded-(--radius-md) border border-(--border-subtle) bg-(--bg-surface-1) px-1.5 py-1 text-xs text-(--txt-secondary)"
                          >
                            <option value="">Duplicate of…</option>
                            {activeIssues.map((i) => (
                              <option key={i.id} value={i.id}>
                                {identifier}-{i.sequence_id} {i.name}
                              </option>
                            ))}
                          </select>
                        </div>
                      ) : (
                        <span className="text-xs text-(--txt-tertiary)">—</span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
