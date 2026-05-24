import { useState } from 'react';
import { toast } from 'sonner';

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { useStore } from '@/lib/store';

// ConflictDialog appears when a save fails with 412 Precondition Failed
// — meaning the file changed on disk (another writer, another tab,
// an external editor) since the user last read it. The user picks:
//
//   • Overwrite — push the local content, blowing away the other change
//   • Reload    — discard local edits and re-read from disk
//   • Cancel    — leave the tab dirty, decide later
export function ConflictDialog() {
  const conflict = useStore((s) => s.conflict);
  const overwrite = useStore((s) => s.resolveConflictOverwrite);
  const reload = useStore((s) => s.resolveConflictReload);
  const dismiss = useStore((s) => s.dismissConflict);

  const [busy, setBusy] = useState(false);

  const onOverwrite = async () => {
    setBusy(true);
    try {
      await overwrite();
      toast.success('Overwritten with your version.');
    } catch (err) {
      toast.error((err as Error).message ?? 'Overwrite failed');
    } finally {
      setBusy(false);
    }
  };

  const onReload = async () => {
    setBusy(true);
    try {
      await reload();
      toast.message('Reloaded from disk. Your edits were discarded.');
    } catch (err) {
      toast.error((err as Error).message ?? 'Reload failed');
    } finally {
      setBusy(false);
    }
  };

  return (
    <AlertDialog
      open={conflict !== null}
      onOpenChange={(open) => {
        if (!open) dismiss();
      }}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>File changed on disk</AlertDialogTitle>
          <AlertDialogDescription>
            <span className="font-mono">{conflict?.path}</span> was modified by
            something else since you opened it. Choose how to resolve:
            <ul className="mt-2 list-disc pl-5 text-xs">
              <li>
                <b>Overwrite</b> — your version wins; the other change is lost.
              </li>
              <li>
                <b>Reload</b> — discard your edits and load what's on disk.
              </li>
              <li>
                <b>Cancel</b> — keep editing locally, decide later.
              </li>
            </ul>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={busy}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={onReload}
            disabled={busy}
            className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
          >
            Reload
          </AlertDialogAction>
          <AlertDialogAction onClick={onOverwrite} disabled={busy}>
            Overwrite
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
