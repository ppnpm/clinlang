import { useEffect, useState } from 'react';
import { Cloud, HardDrive } from 'lucide-react';
import { api } from '@/lib/api';
import { cn } from '@/lib/utils';

// ModeBadge fetches /health once and renders a small chip showing
// the current deployment mode. Purely informational.
export function ModeBadge() {
  const [mode, setMode] = useState<string | null>(null);

  useEffect(() => {
    let alive = true;
    api
      .health()
      .then((h) => {
        if (alive) setMode(h.mode);
      })
      .catch(() => {
        if (alive) setMode('offline');
      });
    return () => {
      alive = false;
    };
  }, []);

  if (!mode) return null;

  const Icon = mode === 'hosted' ? Cloud : HardDrive;
  const label =
    mode === 'hosted' ? 'Hosted' : mode === 'local' ? 'Local' : 'Offline';

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border border-border bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground'
      )}
      aria-label={`Deployment mode: ${label}`}
    >
      <Icon className="h-3.5 w-3.5" />
      {label}
    </span>
  );
}
