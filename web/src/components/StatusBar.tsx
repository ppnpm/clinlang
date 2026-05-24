import { Cloud, HardDrive, Check, Circle } from 'lucide-react';

import { useStore } from '@/lib/store';

// StatusBar — bottom strip showing deployment mode, workspace path,
// the active file path, line count, and save status.
export function StatusBar() {
  const workspace = useStore((s) => s.workspace);
  const open = useStore((s) => s.open);
  const activePath = useStore((s) => s.activePath);

  const active = activePath ? open[activePath] : null;
  const lineCount = active ? active.content.split('\n').length : 0;

  const ModeIcon =
    workspace?.mode === 'hosted'
      ? Cloud
      : workspace?.mode === 'local'
        ? HardDrive
        : Circle;

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
            <span className="flex items-center gap-1">
              <Circle className="h-2 w-2 fill-current" />
              Unsaved
            </span>
          ) : (
            <span className="flex items-center gap-1">
              <Check className="h-3 w-3" />
              Saved
            </span>
          )}
        </>
      )}
    </footer>
  );
}
