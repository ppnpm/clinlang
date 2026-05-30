import { useEffect, useMemo, useRef } from 'react';
import CodeMirror, { type ReactCodeMirrorRef } from '@uiw/react-codemirror';
import { EditorView } from '@codemirror/view';
import { toast } from 'sonner';

import { useTheme } from '@/components/theme-provider';
import { clnExtensions } from '@/lib/cln-language';
import { clnAutocomplete } from '@/lib/cln-autocomplete';
import { clnLinterExtension } from '@/lib/cln-lint';
import {
  rangeTooltipExtensions,
  setRangeMarkers,
} from '@/lib/cln-range-tooltips';
import type { RangeMarker } from '@/lib/types';
import { api } from '@/lib/api';

export interface EditorProps {
  value: string;
  onChange: (next: string) => void;
  // Path of the active file — used by autocomplete cache busting and
  // to re-instantiate the linter when the file changes.
  filePath?: string | null;
  // Latest range markers from the server. Optional; when omitted, no
  // hover tooltips are shown.
  rangeMarkers?: RangeMarker[];
}

// Editor — the .cln editor surface. CodeMirror 6 with our custom
// language pack, autocomplete, lint diagnostics, range-marker hover
// tooltips, and theme-aware appearance.
//
// Visual choices:
//   - No line numbers (cleaner notebook feel, like Obsidian)
//   - Word wrap on (clinical text often runs long)
//   - Monospace font from the system stack
//   - Diagnostic + tooltip styling is neutral (no urgency colors) to
//     preserve the medicolegal posture
export function Editor({ value, onChange, filePath, rangeMarkers }: EditorProps) {
  const { theme } = useTheme();
  const isDark =
    theme === 'dark' ||
    (theme === 'system' &&
      typeof window !== 'undefined' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches);

  const cmRef = useRef<ReactCodeMirrorRef>(null);

  // Push the latest range markers into the editor's StateField so the
  // hover-tooltip extension can read them at hover time.
  useEffect(() => {
    const view = cmRef.current?.view;
    if (!view) return;
    view.dispatch({ effects: setRangeMarkers.of(rangeMarkers ?? []) });
  }, [rangeMarkers]);

  const extensions = useMemo(
    () => [
      EditorView.lineWrapping,
      EditorView.theme({
        '&': { height: '100%', fontSize: 'var(--editor-font-size, 14px)' },
        '.cm-content': {
          padding: '12px 16px',
          fontFamily: 'var(--editor-font-family, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace)',
          caretColor: 'hsl(var(--foreground))',
          lineHeight: 'var(--editor-line-spacing, 1.45)',
        },
        '.cm-line': {
          lineHeight: 'var(--editor-line-spacing, 1.45)',
        },
        '.cm-focused': { outline: 'none' },
        '.cm-gutters': {
          backgroundColor: 'transparent',
          border: 'none',
        },
        '.cm-activeLine': { backgroundColor: 'hsl(var(--accent) / 0.25)' },
        '.cm-activeLineGutter': { backgroundColor: 'transparent' },
        // Neutral lint diagnostic styling — info-level only, no red,
        // no urgency. Matches the medicolegal posture.
        '.cm-lint-marker-info': {
          color: 'hsl(var(--muted-foreground))',
        },
        '.cm-tooltip.cm-tooltip-lint': {
          fontSize: '12px',
          border: '1px solid hsl(var(--border))',
          backgroundColor: 'hsl(var(--popover))',
          color: 'hsl(var(--popover-foreground))',
        },
      }),
      ...clnExtensions,
      clnAutocomplete,
      ...rangeTooltipExtensions,
      ...clnLinterExtension(filePath),
    ],
    [filePath]
  );

  const handleDragOver = (e: React.DragEvent) => {
    if (e.dataTransfer.types.includes('Files')) {
      e.preventDefault();
    }
  };

  const handleDrop = async (e: React.DragEvent) => {
    const files = Array.from(e.dataTransfer.files);
    const imageFiles = files.filter(f => f.type.startsWith('image/'));
    if (imageFiles.length === 0) return;

    e.preventDefault();

    const toastId = toast.loading('Uploading image...');
    try {
      for (const file of imageFiles) {
        const arrayBuffer = await file.arrayBuffer();
        const data = new Uint8Array(arrayBuffer);
        const filename = file.name.replace(/\s+/g, '_');
        const timestamp = Date.now();
        const path = `images/${timestamp}-${filename}`;

        await api.writeBinaryFile(path, data);

        const view = cmRef.current?.view;
        if (view) {
          const coords = view.posAtCoords({ x: e.clientX, y: e.clientY });
          const pos = coords !== null ? coords : view.state.selection.main.from;
          const insertText = `\nimg ${path}\n`;
          view.dispatch({
            changes: { from: pos, to: pos, insert: insertText },
            selection: { anchor: pos + insertText.length }
          });
        }
      }
      toast.success('Uploaded successfully', { id: toastId });
    } catch (err) {
      console.error('Failed to upload image:', err);
      toast.error(err instanceof Error ? err.message : 'Failed to upload image', { id: toastId });
    }
  };

  return (
    <div
      onDragOver={handleDragOver}
      onDrop={handleDrop}
      className="h-full"
    >
      <CodeMirror
        ref={cmRef}
        value={value}
        onChange={onChange}
        theme={isDark ? 'dark' : 'light'}
        extensions={extensions}
        basicSetup={{
          lineNumbers: false,
          foldGutter: false,
          highlightActiveLine: true,
          highlightActiveLineGutter: false,
          bracketMatching: true,
          autocompletion: true,
          searchKeymap: true,
          indentOnInput: false,
          defaultKeymap: true,
          history: true,
        }}
        className="h-full"
      />
    </div>
  );
}
