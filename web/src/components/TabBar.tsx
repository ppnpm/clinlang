import { X } from 'lucide-react';

import { cn } from '@/lib/utils';
import { useStore } from '@/lib/store';

// TabBar shows the currently open files. Click to focus, × to close.
// A small dot next to the filename indicates an unsaved edit.
export function TabBar() {
  const open = useStore((s) => s.open);
  const activePath = useStore((s) => s.activePath);
  const setActive = useStore((s) => s.setActive);
  const closeFile = useStore((s) => s.closeFile);

  const tabs = Object.values(open);
  if (tabs.length === 0) return null;

  const fileName = (p: string) => p.split('/').pop() ?? p;

  return (
    <div className="flex items-end overflow-x-auto border-b border-border bg-muted/20">
      {tabs.map((tab) => {
        const isActive = tab.path === activePath;
        return (
          <div
            key={tab.path}
            className={cn(
              'group flex h-9 shrink-0 cursor-pointer items-center gap-2 border-r border-border px-3 text-sm',
              isActive
                ? 'bg-background text-foreground'
                : 'text-muted-foreground hover:bg-background/60'
            )}
            onClick={() => setActive(tab.path)}
            title={tab.path}
          >
            <span className="max-w-[180px] truncate">{fileName(tab.path)}</span>
            {tab.dirty && (
              <span
                className="h-1.5 w-1.5 rounded-full bg-foreground"
                aria-label="Unsaved changes"
              />
            )}
            <button
              className="rounded-sm p-0.5 opacity-0 hover:bg-muted hover:text-foreground group-hover:opacity-100"
              onClick={(e) => {
                e.stopPropagation();
                closeFile(tab.path);
              }}
              aria-label={`Close ${tab.path}`}
            >
              <X className="h-3 w-3" />
            </button>
          </div>
        );
      })}
    </div>
  );
}
