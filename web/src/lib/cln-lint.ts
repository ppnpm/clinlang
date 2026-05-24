import { linter, lintGutter, type Diagnostic } from '@codemirror/lint';
import type { EditorView } from '@codemirror/view';

import { api } from './api';

// clnLinter — CodeMirror extension that calls /api/v1/lint on debounced
// input and surfaces parser warnings as gutter dots + inline messages.
//
// All diagnostics use severity "info" (neutral). Parser warnings are
// transcription/syntax feedback, NOT clinical urgency signals — the
// medicolegal posture requires that the editor never render alarming
// styling for clinical-looking text.
//
// Range markers are NOT shown here; they get hover tooltips via a
// separate extension (see cln-range-tooltips.ts). Inline diagnostic
// gutter dots are reserved for "you typed something the parser can't
// handle" feedback.

// Warning text from the engine often looks like:
//   "Line 3: command 'foo' has no arguments"
//   "Patient sex not specified"
//   "Unrecognized vitals token: xyz"
// We try to extract a line number from the "Line N:" prefix; failing
// that, we attach the diagnostic to line 1 so the user sees something.
const LINE_PREFIX = /^Line (\d+):\s*/;

function diagnosticsFromWarnings(
  view: EditorView,
  warnings: string[]
): Diagnostic[] {
  const out: Diagnostic[] = [];
  for (const raw of warnings) {
    const m = raw.match(LINE_PREFIX);
    const lineNo = m ? Math.max(1, parseInt(m[1], 10)) : 1;
    const message = m ? raw.slice(m[0].length) : raw;

    const totalLines = view.state.doc.lines;
    const safeLine = Math.min(lineNo, totalLines);
    const line = view.state.doc.line(safeLine);

    out.push({
      from: line.from,
      to: line.to,
      severity: 'info',
      message,
      source: 'clinlang',
    });
  }
  return out;
}

export function clnLinterExtension(filePath?: string | null) {
  // The filePath is unused for now — we lint the in-memory document
  // content. Including it in the signature lets callers re-instantiate
  // the linter when the file changes, which is what we want.
  void filePath;
  return [
    linter(
      async (view: EditorView) => {
        const text = view.state.doc.toString();
        if (!text.trim()) return [];
        try {
          const res = await api.lint(text);
          return diagnosticsFromWarnings(view, res.warnings ?? []);
        } catch {
          // Network/transport error — fail silently. Linter failures
          // shouldn't block the editor.
          return [];
        }
      },
      {
        // Default delay is 750ms; we tighten to 500ms to feel responsive
        // without flooding the API as the user types.
        delay: 500,
      }
    ),
    lintGutter(),
  ];
}
