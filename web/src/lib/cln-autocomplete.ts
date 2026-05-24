import {
  autocompletion,
  type CompletionContext,
  type CompletionResult,
} from '@codemirror/autocomplete';

import { api } from './api';

// Core commands list — what the editor offers at the beginning of a
// line before the user has typed any command. Kept here (not fetched)
// because it's small and stable.
const CORE_COMMANDS: Array<{ name: string; description: string }> = [
  { name: 'pt', description: 'Patient demographics' },
  { name: 'id', description: 'Patient identifier' },
  { name: 'cc', description: 'Chief complaint' },
  { name: 'hpi', description: 'History of present illness' },
  { name: 'pmh', description: 'Past medical history' },
  { name: 'sh', description: 'Social history' },
  { name: 'fh', description: 'Family history' },
  { name: 'sx', description: 'Symptoms' },
  { name: 'pe', description: 'Physical exam' },
  { name: 'vitals', description: 'Vital signs' },
  { name: 'lab', description: 'Lab values' },
  { name: 'ix', description: 'Investigations (umbrella)' },
  { name: 'rad', description: 'Imaging / radiology' },
  { name: 'rx', description: 'Prescription' },
  { name: 'rhx', description: 'Past treatment' },
  { name: 'dx', description: 'Diagnosis' },
  { name: 'ddx', description: 'Differential diagnosis' },
  { name: 'day', description: 'Hospital day / timeline' },
  { name: 'alg', description: 'Allergies' },
];

// Cached drug list per query prefix. CodeMirror calls the completion
// source on every keystroke; without caching we'd hit /api/v1/drugs
// dozens of times for the same prefix.
const drugCache = new Map<string, string[]>();

async function fetchDrugs(prefix: string): Promise<string[]> {
  const key = prefix.toLowerCase();
  if (drugCache.has(key)) return drugCache.get(key)!;
  try {
    const matches = await api.drugs(prefix);
    drugCache.set(key, matches ?? []);
    return matches ?? [];
  } catch {
    return [];
  }
}

// completionSource — the brain of the autocomplete. Decides what
// kind of completion to offer based on cursor position:
//
//   - Beginning of line (no command yet) → core command list
//   - Inside `rx <prefix>` → drug names from /api/v1/drugs
//   - Anywhere else → no completion (let the user type freely)
async function completionSource(
  ctx: CompletionContext
): Promise<CompletionResult | null> {
  const line = ctx.state.doc.lineAt(ctx.pos);
  const before = line.text.slice(0, ctx.pos - line.from);

  // Case 1: line start, partial command. Match the word the user is
  // currently typing (or empty if just started a new line).
  const cmdMatch = before.match(/^([a-zA-Z][a-zA-Z0-9_]*)?$/);
  if (cmdMatch !== null) {
    // Don't fire on whitespace-only triggers — user might just be
    // pressing space at line start with no intent to complete.
    if (!ctx.explicit && before.length === 0) return null;
    const partial = (cmdMatch[1] ?? '').toLowerCase();
    return {
      from: line.from,
      options: CORE_COMMANDS.map((c) => ({
        label: c.name,
        detail: c.description,
        type: 'keyword',
        apply: c.name + ' ',
      })).filter((o) =>
        partial === '' ? true : o.label.startsWith(partial)
      ),
      validFor: /^[a-zA-Z]*$/,
    };
  }

  // Case 2: inside an `rx` line — drug-name autocomplete on the last
  // word the user is typing.
  const rxMatch = before.match(/^rx\s+(?:.*\s)?([a-zA-Z][a-zA-Z0-9_-]*)$/i);
  if (rxMatch) {
    const word = rxMatch[1];
    const start = ctx.pos - word.length;
    const matches = await fetchDrugs(word);
    if (matches.length === 0) return null;
    return {
      from: start,
      options: matches.map((m) => ({
        label: m,
        type: 'class',
        apply: m + ' ',
      })),
      validFor: /^[a-zA-Z0-9_-]*$/,
    };
  }

  return null;
}

export const clnAutocomplete = autocompletion({
  override: [completionSource],
  // CodeMirror's default is fine; just a touch shorter so it feels
  // responsive without firing on every keystroke.
  activateOnTyping: true,
  closeOnBlur: true,
  defaultKeymap: true,
  maxRenderedOptions: 12,
});
