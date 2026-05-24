import { useState } from 'react';
import { Folder, FolderSearch } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { api, ApiError } from '@/lib/api';
import { cn } from '@/lib/utils';

export interface WorkspacePickerProps {
  value: string;
  onChange: (next: string) => void;
  // Hide the "Type manually" toggle (used in settings where we may
  // want a different fallback). Defaults to false.
  hideManualToggle?: boolean;
}

// WorkspacePicker — read-only path display + Browse button. Used in
// both the welcome carousel and the Settings drawer for consistency.
//
// Falls back gracefully when no native folder picker is available on
// the host OS (e.g. headless Linux): shows an inline notice and the
// user can switch to manual entry via the toggle.
export function WorkspacePicker({
  value,
  onChange,
  hideManualToggle,
}: WorkspacePickerProps) {
  const [busy, setBusy] = useState(false);
  const [manualMode, setManualMode] = useState(false);

  const onBrowse = async () => {
    setBusy(true);
    try {
      const res = await api.browseWorkspace();
      if (res.cancelled || !res.path) return; // user cancelled — no toast
      onChange(res.path);
    } catch (err) {
      if (err instanceof ApiError && err.status === 501) {
        toast.message(
          'Native folder picker not available on this system — switching to manual entry.'
        );
        setManualMode(true);
      } else {
        toast.error((err as Error).message ?? 'Could not open folder picker');
      }
    } finally {
      setBusy(false);
    }
  };

  if (manualMode) {
    return (
      <div className="grid gap-2">
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder="/absolute/path/to/your/notes"
          spellCheck={false}
          autoFocus
        />
        {!hideManualToggle && (
          <button
            type="button"
            className="text-left text-xs text-muted-foreground underline-offset-2 hover:underline"
            onClick={() => setManualMode(false)}
          >
            Use the folder picker instead
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="grid gap-2">
      <div
        className={cn(
          'flex h-9 items-center gap-2 rounded-md border border-input bg-muted/30 px-3 text-sm',
          value ? 'text-foreground' : 'text-muted-foreground'
        )}
      >
        <Folder className="h-4 w-4 shrink-0 text-muted-foreground" />
        <span className="truncate" title={value || 'No folder chosen'}>
          {value || 'No folder chosen'}
        </span>
      </div>
      <div className="flex items-center justify-between">
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={onBrowse}
          disabled={busy}
          className="gap-2"
        >
          <FolderSearch className="h-4 w-4" />
          {value ? 'Change…' : 'Browse…'}
        </Button>
        {!hideManualToggle && (
          <button
            type="button"
            className="text-xs text-muted-foreground underline-offset-2 hover:underline"
            onClick={() => setManualMode(true)}
          >
            Type manually
          </button>
        )}
      </div>
    </div>
  );
}
