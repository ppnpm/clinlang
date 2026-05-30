import { lazy, Suspense, useCallback, useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';
import { Copy, Download, Check } from 'lucide-react';

import { TabBar } from '@/components/TabBar';
import { Button } from '@/components/ui/button';
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
  const autosaveOn = useStore((s) => s.autosaveOn);
  const fontSize = useStore((s) => s.fontSize);
  const editorFont = useStore((s) => s.editorFont);
  const lineSpacing = useStore((s) => s.lineSpacing);

  const active = activePath ? open[activePath] : null;

  const [soap, setSoap] = useState<string>('');
  const [markers, setMarkers] = useState<RangeMarker[]>([]);
  const [images, setImages] = useState<string[]>([]);
  const [copied, setCopied] = useState(false);
  const soapTimer = useRef<number | null>(null);
  const autosaveTimer = useRef<number | null>(null);

  const onCopy = async () => {
    if (!soap) return;
    try {
      await navigator.clipboard.writeText(soap);
      setCopied(true);
      toast.success('SOAP note copied to clipboard');
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      toast.error('Failed to copy to clipboard');
    }
  };

  const onExport = (format: 'md' | 'txt') => {
    if (!soap) return;
    const blob = new Blob([soap], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    const baseName = activePath ? activePath.split('/').pop()?.replace(/\.[^/.]+$/, "") : 'soap-note';
    a.download = `${baseName}-soap.${format}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    toast.success(`Exported as ${format.toUpperCase()}`);
  };

  // Debounced SOAP preview from the server.
  useEffect(() => {
    if (soapTimer.current) window.clearTimeout(soapTimer.current);
    if (!active || !active.content.trim()) {
      setSoap('');
      setMarkers([]);
      setImages([]);
      return;
    }
    soapTimer.current = window.setTimeout(() => {
      api
        .soap(active.content, markersOn)
        .then((res) => {
          setSoap(res.soap);
          setMarkers(res.range_markers ?? []);
          setImages(res.images ?? []);
        })
        .catch((err) => {
          setSoap(`Error: ${err.message ?? err}`);
          setMarkers([]);
          setImages([]);
        });
    }, 250);
    return () => {
      if (soapTimer.current) window.clearTimeout(soapTimer.current);
    };
  }, [active, markersOn]);

  // Debounced Autosave.
  useEffect(() => {
    if (autosaveTimer.current) window.clearTimeout(autosaveTimer.current);
    if (!autosaveOn || !active || !active.dirty) return;

    autosaveTimer.current = window.setTimeout(async () => {
      try {
        await saveActive();
      } catch (err) {
        console.error('Autosave failed:', err);
      }
    }, 1500);

    return () => {
      if (autosaveTimer.current) window.clearTimeout(autosaveTimer.current);
    };
  }, [active?.content, active?.dirty, autosaveOn, saveActive]);

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

  const fontSizes = {
    sm: '12px',
    md: '14px',
    lg: '16px',
    xl: '18px',
  };

  const fontFamilies = {
    mono: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
    atkinson: '"Atkinson Hyperlegible", ui-sans-serif, system-ui, -apple-system, sans-serif',
    dyslexic: '"OpenDyslexic", "OpenDyslexicRegular", "Dyslexic", "Comic Sans MS", sans-serif',
    sans: '"Inter", ui-sans-serif, system-ui, -apple-system, sans-serif',
  };

  const lineSpacings = {
    normal: '1.45',
    relaxed: '1.75',
    double: '2.25',
  };

  const dynamicStyles = {
    '--editor-font-size': fontSizes[fontSize],
    '--preview-font-size': fontSizes[fontSize],
    '--editor-font-family': fontFamilies[editorFont],
    '--editor-line-spacing': lineSpacings[lineSpacing],
    '--preview-line-spacing': lineSpacings[lineSpacing],
  } as React.CSSProperties;

  return (
    <div
      className="flex h-full flex-col overflow-hidden"
      style={dynamicStyles}
    >
      <TabBar />

      <div className="flex flex-1 overflow-hidden relative">
        {/* Editor pane */}
        <section
          className={cn(
            'flex flex-1 flex-col overflow-hidden',
            previewOpen && 'hidden md:flex'
          )}
        >
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
            'shrink-0 overflow-hidden bg-muted/10 transition-all duration-150',
            previewOpen
              ? 'w-full md:w-[40%] md:min-w-[280px] md:border-l md:border-border'
              : 'w-0'
          )}
        >
          <div className="flex h-full flex-col">
            <div className="flex items-center justify-between border-b border-border px-3 py-1">
              <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                SOAP preview
              </span>
              {soap && (
                <div className="flex items-center gap-0.5">
                  <Button
                    size="icon"
                    variant="ghost"
                    className="h-7 w-7 text-muted-foreground hover:text-foreground"
                    onClick={onCopy}
                    aria-label="Copy SOAP note"
                    title="Copy SOAP note"
                  >
                    {copied ? (
                      <Check className="h-3.5 w-3.5 text-emerald-600 dark:text-emerald-400" />
                    ) : (
                      <Copy className="h-3.5 w-3.5" />
                    )}
                  </Button>
                  <Button
                    size="icon"
                    variant="ghost"
                    className="h-7 w-7 text-muted-foreground hover:text-foreground"
                    onClick={() => onExport('md')}
                    aria-label="Export as Markdown"
                    title="Export as Markdown"
                  >
                    <Download className="h-3.5 w-3.5" />
                  </Button>
                </div>
              )}
            </div>

            {markersOn && markers.length > 0 && (
              <div className="flex flex-wrap gap-1.5 border-b border-border px-3 py-2">
                {markers.map((m, i) => (
                  <RangeMarkerChip key={i} marker={m} />
                ))}
              </div>
            )}

            <pre
              className="flex-1 overflow-auto whitespace-pre-wrap px-4 py-3"
              style={{
                fontSize: 'var(--preview-font-size)',
                fontFamily: 'var(--editor-font-family)',
                lineHeight: 'var(--preview-line-spacing)',
              }}
            >
              {soap || (
                <span className="text-muted-foreground">
                  {active
                    ? 'Empty file. Start typing to see the preview.'
                    : 'No file open.'}
                </span>
              )}
            </pre>

            {images.length > 0 && (
              <div className="border-t border-border bg-muted/20 p-3">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground block mb-2">
                  Attachments ({images.length})
                </span>
                <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-thin">
                  {images.map((imgUrl, i) => {
                    const encodedUrl = `/api/v1/files/${imgUrl.split('/').map(encodeURIComponent).join('/')}?raw=true`;
                    const filename = imgUrl.split('/').pop() || imgUrl;
                    return (
                      <a
                        key={i}
                        href={encodedUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="relative group shrink-0 block border border-border rounded overflow-hidden bg-background hover:border-muted-foreground transition-colors"
                        title={filename}
                      >
                        <img
                          src={encodedUrl}
                          alt={filename}
                          className="h-16 w-20 object-cover"
                        />
                        <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity">
                          <span className="text-[10px] text-white font-medium">View</span>
                        </div>
                      </a>
                    );
                  })}
                </div>
              </div>
            )}
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
