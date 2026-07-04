import { Button, Modal } from '../../ui';
import { IconChevronDown } from '../icons';
import { issueService } from '../../../services/issueService';
import type { ProjectApiResponse, WorkspaceApiResponse } from '../../../api/types';

interface ExportModalProps {
  open: boolean;
  onClose: () => void;
  workspaceSlug: string | undefined;
  workspace: WorkspaceApiResponse | null;
  projects: ProjectApiResponse[];
  exportFormat: string;
  exportProjectValue: string;
  setExportProjectValue: (value: string) => void;
  exporting: boolean;
  setExporting: (value: boolean) => void;
  setExportProjectOpen: (value: boolean) => void;
}

/** Export-issues modal: choose a project (or all) and download a CSV/JSON snapshot. */
export function ExportModal({
  open,
  onClose,
  workspaceSlug,
  workspace,
  projects,
  exportFormat,
  exportProjectValue,
  setExportProjectValue,
  exporting,
  setExporting,
  setExportProjectOpen,
}: ExportModalProps) {
  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`Export ${exportFormat.toUpperCase()}`}
      footer={
        <>
          <Button
            variant="secondary"
            onClick={() => setExportProjectOpen(false)}
            disabled={exporting}
          >
            Cancel
          </Button>
          <Button
            disabled={exporting}
            onClick={async () => {
              if (!workspaceSlug) return;
              setExporting(true);
              try {
                const projectIds =
                  exportProjectValue === 'all' ? projects.map((p) => p.id) : [exportProjectValue];
                const allIssues: Array<
                  Record<string, unknown> & {
                    project_id?: string;
                    project_name?: string;
                  }
                > = [];
                const limit = 2000;
                for (const pid of projectIds) {
                  const proj = projects.find((p) => p.id === pid);
                  let offset = 0;
                  while (true) {
                    const issues = await issueService.list(workspaceSlug, pid, { limit, offset });
                    if (!issues.length) break;
                    for (const i of issues) {
                      allIssues.push({
                        ...i,
                        project_id: pid,
                        project_name: proj?.name,
                      });
                    }
                    if (issues.length < limit) break;
                    offset += issues.length;
                  }
                }
                const fmt = exportFormat === 'xlsx' ? 'csv' : exportFormat;
                let blob: Blob;
                let filename: string;
                const base = `export-${workspace?.slug ?? workspaceSlug}-${new Date().toISOString().slice(0, 10)}`;
                if (fmt === 'json') {
                  const str = JSON.stringify(allIssues, null, 2);
                  blob = new Blob([str], { type: 'application/json' });
                  filename = `${base}.json`;
                } else {
                  const headers = [
                    'id',
                    'project_id',
                    'project_name',
                    'name',
                    'priority',
                    'state_id',
                    'created_at',
                    'updated_at',
                    'description',
                  ];
                  const rows = allIssues.map((row) =>
                    headers
                      .map((h) => {
                        const v = row[h];
                        if (v === null || v === undefined) return '';
                        let s: string;
                        if (typeof v === 'object' && h === 'description') {
                          s = (row.description_html as string) ?? JSON.stringify(v);
                        } else {
                          s = String(v);
                        }
                        return s.includes(',') || s.includes('"') || s.includes('\n')
                          ? `"${s.replace(/"/g, '""')}"`
                          : s;
                      })
                      .join(','),
                  );
                  const csv = [headers.join(','), ...rows].join('\r\n');
                  blob = new Blob([csv], { type: 'text/csv' });
                  filename = `${base}.csv`;
                }
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = filename;
                a.click();
                setTimeout(() => URL.revokeObjectURL(url), 0);
                setExportProjectOpen(false);
              } catch {
                // could set export error state
              } finally {
                setExporting(false);
              }
            }}
          >
            {exporting ? 'Exporting…' : 'Export'}
          </Button>
        </>
      }
    >
      <div>
        <label className="mb-1 block text-sm font-medium text-(--txt-secondary)">Projects</label>
        <div className="relative">
          <select
            value={exportProjectValue}
            onChange={(e) => setExportProjectValue(e.target.value)}
            className="w-full appearance-none rounded-(--radius-md) border border-(--border-subtle) bg-(--bg-surface-1) px-3 py-2 pr-8 text-sm text-(--txt-primary) focus:outline-none focus:border-(--border-strong)"
          >
            <option value="all">All projects</option>
            {projects.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
          <span className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 text-(--txt-icon-tertiary)">
            <IconChevronDown />
          </span>
        </div>
        {exportFormat === 'xlsx' && (
          <p className="mt-2 text-sm text-(--txt-tertiary)">
            Excel export will download as CSV for now.
          </p>
        )}
      </div>
    </Modal>
  );
}
