import { useEffect, useRef, useState } from 'react';
import { Modal, Button } from '../ui';
import { importerService } from '../../services/importerService';
import type { ImporterApiResponse } from '../../api/types';

interface ImportCSVModalProps {
  open: boolean;
  onClose: () => void;
  workspaceSlug: string;
  projectId: string;
  /** Called once an import finishes so the caller can refresh the issue list. */
  onImported?: () => void;
}

const isTerminal = (s: ImporterApiResponse['status']) =>
  s === 'completed' || s === 'completed_with_errors' || s === 'failed';

/**
 * Upload a CSV of work items and watch the import progress (issue #207). The
 * CSV needs a name/title/summary column; description, priority, and state/status
 * columns are mapped when present. Progress is polled until the job finishes.
 */
export function ImportCSVModal({
  open,
  onClose,
  workspaceSlug,
  projectId,
  onImported,
}: ImportCSVModalProps) {
  const [file, setFile] = useState<File | null>(null);
  const [job, setJob] = useState<ImporterApiResponse | null>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const notifiedRef = useRef(false);

  // Reset when the modal is (re)opened.
  useEffect(() => {
    if (open) {
      setFile(null);
      setJob(null);
      setUploading(false);
      setError(null);
      notifiedRef.current = false;
    }
  }, [open]);

  // Poll until the job reaches a terminal state.
  useEffect(() => {
    if (!job || isTerminal(job.status)) {
      if (job && isTerminal(job.status) && !notifiedRef.current) {
        notifiedRef.current = true;
        onImported?.();
      }
      return;
    }
    let cancelled = false;
    const timer = setTimeout(async () => {
      try {
        const next = await importerService.get(workspaceSlug, projectId, job.id);
        if (!cancelled) setJob(next);
      } catch {
        // Keep the last known state; a transient poll error shouldn't crash the UI.
      }
    }, 1000);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [job, workspaceSlug, projectId, onImported]);

  const handleUpload = async () => {
    if (!file) return;
    setUploading(true);
    setError(null);
    try {
      const created = await importerService.createCSV(workspaceSlug, projectId, file);
      setJob(created);
    } catch (err: unknown) {
      const res = (err as { response?: { status?: number; data?: { error?: string } } })?.response;
      setError(
        res?.data?.error ||
          (res?.status === 403
            ? 'You do not have access to import into this project.'
            : 'Could not start the import. Check the file and try again.'),
      );
    } finally {
      setUploading(false);
    }
  };

  const percent =
    job && job.total_count > 0
      ? Math.round(((job.processed_count + job.error_count) / job.total_count) * 100)
      : 0;

  return (
    <Modal open={open} onClose={onClose} title="Import work items from CSV">
      {job ? (
        <div className="space-y-4">
          <div>
            <div className="mb-1 flex items-center justify-between text-sm">
              <span className="font-medium text-(--txt-primary)">
                {job.status === 'processing' || job.status === 'queued'
                  ? 'Importing…'
                  : job.status === 'failed'
                    ? 'Import failed'
                    : job.error_count > 0
                      ? 'Imported with some errors'
                      : 'Import complete'}
              </span>
              <span className="text-(--txt-tertiary)">{percent}%</span>
            </div>
            <div className="h-2 w-full overflow-hidden rounded-full bg-(--bg-layer-2)">
              <div
                className="h-full rounded-full bg-(--bg-accent-primary) transition-all"
                style={{ width: `${percent}%` }}
              />
            </div>
          </div>
          <dl className="grid grid-cols-3 gap-2 text-center text-sm">
            <div className="rounded-(--radius-md) border border-(--border-subtle) py-2">
              <dt className="text-xs text-(--txt-tertiary)">Total</dt>
              <dd className="font-medium text-(--txt-primary)">{job.total_count}</dd>
            </div>
            <div className="rounded-(--radius-md) border border-(--border-subtle) py-2">
              <dt className="text-xs text-(--txt-tertiary)">Created</dt>
              <dd className="font-medium text-(--txt-primary)">{job.processed_count}</dd>
            </div>
            <div className="rounded-(--radius-md) border border-(--border-subtle) py-2">
              <dt className="text-xs text-(--txt-tertiary)">Errors</dt>
              <dd className="font-medium text-(--txt-primary)">{job.error_count}</dd>
            </div>
          </dl>
          {job.error_message && (
            <p className="text-xs text-(--txt-danger-primary)">{job.error_message}</p>
          )}
          {isTerminal(job.status) && (
            <div className="flex justify-end">
              <Button onClick={onClose}>Done</Button>
            </div>
          )}
        </div>
      ) : (
        <div className="space-y-4">
          <p className="text-sm text-(--txt-secondary)">
            Upload a CSV with a <code className="text-xs">name</code> (or{' '}
            <code className="text-xs">title</code>) column. Optional{' '}
            <code className="text-xs">description</code>, <code className="text-xs">priority</code>,
            and <code className="text-xs">state</code> columns are mapped when present.
          </p>
          <input
            type="file"
            accept=".csv,text/csv"
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            className="block w-full text-sm text-(--txt-secondary) file:mr-3 file:rounded-(--radius-md) file:border file:border-(--border-subtle) file:bg-(--bg-surface-1) file:px-3 file:py-1.5 file:text-sm file:text-(--txt-primary) hover:file:bg-(--bg-layer-1-hover)"
          />
          {error && <p className="text-sm text-(--txt-danger-primary)">{error}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button disabled={!file || uploading} onClick={handleUpload}>
              {uploading ? 'Uploading…' : 'Start import'}
            </Button>
          </div>
        </div>
      )}
    </Modal>
  );
}
