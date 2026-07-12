import { useState } from 'react';
import { Button, Modal } from '../../ui';
import { stateService } from '../../../services/stateService';
import type { StateApiResponse } from '../../../api/types';

interface ProjectStateModalProps {
  open: boolean;
  onClose: () => void;
  workspaceSlug: string | undefined;
  selectedProjectId: string | null;
  projectStateEdit: StateApiResponse | null;
  setProjectStateEdit: (state: StateApiResponse | null) => void;
  projectStateName: string;
  setProjectStateName: (value: string) => void;
  projectStateColor: string;
  setProjectStateColor: (value: string) => void;
  projectStateGroup: string;
  setProjectStateGroup: (value: string) => void;
  setProjectStates: (states: StateApiResponse[]) => void;
  setProjectStateModalOpen: (value: boolean) => void;
}

/** Add/edit modal for a project workflow state. */
export function ProjectStateModal({
  open,
  onClose,
  workspaceSlug,
  selectedProjectId,
  projectStateEdit,
  setProjectStateEdit,
  projectStateName,
  setProjectStateName,
  projectStateColor,
  setProjectStateColor,
  projectStateGroup,
  setProjectStateGroup,
  setProjectStates,
  setProjectStateModalOpen,
}: ProjectStateModalProps) {
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={projectStateEdit ? 'Edit state' : 'Add state'}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button
            disabled={!projectStateName.trim() || saving}
            onClick={async () => {
              if (!workspaceSlug || !selectedProjectId || !projectStateName.trim()) return;
              setSaving(true);
              setError(null);
              try {
                if (projectStateEdit) {
                  await stateService.update(workspaceSlug, selectedProjectId, projectStateEdit.id, {
                    name: projectStateName.trim(),
                    color: projectStateColor,
                    group: projectStateGroup,
                  });
                } else {
                  await stateService.create(workspaceSlug, selectedProjectId, {
                    name: projectStateName.trim(),
                    color: projectStateColor,
                    group: projectStateGroup,
                  });
                }
                const list = await stateService.list(workspaceSlug, selectedProjectId);
                setProjectStates(list ?? []);
                setProjectStateModalOpen(false);
                setProjectStateEdit(null);
              } catch {
                setError('Failed to save state. Please try again.');
              } finally {
                setSaving(false);
              }
            }}
          >
            {saving ? 'Saving…' : projectStateEdit ? 'Save' : 'Create'}
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        {error ? (
          <p className="rounded-(--radius-md) bg-(--bg-danger-subtle) px-3 py-2 text-sm text-(--txt-danger)">
            {error}
          </p>
        ) : null}
        <div>
          <label className="mb-1 block text-sm font-medium text-(--txt-secondary)">Name</label>
          <input
            type="text"
            value={projectStateName}
            onChange={(e) => setProjectStateName(e.target.value)}
            className="w-full rounded-(--radius-md) border border-(--border-subtle) bg-(--bg-surface-1) px-3 py-2 text-sm text-(--txt-primary) focus:outline-none focus:border-(--border-strong)"
            placeholder="e.g. In Progress"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-(--txt-secondary)">Color</label>
          <div className="flex items-center gap-2">
            <input
              type="color"
              value={projectStateColor}
              onChange={(e) => setProjectStateColor(e.target.value)}
              className="h-9 w-14 cursor-pointer rounded border border-(--border-subtle)"
            />
            <input
              type="text"
              value={projectStateColor}
              onChange={(e) => setProjectStateColor(e.target.value)}
              className="flex-1 rounded-(--radius-md) border border-(--border-subtle) bg-(--bg-surface-1) px-3 py-2 text-sm text-(--txt-primary) focus:outline-none focus:border-(--border-strong)"
            />
          </div>
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-(--txt-secondary)">Group</label>
          <select
            value={projectStateGroup}
            onChange={(e) => setProjectStateGroup(e.target.value)}
            className="w-full rounded-(--radius-md) border border-(--border-subtle) bg-(--bg-surface-1) px-3 py-2 text-sm text-(--txt-primary) focus:outline-none focus:border-(--border-strong)"
          >
            <option value="backlog">Backlog</option>
            <option value="unstarted">Unstarted</option>
            <option value="started">Started</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>
      </div>
    </Modal>
  );
}
