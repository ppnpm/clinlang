// Frontend type surface.
//
// Engine / workspace / autocomplete types are GENERATED from the Go
// source by tygo. Run `make types` after changing the Go structs.
// `make check-types` verifies the committed files match a fresh run
// (used by CI to catch drift).
//
// HTTP response wrappers (the `{note, warnings, range_markers}`
// envelopes that pkg/api hand-builds from `map[string]interface{}`)
// are written by hand below because they have no Go struct counterpart.

export type {
  RangeMarker,
  PluginInfo,
  ClinicalCase,
  Patient,
  Vitals,
  Symptom,
  Prescription,
} from './types-engine';
export type { FileEntry } from './types-workspace';
export type { Suggestion } from './types-autocomplete';

import type { RangeMarker } from './types-engine';

export interface HealthResponse {
  status: string;
  service: string;
  mode: string;
  disclaimer: string;
}

export interface SOAPResponse {
  soap: string;
  warnings: string[];
  range_markers: RangeMarker[];
  images?: string[];
}

export interface NoteResponse {
  note: string;
  warnings: string[];
  range_markers: RangeMarker[];
}

export interface MarkdownResponse {
  markdown: string;
  warnings: string[];
  range_markers: RangeMarker[];
  images?: string[];
}

export interface LintResponse {
  warnings: string[];
  range_markers: RangeMarker[];
}

export interface FileContent {
  path: string;
  content: string;
}

export interface WorkspaceInfo {
  path: string;
  mode: 'local' | 'hosted';
  configured: boolean;
  suggested?: string;
}
