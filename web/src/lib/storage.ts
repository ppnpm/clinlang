// Lightweight per-device session persistence to localStorage. Survives
// browser reloads so the user doesn't lose their open tabs or UI
// preferences when they refresh. Authoritative file content lives on
// the server — we persist only the LIST of open paths and re-fetch
// content on restore. Dirty edits made between reloads are NOT
// persisted (intentional: the beforeunload warning gives the user a
// chance to save first).

const KEY = 'clinlang.session.v1';

export interface SessionSnapshot {
  workspacePath: string;
  openPaths: string[];
  activePath: string | null;
  sidebarOpen: boolean;
  previewOpen: boolean;
  markersOn: boolean;
  autosaveOn?: boolean;
  fontSize?: 'sm' | 'md' | 'lg' | 'xl';
  editorFont?: 'mono' | 'atkinson' | 'dyslexic' | 'sans';
  lineSpacing?: 'normal' | 'relaxed' | 'double';
  highContrastFocus?: boolean;
}

export function loadSession(workspacePath: string): SessionSnapshot | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.localStorage.getItem(KEY);
    if (!raw) return null;
    const s = JSON.parse(raw) as SessionSnapshot;
    // Sessions are scoped by workspace. A different workspace means
    // a different file tree, different open files — don't cross over.
    if (!s || s.workspacePath !== workspacePath) return null;
    return s;
  } catch {
    return null;
  }
}

export function saveSession(snapshot: SessionSnapshot): void {
  if (typeof window === 'undefined') return;
  if (!snapshot.workspacePath) return;
  try {
    window.localStorage.setItem(KEY, JSON.stringify(snapshot));
  } catch {
    // Quota exceeded, private browsing, etc. — drop silently.
  }
}

export function clearSession(): void {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.removeItem(KEY);
  } catch {
    /* noop */
  }
}
