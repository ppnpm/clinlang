import { hoverTooltip, type Tooltip } from '@codemirror/view';
import { StateField, StateEffect } from '@codemirror/state';

import type { RangeMarker } from './types';

// cln-range-tooltips — when the user hovers over a token that the
// engine flagged as out-of-range (HR 150, BP 160/100, etc.), show a
// small neutral tooltip with the reference range and source.
//
// Markers arrive asynchronously after every parse (see Workspace.tsx
// where we call /api/v1/soap). The Workspace dispatches them into a
// StateField via setRangeMarkers; the hoverTooltip extension reads
// the StateField at hover time.
//
// Token matching is intentionally loose: we look for `<field-keyword><value>`
// substrings on each line, where field-keyword is e.g. "bp", "hr",
// "spo2", "temp", "rr", "hb", "wbc", "creatinine", etc.

// ─────────────────────────────────────────────────────────────────
// StateField holding the latest markers
// ─────────────────────────────────────────────────────────────────

export const setRangeMarkers = StateEffect.define<RangeMarker[]>();

export const rangeMarkersField = StateField.define<RangeMarker[]>({
  create: () => [],
  update(value, tr) {
    for (const e of tr.effects) {
      if (e.is(setRangeMarkers)) return e.value;
    }
    return value;
  },
});

// ─────────────────────────────────────────────────────────────────
// Token → field-keyword map for source scanning
// ─────────────────────────────────────────────────────────────────

// Maps the human-readable Field (as returned by the engine) to the
// source-text prefix the parser recognises. Used to find the relevant
// token on the line the user is hovering.
const FIELD_TO_PREFIX: Record<string, string[]> = {
  BP: ['bp'],
  HR: ['hr'],
  SpO2: ['spo2', 'spo'],
  Temp: ['temp'],
  RR: ['rr'],
  Hb: ['hb'],
  WBC: ['wbc'],
  Creatinine: ['creatinine', 'cr'],
  'Na+': ['na'],
  'K+': ['k'],
  Glucose: ['glucose', 'glu'],
};

interface TokenHit {
  from: number;
  to: number;
}

// findTokensOnLine returns all ranges of `<prefix><non-space>` tokens
// on the given line, for any of the prefixes. Case-insensitive prefix
// match; the value portion can be anything non-whitespace.
function findTokensOnLine(
  lineText: string,
  lineStart: number,
  prefixes: string[]
): TokenHit[] {
  const hits: TokenHit[] = [];
  const lower = lineText.toLowerCase();
  for (const prefix of prefixes) {
    let i = 0;
    while (i < lower.length) {
      const idx = lower.indexOf(prefix, i);
      if (idx < 0) break;
      // Must be a fresh token: previous char is whitespace or line-start.
      const prev = idx > 0 ? lower[idx - 1] : ' ';
      if (!/\s/.test(prev)) {
        i = idx + 1;
        continue;
      }
      // Next char after the prefix should be a digit (so we don't match
      // "rr" inside "thrombocytosis" or similar).
      const after = lower[idx + prefix.length];
      if (!after || !/[0-9.]/.test(after)) {
        i = idx + prefix.length;
        continue;
      }
      // Extend to the next whitespace.
      let end = idx + prefix.length;
      while (end < lineText.length && !/\s/.test(lineText[end])) end += 1;
      hits.push({ from: lineStart + idx, to: lineStart + end });
      i = end;
    }
  }
  return hits;
}

// ─────────────────────────────────────────────────────────────────
// hoverTooltip extension
// ─────────────────────────────────────────────────────────────────

const tooltipExtension = hoverTooltip((view, pos): Tooltip | null => {
  const markers = view.state.field(rangeMarkersField, false) ?? [];
  if (markers.length === 0) return null;

  const line = view.state.doc.lineAt(pos);

  // For each marker, scan the line for matching tokens. Return the
  // first hit that contains the hover position.
  for (const marker of markers) {
    const prefixes = FIELD_TO_PREFIX[marker.field];
    if (!prefixes) continue;

    const hits = findTokensOnLine(line.text, line.from, prefixes);
    for (const hit of hits) {
      if (pos < hit.from || pos > hit.to) continue;
      return {
        pos: hit.from,
        end: hit.to,
        above: true,
        create: () => {
          const el = document.createElement('div');
          el.className = 'cln-range-tooltip';
          el.style.cssText = [
            'font-family: ui-sans-serif, system-ui, sans-serif',
            'font-size: 12px',
            'padding: 6px 10px',
            'border: 1px solid hsl(var(--border))',
            'background: hsl(var(--popover))',
            'color: hsl(var(--popover-foreground))',
            'border-radius: 6px',
            'box-shadow: 0 4px 10px rgba(0,0,0,0.08)',
            'max-width: 280px',
          ].join(';');
          el.innerHTML = `
            <div style="font-weight:600;margin-bottom:2px">
              ${escapeHtml(marker.field)} ${escapeHtml(marker.value)}
            </div>
            <div style="color:hsl(var(--muted-foreground));font-size:11px">
              outside ${escapeHtml(marker.reference_range)} · ${escapeHtml(marker.source)}
            </div>
          `;
          return { dom: el };
        },
      };
    }
  }
  return null;
});

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

export const rangeTooltipExtensions = [rangeMarkersField, tooltipExtension];
