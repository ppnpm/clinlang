import type {
  FileContent,
  FileEntry,
  HealthResponse,
  LintResponse,
  MarkdownResponse,
  NoteResponse,
  PluginInfo,
  SOAPResponse,
  Suggestion,
  WorkspaceInfo,
} from './types';

// API_BASE:
//   - During `npm run dev` we proxy through Vite to the Go server
//     on :8080 (see vite.config.ts), so an empty base = same origin
//     and the proxy handles it.
//   - In production the Go binary embeds the SPA and serves it from
//     the same origin as the API, so an empty base works as-is.
//   - VITE_API_BASE lets a developer point the SPA at a different
//     backend during dev if needed.
const API_BASE = import.meta.env.VITE_API_BASE ?? '';

class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

// requestWithMeta returns both the parsed body and the raw Response so
// callers can read headers like ETag. Used by readFile / writeFile to
// surface concurrency metadata.
async function requestWithMeta<T>(
  path: string,
  init: RequestInit = {}
): Promise<{ data: T; response: Response }> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(init.headers ?? {}),
    },
    ...init,
  });

  if (!res.ok) {
    const body = await res.text().catch(() => '');
    throw new ApiError(res.status, body || res.statusText);
  }

  if (res.status === 204) {
    return { data: undefined as T, response: res };
  }
  const ct = res.headers.get('content-type') ?? '';
  if (ct.includes('application/json')) {
    return { data: (await res.json()) as T, response: res };
  }
  return { data: (await res.text()) as unknown as T, response: res };
}

// FileReadResult bundles the file content with the server-issued ETag
// so the caller can do an If-Match save later.
export interface FileReadResult extends FileContent {
  etag: string;
}

// FileWriteResult is the response from a successful PUT — includes the
// new ETag the client should remember for next save.
export interface FileWriteResult {
  path: string;
  etag: string;
}

async function request<T>(
  path: string,
  init: RequestInit = {}
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(init.headers ?? {}),
    },
    ...init,
  });

  if (!res.ok) {
    const body = await res.text().catch(() => '');
    throw new ApiError(res.status, body || res.statusText);
  }

  // 204 / empty body
  if (res.status === 204) return undefined as T;

  const ct = res.headers.get('content-type') ?? '';
  if (ct.includes('application/json')) {
    return (await res.json()) as T;
  }
  return (await res.text()) as unknown as T;
}
// encodePath helper splits the path by '/' and encodes each individual
// segment with encodeURIComponent to support special characters like '&', '?', '#'
// while preserving path separators.
function encodePath(path: string): string {
  return path.split('/').map(encodeURIComponent).join('/');
}

export const api = {
  health: () => request<HealthResponse>('/api/v1/health'),

  parse: (input: string) =>
    request<unknown>('/api/v1/parse', {
      method: 'POST',
      body: JSON.stringify({ input }),
    }),

  note: (input: string) =>
    request<NoteResponse>('/api/v1/note', {
      method: 'POST',
      body: JSON.stringify({ input }),
    }),

  soap: (input: string, markers = false) =>
    request<SOAPResponse>('/api/v1/soap', {
      method: 'POST',
      body: JSON.stringify({ input, markers }),
    }),

  markdown: (input: string, markers = false) =>
    request<MarkdownResponse>('/api/v1/markdown', {
      method: 'POST',
      body: JSON.stringify({ input, markers }),
    }),

  lint: (input: string) =>
    request<LintResponse>('/api/v1/lint', {
      method: 'POST',
      body: JSON.stringify({ input }),
    }),

  autocomplete: (command: string, query: string) =>
    request<Suggestion[]>('/api/v1/autocomplete', {
      method: 'POST',
      body: JSON.stringify({ command, query }),
    }),

  drugs: (prefix: string) =>
    request<string[]>(`/api/v1/drugs?prefix=${encodeURIComponent(prefix)}`),

  plugins: () => request<PluginInfo[]>('/api/v1/plugins'),

  listFiles: () => request<FileEntry[]>('/api/v1/files'),

  readFile: async (path: string): Promise<FileReadResult> => {
    const { data, response } = await requestWithMeta<FileContent & { etag?: string }>(
      `/api/v1/files/${encodePath(path)}`
    );
    return {
      path: data.path,
      content: data.content,
      etag: data.etag ?? response.headers.get('ETag') ?? '',
    };
  },

  // writeFile sends If-Match when ifMatchETag is provided. The server
  // returns 412 Precondition Failed if the file has changed on disk
  // since the caller's last read; the ApiError is thrown with status
  // 412 so the UI can offer overwrite-or-reload.
  writeFile: async (
    path: string,
    content: string,
    ifMatchETag?: string
  ): Promise<FileWriteResult> => {
    const headers: Record<string, string> = {};
    if (ifMatchETag) headers['If-Match'] = ifMatchETag;
    const { data, response } = await requestWithMeta<{ path: string; etag?: string }>(
      `/api/v1/files/${encodePath(path)}`,
      {
        method: 'PUT',
        body: JSON.stringify({ content }),
        headers,
      }
    );
    return {
      path: data.path,
      etag: data.etag ?? response.headers.get('ETag') ?? '',
    };
  },

  deleteFile: (path: string) =>
    request<{ path: string }>(`/api/v1/files/${encodePath(path)}`, {
      method: 'DELETE',
    }),

  renameFile: (from: string, to: string) =>
    request<{ from: string; to: string }>('/api/v1/files/rename', {
      method: 'POST',
      body: JSON.stringify({ from, to }),
    }),

  mkdir: (path: string) =>
    request<{ path: string }>('/api/v1/files/mkdir', {
      method: 'POST',
      body: JSON.stringify({ path }),
    }),

  getWorkspace: () => request<WorkspaceInfo>('/api/v1/workspace'),

  setWorkspace: (path: string) =>
    request<WorkspaceInfo>('/api/v1/workspace', {
      method: 'PUT',
      body: JSON.stringify({ path }),
    }),

  // browseWorkspace asks the Go binary to open the OS-native folder
  // picker. Returns the chosen path (or "" if the user cancelled).
  // Throws on transport failure; HTTP 501 means the OS picker is not
  // available (typical on headless Linux) and the UI should fall back
  // to manual entry.
  browseWorkspace: () =>
    request<{ path: string; cancelled: boolean }>(
      '/api/v1/workspace/browse',
      { method: 'POST' }
    ),

  getConfig: (filename: string) =>
    request<any>(`/api/v1/config/${encodeURIComponent(filename)}`),

  saveConfig: (filename: string, content: string) =>
    request<{ status: string; path: string }>(`/api/v1/config/${encodeURIComponent(filename)}`, {
      method: 'PUT',
      body: content,
    }),

  writeBinaryFile: async (path: string, data: Uint8Array): Promise<{ path: string; etag: string }> => {
    const res = await fetch(`${API_BASE}/api/v1/files/${encodePath(path)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/octet-stream' },
      body: data as any,
    });
    if (!res.ok) throw new ApiError(res.status, await res.text().catch(() => '') || res.statusText);
    const json = await res.json();
    return { path: json.path, etag: json.etag || res.headers.get('ETag') || '' };
  },
};

export { ApiError };
