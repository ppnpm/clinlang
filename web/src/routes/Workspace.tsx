import { lazy, Suspense, useCallback, useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';

import { TabBar } from '@/components/TabBar';
import { useStore } from '@/lib/store';
import { api } from '@/lib/api';
import { cn } from '@/lib/utils';
import type { RangeMarker } from '@/lib/types';

// Editor is heavy (CodeMirror + extensions + autocomplete + language).
// Lazy-loading lets the welcome dialog, file tree, and settings drawer
// paint without waiting for ~150 KB gzipped of editor code. The editor
// chunk loads only when an actual file is opened.
const Editor = lazy(() =>
  import('@/components/Editor').then((m) => ({ default: m.Editor }))
);

function EditorSkeleton() {
  return (
    <div className="flex flex-1 items-center justify-center text-xs text-muted-foreground">
      Loading editor…
    </div>
  );
}

const EMPTY_PLACEHOLDER = `Open a file from the sidebar, or click + to create a new one.

As you type:
  - the preview pane on the right shows the SOAP rendering live.
  - Ctrl+S saves to the current file. No new files are created.

Try right-clicking any file in the sidebar for Rename / Duplicate / Delete,
or drag a file into a folder to move it.`;

export function Workspace() {
  const open = useStore((s) => s.open);
  const activePath = useStore((s) => s.activePath);
  const updateActiveContent = useStore((s) => s.updateActiveContent);
  const saveActive = useStore((s) => s.saveActive);
  const previewOpen = useStore((s) => s.previewOpen);
  const markersOn = useStore((s) => s.markersOn);

  const active = activePath ? open[activePath] : null;

  const [soap, setSoap] = useState<string>('');
  const [markers, setMarkers] = useState<RangeMarker[]>([]);
  const soapTimer = useRef<number | null>(null);

  // Debounced SOAP preview from the server.
  useEffect(() => {
    if (soapTimer.current) window.clearTimeout(soapTimer.current);
    if (!active || !active.content.trim()) {
      setSoap('');
      setMarkers([]);
      return;
    }
    soapTimer.current = window.setTimeout(() => {
      api
        .soap(active.content, markersOn)
        .then((res) => {
          setSoap(res.soap);
          setMarkers(res.range_markers ?? []);
        })
        .catch((err) => {
          setSoap(`Error: ${err.message ?? err}`);
          setMarkers([]);
        });
    }, 250);
    return () => {
      if (soapTimer.current) window.clearTimeout(soapTimer.current);
    };
  }, [active, markersOn]);

  // Ctrl+S / Cmd+S → save the active file.
  const onSave = useCallback(async () => {
    if (!active) {
      toast.warning('No file is open. Create or open one first.');
      return;
    }
    try {
      await saveActive();
      toast.success(`Saved ${active.path}`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Save failed';
      toast.error(msg);
    }
  }, [active, saveActive]);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
        e.preventDefault();
        void onSave();
      }
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onSave]);

  return (
    <div className="flex h-full flex-col overflow-hidden">
      <TabBar />

      <div className="flex flex-1 overflow-hidden">
        {/* Editor pane */}
        <section className="flex flex-1 flex-col overflow-hidden">
          {active ? (
            <Suspense fallback={<EditorSkeleton />}>
              <Editor
                value={active.content}
                onChange={updateActiveContent}
                filePath={active.path}
                rangeMarkers={markers}
              />
            </Suspense>
          ) : (
            <div className="flex flex-1 items-center justify-center px-6 py-10 text-center text-sm text-muted-foreground">
              <pre className="max-w-md whitespace-pre-wrap font-sans leading-relaxed">
                {EMPTY_PLACEHOLDER}
              </pre>
            </div>
          )}
        </section>

        {/* Preview pane */}
        <aside
          className={cn(
            'shrink-0 overflow-hidden border-l border-border bg-muted/10 transition-[width] duration-150',
            previewOpen ? 'w-[40%] min-w-[280px]' : 'w-0'
          )}
        >
          <div className="flex h-full flex-col">
            <div className="flex items-center justify-between border-b border-border px-3 py-1.5">
              <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                SOAP preview
              </span>
            </div>

            {markersOn && markers.length > 0 && (
              <div className="flex flex-wrap gap-1.5 border-b border-border px-3 py-2">
                {markers.map((m, i) => (
                  <RangeMarkerChip key={i} marker={m} />
                ))}
              </div>
            )}

            <pre className="flex-1 overflow-auto whitespace-pre-wrap px-4 py-3 font-mono text-xs leading-relaxed">
              {soap || (
                <span className="text-muted-foreground">
                  {active
                    ? 'Empty file. Start typing to see the preview.'
                    : 'No file open.'}
                </span>
              )}
            </pre>
          </div>
        </aside>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────
// RangeMarkerChip
// ─────────────────────────────────────────────────────────────────

// A neutral, transcription-only annotation chip. Deliberately styled
// using muted/border colors — no red, no orange, no warning icon.
// The medicolegal posture requires that out-of-range markers do NOT
// imply clinical urgency.
function RangeMarkerChip({ marker }: { marker: RangeMarker }) {
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-full border border-border bg-muted/50 px-2 py-0.5 font-mono text-[11px] text-muted-foreground"
      title={`${marker.field} ${marker.value} · outside ref ${marker.reference_range} · ${marker.source}`}
    >
      <span className="font-semibold text-foreground">{marker.field}</span>
      <span className="text-foreground">{marker.value}</span>
      <span className="text-muted-foreground/60">·</span>
      <span>outside {marker.reference_range}</span>
    </span>
  );
}
