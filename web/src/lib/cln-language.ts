import {
  StreamLanguage,
  type StreamParser,
  HighlightStyle,
  syntaxHighlighting,
} from '@codemirror/language';
import { tags as t } from '@lezer/highlight';

// Core command keywords recognised by the engine. Kept in sync with
// pkg/engine/parser.go coreCommands.
const CORE_COMMANDS = new Set([
  'pt', 'cc', 'hpi', 'pmh', 'dx', 'ddx', 'sx', 'vitals', 'rx', 'rhx',
  'day', 'alg', 'allergy', 'sh', 'fh', 'pe', 'oe', 'lab', 'labs',
  'rad', 'ix', 'id',
]);

// Plugin standalone commands. Currently only obgyn — kept here so the
// editor highlights them when a @profile is loaded. Unknown commands
// still parse via the extension system, just without highlighting.
const PLUGIN_COMMANDS = new Set([
  'lmp', 'edd', 'gpal', 'fhs', 'ctx',
]);

interface ClnState {
  // Once we've seen the first non-whitespace token on a line we
  // remember whether it was a command keyword so the rest of the
  // line tokenises as "argument" text rather than re-matching the
  // keyword table per token.
  afterCommand: boolean;
}

const parser: StreamParser<ClnState> = {
  name: 'cln',
  startState: () => ({ afterCommand: false }),
  copyState: (s) => ({ ...s }),

  token(stream, state) {
    // Comments: # … or // … to end of line. Match at any column.
    if (stream.match(/^#.*$/) || stream.match(/^\/\/.*$/)) {
      return 'lineComment';
    }

    // @profile directive at line start: highlight whole token.
    if (stream.sol() && stream.match(/^@[a-zA-Z0-9_+]+/)) {
      state.afterCommand = true;
      return 'meta';
    }

    if (stream.eatSpace()) return null;

    if (stream.sol()) {
      state.afterCommand = false;
      const word = stream.match(/^[a-zA-Z][a-zA-Z0-9_]*/);
      if (word) {
        const w = (word as RegExpMatchArray)[0].toLowerCase();
        if (CORE_COMMANDS.has(w) || PLUGIN_COMMANDS.has(w)) {
          state.afterCommand = true;
          return 'keyword';
        }
        // Not a known command — still treat as a typed-out word
        // (could be a custom extension command). Fall through to the
        // generic identifier tag.
        state.afterCommand = true;
        return 'variableName';
      }
    }

    // Numbers (possibly decimal, with optional unit suffix like
    // mg, ml, kg, cm, bpm, w, h, m, d, F, C). Match liberally — we
    // only need a token boundary here, not strict validation.
    if (stream.match(/^\d+(\.\d+)?[a-zA-Z%/]*/)) {
      return 'number';
    }

    // BP-style "120/80" and ratios. Two integers separated by "/".
    if (stream.match(/^\d+\/\d+/)) {
      return 'number';
    }

    // Inline plugin-token like "ga:34w" / "fhr:142bpm".
    if (stream.match(/^[a-zA-Z][a-zA-Z0-9_]*:/)) {
      return 'propertyName';
    }

    // Intensity suffixes: +, ++, +++ and the negative form -.
    if (stream.match(/^[+\-]{1,3}/)) {
      return 'operator';
    }

    // Eat one fall-through character so the stream advances.
    stream.next();
    return null;
  },

  languageData: {
    commentTokens: { line: '#' },
  },
};

// Map our token names to highlight tags.
export const clnLanguage = StreamLanguage.define(parser).extension;

// HighlightStyle defines the actual colors. Theme-agnostic — uses CSS
// variables from the active CodeMirror theme. We rely on default
// theme colours for keywords / numbers / etc. to stay theme-neutral.
const clnHighlight = HighlightStyle.define([
  { tag: t.keyword, color: 'var(--cln-keyword, #2563eb)', fontWeight: '600' },
  { tag: t.meta, color: 'var(--cln-meta, #7c3aed)', fontWeight: '600' },
  { tag: t.lineComment, color: 'var(--cln-comment, #64748b)', fontStyle: 'italic' },
  { tag: t.number, color: 'var(--cln-number, #0f766e)' },
  { tag: t.propertyName, color: 'var(--cln-property, #7c3aed)' },
  { tag: t.variableName, color: 'var(--cln-variable, #b45309)' },
  { tag: t.operator, color: 'var(--cln-operator, #b45309)' },
]);

export const clnHighlighting = syntaxHighlighting(clnHighlight);

// All-in-one extension bundle for the editor.
export const clnExtensions = [clnLanguage, clnHighlighting];
