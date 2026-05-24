import { create } from 'zustand';

import { api, ApiError } from './api';
import { loadSession, saveSession, clearSession } from './storage';
import type { FileEntry, WorkspaceInfo } from './types';

// OpenFile is one tab in the editor area: the path it was loaded from
// plus the in-memory content, a dirty flag, and the ETag of the
// last-saved version (used for concurrency checks on the next save).
export interface OpenFile {
  path: string;
  content: string;
  dirty: boolean;
  etag: string;
}

// ConflictInfo is set when a save fails with 412 Precondition Failed.
// The Workspace route observes this and shows the conflict dialog.
export interface ConflictInfo {
  path: string;
  localContent: string;
}

interface AppState {
  workspace: WorkspaceInfo | null;
  files: FileEntry[];
  treeLoaded: boolean;
  open: Record<string, OpenFile>;
  activePath: string | null;
  sidebarOpen: boolean;
  previewOpen: boolean;
  markersOn: boolean;
  welcomeOpen: boolean;
  conflict: ConflictInfo | null;

  init: () => Promise<void>;
  refreshTree: () => Promise<void>;
  setWorkspace: (path: string) => Promise<void>;
  openFile: (path: string) => Promise<void>;
  closeFile: (path: string) => void;
  updateActiveContent: (next: string) => void;
  saveActive: () => Promise<void>;
  newFile: () => Promise<void>;
  newFileAt: (folder: string) => Promise<void>;
  deleteFile: (path: string) => Promise<void>;
  renameFile: (from: string, to: string) => Promise<void>;
  duplicateFile: (path: string) => Promise<string>;
  newFolder: (path: string) => Promise<void>;
  setActive: (path: string) => void;
  toggleSidebar: () => void;
  togglePreview: () => void;
  setMarkers: (on: boolean) => void;
  closeWelcome: () => void;
  restoreSession: () => Promise<void>;

  // Conflict resolution: force-save the active file ignoring the ETag,
  // OR reload the active file from disk discarding local edits.
  resolveConflictOverwrite: () => Promise<void>;
  resolveConflictReload: () => Promise<void>;
  dismissConflict: () => void;
}

function newUntitledName(open: Record<string, OpenFile>): string {
  let n = 1;
  // eslint-disable-next-line no-constant-condition
  while (true) {
    const candidate = n === 1 ? 'untitled.cln' : `untitled-${n}.cln`;
    if (!(candidate in open)) return candidate;
    n += 1;
  }
}

// hasDirtyFiles is exposed for the beforeunload handler.
export function hasDirtyFiles(open: Record<string, OpenFile>): boolean {
  return Object.values(open).some((f) => f.dirty);
}

// Persist the parts of state that survive a browser reload. Open-file
// CONTENT is intentionally NOT persisted — server is authoritative;
// localStorage holds only the list of paths + UI prefs.
function persist(s: {
  workspace: WorkspaceInfo | null;
  open: Record<string, OpenFile>;
  activePath: string | null;
  sidebarOpen: boolean;
  previewOpen: boolean;
  markersOn: boolean;
}): void {
  if (!s.workspace?.path) return;
  saveSession({
    workspacePath: s.workspace.path,
    openPaths: Object.keys(s.open),
    activePath: s.activePath,
    sidebarOpen: s.sidebarOpen,
    previewOpen: s.previewOpen,
    markersOn: s.markersOn,
  });
}

export const useStore = create<AppState>((set, get) => ({
  workspace: null,
  files: [],
  treeLoaded: false,
  open: {},
  activePath: null,
  sidebarOpen: true,
  previewOpen: true,
  markersOn: false,
  welcomeOpen: false,
  conflict: null,

  init: async () => {
    try {
      const ws = await api.getWorkspace();
      if (!ws.configured) {
        // First launch (or workspace was cleared from config). Show
        // the welcome dialog; don't try to list files (it would 409).
        set({ workspace: ws, welcomeOpen: true, treeLoaded: true });
        return;
      }
      const tree = await api.listFiles();
      set({ workspace: ws, files: tree ?? [], treeLoaded: true });
      await get().restoreSession();
    } catch {
      set({ treeLoaded: true });
    }
  },

  refreshTree: async () => {
    try {
      const tree = await api.listFiles();
      set({ files: tree ?? [], treeLoaded: true });
    } catch {
      set({ treeLoaded: true });
    }
  },

  setWorkspace: async (path) => {
    const ws = await api.setWorkspace(path);
    // New workspace → discard any session that belonged to the old one.
    clearSession();
    set({
      workspace: ws,
      open: {},
      activePath: null,
      treeLoaded: false,
      welcomeOpen: false,
    });
    await get().refreshTree();
    persist(get());
  },

  openFile: async (path) => {
    const { open } = get();
    if (open[path]) {
      set({ activePath: path });
      persist(get());
      return;
    }
    const file = await api.readFile(path);
    set({
      open: {
        ...open,
        [path]: { path, content: file.content, dirty: false, etag: file.etag },
      },
      activePath: path,
    });
    persist(get());
  },

  closeFile: (path) => {
    const { open, activePath } = get();
    const next = { ...open };
    delete next[path];
    let nextActive = activePath;
    if (activePath === path) {
      const remaining = Object.keys(next);
      nextActive = remaining[remaining.length - 1] ?? null;
    }
    set({ open: next, activePath: nextActive });
    persist(get());
  },

  updateActiveContent: (next) => {
    const { open, activePath } = get();
    if (!activePath) return;
    const cur = open[activePath];
    if (!cur) return;
    set({
      open: {
        ...open,
        [activePath]: {
          ...cur,
          content: next,
          dirty: cur.content !== next ? true : cur.dirty,
        },
      },
    });
    // Open-list and active didn't change → no persist needed.
  },

  saveActive: async () => {
    const { open, activePath } = get();
    if (!activePath) return;
    const cur = open[activePath];
    if (!cur) return;
    try {
      const res = await api.writeFile(activePath, cur.content, cur.etag);
      set({
        open: {
          ...open,
          [activePath]: { ...cur, dirty: false, etag: res.etag },
        },
      });
      await get().refreshTree();
    } catch (err) {
      // 412 Precondition Failed → another writer touched the file.
      // Set conflict state so Workspace can show the dialog. Re-throw
      // so the caller's toast UX still fires for other errors.
      if (err instanceof ApiError && err.status === 412) {
        set({ conflict: { path: activePath, localContent: cur.content } });
        return;
      }
      throw err;
    }
  },

  newFile: async () => {
    await get().newFileAt('');
  },

  newFileAt: async (folder) => {
    const { open } = get();
    const base = newUntitledName(open);
    const name = folder ? `${folder.replace(/\/+$/, '')}/${base}` : base;
    const res = await api.writeFile(name, '');
    set({
      open: {
        ...open,
        [name]: { path: name, content: '', dirty: false, etag: res.etag },
      },
      activePath: name,
    });
    await get().refreshTree();
    persist(get());
  },

  deleteFile: async (path) => {
    await api.deleteFile(path);
    const { open, activePath } = get();
    const next = { ...open };
    delete next[path];
    let nextActive = activePath;
    if (activePath === path) {
      const remaining = Object.keys(next);
      nextActive = remaining[remaining.length - 1] ?? null;
    }
    set({ open: next, activePath: nextActive });
    await get().refreshTree();
    persist(get());
  },

  renameFile: async (from, to) => {
    await api.renameFile(from, to);
    const { open, activePath } = get();
    const next = { ...open };
    if (next[from]) {
      next[to] = { ...next[from], path: to };
      delete next[from];
    }
    set({
      open: next,
      activePath: activePath === from ? to : activePath,
    });
    await get().refreshTree();
    persist(get());
  },

  duplicateFile: async (path) => {
    // Read original, write to "<base>-copy.cln". If a copy already
    // exists, append -2, -3, … so duplicating is always safe.
    const file = await api.readFile(path);
    const dot = path.lastIndexOf('.');
    const base = dot > 0 ? path.slice(0, dot) : path;
    const ext = dot > 0 ? path.slice(dot) : '';
    const existing = new Set<string>();
    const collect = (entries: FileEntry[]) => {
      for (const e of entries) {
        existing.add(e.path);
        if (e.items) collect(e.items);
      }
    };
    collect(get().files);

    let candidate = `${base}-copy${ext}`;
    let n = 2;
    while (existing.has(candidate)) {
      candidate = `${base}-copy-${n}${ext}`;
      n += 1;
    }
    await api.writeFile(candidate, file.content);
    await get().refreshTree();
    return candidate;
  },

  newFolder: async (path) => {
    await api.mkdir(path);
    await get().refreshTree();
  },

  setActive: (path) => {
    set({ activePath: path });
    persist(get());
  },
  toggleSidebar: () => {
    set((s) => ({ sidebarOpen: !s.sidebarOpen }));
    persist(get());
  },
  togglePreview: () => {
    set((s) => ({ previewOpen: !s.previewOpen }));
    persist(get());
  },
  setMarkers: (on) => {
    set({ markersOn: on });
    persist(get());
  },
  closeWelcome: () => set({ welcomeOpen: false }),

  // Conflict resolution: user chose to overwrite the on-disk version.
  // We re-issue the save WITHOUT If-Match, which the backend always
  // accepts. Clears the conflict afterwards.
  resolveConflictOverwrite: async () => {
    const { conflict, open } = get();
    if (!conflict) return;
    const cur = open[conflict.path];
    if (!cur) {
      set({ conflict: null });
      return;
    }
    const res = await api.writeFile(conflict.path, cur.content); // no If-Match
    set({
      open: {
        ...open,
        [conflict.path]: { ...cur, dirty: false, etag: res.etag },
      },
      conflict: null,
    });
    await get().refreshTree();
  },

  // Conflict resolution: discard local edits and reload the file from
  // the server. The local content is lost — the dialog warned about
  // this.
  resolveConflictReload: async () => {
    const { conflict, open } = get();
    if (!conflict) return;
    const file = await api.readFile(conflict.path);
    set({
      open: {
        ...open,
        [conflict.path]: {
          path: conflict.path,
          content: file.content,
          dirty: false,
          etag: file.etag,
        },
      },
      conflict: null,
    });
  },

  dismissConflict: () => set({ conflict: null }),

  restoreSession: async () => {
    const { workspace } = get();
    if (!workspace?.path) return;
    const session = loadSession(workspace.path);
    if (!session) return;

    set({
      sidebarOpen: session.sidebarOpen,
      previewOpen: session.previewOpen,
      markersOn: session.markersOn,
    });

    const opened: Record<string, OpenFile> = {};
    for (const path of session.openPaths) {
      try {
        const file = await api.readFile(path);
        opened[path] = {
          path,
          content: file.content,
          dirty: false,
          etag: file.etag,
        };
      } catch {
        // File was deleted or renamed since the last session — skip.
      }
    }
    const activeStillExists =
      session.activePath !== null && session.activePath in opened;
    const activePath = activeStillExists
      ? session.activePath
      : (Object.keys(opened)[0] ?? null);

    set({ open: opened, activePath });
  },
}));
