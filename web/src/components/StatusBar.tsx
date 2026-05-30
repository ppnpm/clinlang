import { Cloud, HardDrive, Check, Circle } from 'lucide-react';
import { toast } from 'sonner';

import { useStore } from '@/lib/store';

// StatusBar — bottom strip showing deployment mode, workspace path,
// the active file path, line count, and save status.
export function StatusBar() {
  const workspace = useStore((s) => s.workspace);
  const open = useStore((s) => s.open);
  const activePath = useStore((s) => s.activePath);
  const saveActive = useStore((s) => s.saveActive);

  const active = activePath ? open[activePath] : null;
  const lineCount = active ? active.content.split('\n').length : 0;

  const ModeIcon =
    workspace?.mode === 'hosted'
      ? Cloud
      : workspace?.mode === 'local'
        ? HardDrive
        : Circle;

  const onManualSave = async () => {
    if (!active) return;
    try {
      await saveActive();
      toast.success(`Saved ${active.path.split('/').pop() ?? active.path}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Save failed');
    }
  };

  return (
    <footer className="flex h-7 items-center gap-3 border-t border-border bg-muted/30 px-3 text-xs text-muted-foreground">
      <div className="flex items-center gap-1.5">
        <ModeIcon className="h-3 w-3" />
        <span className="capitalize">{workspace?.mode ?? 'offline'}</span>
      </div>
      {workspace?.path && (
        <span
          className="hidden truncate sm:inline-block sm:max-w-[40ch]"
          title={workspace.path}
        >
          {workspace.path}
        </span>
      )}
      <div className="flex-1" />
      {active && (
        <>
          <span className="truncate" title={active.path}>
            {active.path}
          </span>
          <span>{lineCount} lines</span>
          {active.dirty ? (
            <button
              onClick={onManualSave}
              className="flex items-center gap-1 hover:text-foreground text-muted-foreground hover:bg-accent px-1.5 py-0.5 rounded border border-border/60 transition-all font-medium text-[11px] focus-visible:ring-1 focus-visible:ring-ring"
              title="Click to save changes manually"
            >
              <Circle className="h-1.5 w-1.5 fill-current animate-pulse text-muted-foreground" />
              Save
            </button>
          ) : (
            <span className="flex items-center gap-1 text-emerald-600 dark:text-emerald-400 font-medium">
              <Check className="h-3.5 w-3.5" />
              Saved
            </span>
          )}
        </>
      )}
    </footer>
  );
}
