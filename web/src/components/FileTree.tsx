import { useRef, useState, type KeyboardEvent } from 'react';
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  useDraggable,
  useDroppable,
  pointerWithin,
  type DragEndEvent,
  type DragStartEvent,
} from '@dnd-kit/core';
import {
  restrictToVerticalAxis,
  restrictToFirstScrollableAncestor,
} from '@dnd-kit/modifiers';
import {
  ChevronRight,
  ChevronDown,
  Copy,
  CornerLeftUp,
  File,
  Folder,
  FolderOpen,
  Pencil,
  Plus,
  FolderPlus,
  RefreshCw,
  Trash2,
  MoreVertical,
} from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import { PromptDialog } from '@/components/PromptDialog';
import { cn } from '@/lib/utils';
import { useStore } from '@/lib/store';
import type { FileEntry } from '@/lib/types';

// ─────────────────────────────────────────────────────────────────
// Path helpers
// ─────────────────────────────────────────────────────────────────

function basename(p: string): string {
  const i = p.lastIndexOf('/');
  return i >= 0 ? p.slice(i + 1) : p;
}

function dirname(p: string): string {
  const i = p.lastIndexOf('/');
  return i >= 0 ? p.slice(0, i) : '';
}

// Reject moves where the source would become a child of itself.
function isAncestor(ancestor: string, candidate: string): boolean {
  if (!ancestor) return false;
  return candidate === ancestor || candidate.startsWith(ancestor + '/');
}

// ─────────────────────────────────────────────────────────────────
// Rename dialog state
// ─────────────────────────────────────────────────────────────────

interface RenameTarget {
  path: string;
  isDir: boolean;
}

// ─────────────────────────────────────────────────────────────────
// Tree node
// ─────────────────────────────────────────────────────────────────

interface TreeNodeProps {
  entry: FileEntry;
  depth: number;
  expandedPaths: Set<string>;
  onToggleExpand: (path: string) => void;
  onRequestDelete: (path: string) => void;
  onRequestRename: (target: RenameTarget) => void;
}

function TreeNode({
  entry,
  depth,
  expandedPaths,
  onToggleExpand,
  onRequestDelete,
  onRequestRename,
}: TreeNodeProps) {
  const expanded = expandedPaths.has(entry.path);
  const activePath = useStore((s) => s.activePath);
  const openFile = useStore((s) => s.openFile);
  const duplicateFile = useStore((s) => s.duplicateFile);
  const newFileAt = useStore((s) => s.newFileAt);

  const isActive = activePath === entry.path;
  const [menuOpen, setMenuOpen] = useState(false);

  // dnd-kit hooks. Each node is both draggable (you can move it) and,
  // if it's a folder, a drop target. The workspace root is its own
  // drop target rendered by FileTree.
  const draggable = useDraggable({
    id: entry.path,
    data: { path: entry.path, isDir: entry.is_dir },
  });
  const droppable = useDroppable({
    id: entry.is_dir ? entry.path : `__file__${entry.path}`,
    disabled: !entry.is_dir,
    data: { folderPath: entry.path },
  });

  const dragStyle = draggable.transform
    ? {
        transform: `translate(${draggable.transform.x}px, ${draggable.transform.y}px)`,
      }
    : undefined;

  const onOpen = () =>
    openFile(entry.path).catch((err: Error) =>
      toast.error(err.message ?? 'Failed to open file')
    );

  const onDuplicate = async () => {
    try {
      const newPath = await duplicateFile(entry.path);
      toast.success(`Duplicated to ${newPath}`);
    } catch (err) {
      toast.error((err as Error).message ?? 'Duplicate failed');
    }
  };

  if (entry.is_dir) {
    return (
      <div style={dragStyle} className={cn(draggable.isDragging && 'opacity-50')}>
        <ContextMenu>
          <ContextMenuTrigger asChild>
            {/* Droppable wraps ONLY the folder header row, not its
                children. This way dragging a file out of an expanded
                folder doesn't keep re-targeting the parent folder. */}
            <div
              ref={(el) => {
                draggable.setNodeRef(el);
                droppable.setNodeRef(el);
              }}
              className={cn(
                'rounded-sm relative group flex items-center justify-between',
                droppable.isOver &&
                  'ring-1 ring-inset ring-foreground/40 bg-accent/30'
              )}
            >
              <button
                {...draggable.attributes}
                {...draggable.listeners}
                data-tree-item={entry.path}
                data-tree-is-dir="true"
                className="group flex flex-1 items-center gap-1 rounded-sm px-2 py-1 text-left text-sm hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:bg-accent/70"
                style={{ paddingLeft: `${depth * 12 + 6}px` }}
                onClick={() => onToggleExpand(entry.path)}
              >
                {expanded ? (
                  <ChevronDown className="h-3 w-3 shrink-0 text-muted-foreground" />
                ) : (
                  <ChevronRight className="h-3 w-3 shrink-0 text-muted-foreground" />
                )}
                {expanded ? (
                  <FolderOpen className="h-4 w-4 shrink-0 text-muted-foreground" />
                ) : (
                  <Folder className="h-4 w-4 shrink-0 text-muted-foreground" />
                )}
                <span className="truncate">{entry.name}</span>
              </button>

              {/* Mobile options dropdown */}
              <div className="relative pr-1 flex items-center shrink-0">
                <button
                  type="button"
                  onClick={(e) => {
                    e.stopPropagation();
                    setMenuOpen(!menuOpen);
                  }}
                  className="opacity-100 md:opacity-0 group-hover:opacity-100 p-1 hover:bg-accent rounded-sm text-muted-foreground hover:text-foreground transition-all focus:opacity-100"
                  aria-label="More actions"
                  title="More actions"
                >
                  <MoreVertical className="h-3.5 w-3.5" />
                </button>

                {menuOpen && (
                  <>
                    <div
                      className="fixed inset-0 z-40 bg-transparent"
                      onClick={(e) => {
                        e.stopPropagation();
                        setMenuOpen(false);
                      }}
                    />
                    <div className="absolute right-0 top-full mt-0.5 z-50 w-36 rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setMenuOpen(false);
                          void newFileAt(entry.path);
                        }}
                        className="flex w-full items-center gap-2 rounded-sm px-2 py-1 text-left text-xs hover:bg-accent hover:text-accent-foreground"
                      >
                        <Plus className="h-3 w-3" />
                        New file
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setMenuOpen(false);
                          onRequestRename({ path: entry.path, isDir: true });
                        }}
                        className="flex w-full items-center gap-2 rounded-sm px-2 py-1 text-left text-xs hover:bg-accent hover:text-accent-foreground"
                      >
                        <Pencil className="h-3 w-3" />
                        Rename
                      </button>
                      <hr className="my-1 border-border" />
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setMenuOpen(false);
                          onRequestDelete(entry.path);
                        }}
                        className="flex w-full items-center gap-2 rounded-sm px-2 py-1 text-left text-xs text-destructive hover:bg-destructive/10"
                      >
                        <Trash2 className="h-3 w-3 text-destructive" />
                        Delete
                      </button>
                    </div>
                  </>
                )}
              </div>
            </div>
          </ContextMenuTrigger>
          <ContextMenuContent>
            <ContextMenuItem onSelect={() => void newFileAt(entry.path)}>
              <Plus />
              New file
            </ContextMenuItem>
            <ContextMenuItem
              onSelect={() => onRequestRename({ path: entry.path, isDir: true })}
            >
              <Pencil />
              Rename
            </ContextMenuItem>
            <ContextMenuSeparator />
            <ContextMenuItem
              onSelect={() => onRequestDelete(entry.path)}
              className="focus:bg-destructive/10 focus:text-destructive focus:[&>svg]:text-destructive"
            >
              <Trash2 />
              Delete
            </ContextMenuItem>
          </ContextMenuContent>
        </ContextMenu>

        {/* Children render OUTSIDE the droppable wrapper so each child
            is its own drop-detection target. */}
        {expanded &&
          entry.items?.map((child) => (
            <TreeNode
              key={child.path}
              entry={child}
              depth={depth + 1}
              expandedPaths={expandedPaths}
              onToggleExpand={onToggleExpand}
              onRequestDelete={onRequestDelete}
              onRequestRename={onRequestRename}
            />
          ))}
      </div>
    );
  }

  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>
        <div
          ref={draggable.setNodeRef}
          style={dragStyle}
          className={cn(
            'group flex w-full items-center justify-between rounded-sm pr-1 text-sm relative',
            isActive
              ? 'bg-accent text-accent-foreground'
              : 'hover:bg-accent/60 hover:text-accent-foreground',
            draggable.isDragging && 'opacity-50'
          )}
        >
          <button
            {...draggable.attributes}
            {...draggable.listeners}
            data-tree-item={entry.path}
            data-tree-is-dir="false"
            className="flex flex-1 items-center gap-1 truncate px-2 py-1 text-left focus-visible:outline-none focus-visible:bg-accent/70"
            style={{ paddingLeft: `${depth * 12 + 6}px` }}
            onClick={onOpen}
            onDoubleClick={onOpen}
          >
            <File className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            <span className="truncate">{entry.name}</span>
          </button>

          {/* Mobile options dropdown */}
          <div className="relative pr-1 flex items-center shrink-0">
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                setMenuOpen(!menuOpen);
              }}
              className="opacity-100 md:opacity-0 group-hover:opacity-100 p-1 hover:bg-accent rounded-sm text-muted-foreground hover:text-foreground transition-all focus:opacity-100"
              aria-label="More actions"
              title="More actions"
            >
              <MoreVertical className="h-3.5 w-3.5" />
            </button>

            {menuOpen && (
              <>
                <div
                  className="fixed inset-0 z-40 bg-transparent"
                  onClick={(e) => {
                    e.stopPropagation();
                    setMenuOpen(false);
                  }}
                />
                <div className="absolute right-0 top-full mt-0.5 z-50 w-36 rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setMenuOpen(false);
                      void onOpen();
                    }}
                    className="flex w-full items-center gap-2 rounded-sm px-2 py-1 text-left text-xs hover:bg-accent hover:text-accent-foreground"
                  >
                    <File className="h-3 w-3" />
                    Open
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setMenuOpen(false);
                      onRequestRename({ path: entry.path, isDir: false });
                    }}
                    className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-xs hover:bg-accent hover:text-accent-foreground"
                  >
                    <Pencil className="h-3.5 w-3.5" />
                    Rename
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setMenuOpen(false);
                      void onDuplicate();
                    }}
                    className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-xs hover:bg-accent hover:text-accent-foreground"
                  >
                    <Copy className="h-3.5 w-3.5" />
                    Duplicate
                  </button>
                  <hr className="my-1 border-border" />
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setMenuOpen(false);
                      onRequestDelete(entry.path);
                    }}
                    className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-xs text-destructive hover:bg-destructive/10"
                  >
                    <Trash2 className="h-3.5 w-3.5 text-destructive" />
                    Delete
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      </ContextMenuTrigger>
      <ContextMenuContent>
        <ContextMenuItem onSelect={onOpen}>
          <File />
          Open
        </ContextMenuItem>
        <ContextMenuItem
          onSelect={() => onRequestRename({ path: entry.path, isDir: false })}
        >
          <Pencil />
          Rename
        </ContextMenuItem>
        <ContextMenuItem onSelect={() => void onDuplicate()}>
          <Copy />
          Duplicate
        </ContextMenuItem>
        <ContextMenuSeparator />
        <ContextMenuItem
          onSelect={() => onRequestDelete(entry.path)}
          className="focus:bg-destructive/10 focus:text-destructive focus:[&>svg]:text-destructive"
        >
          <Trash2 />
          Delete
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
}

// ─────────────────────────────────────────────────────────────────
// File tree shell + dnd context + dialogs
// ─────────────────────────────────────────────────────────────────

export function FileTree() {
  const files = useStore((s) => s.files);
  const treeLoaded = useStore((s) => s.treeLoaded);
  const refreshTree = useStore((s) => s.refreshTree);
  const newFile = useStore((s) => s.newFile);
  const newFolder = useStore((s) => s.newFolder);
  const deleteFile = useStore((s) => s.deleteFile);
  const renameFile = useStore((s) => s.renameFile);

  const [folderDialogOpen, setFolderDialogOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [renameTarget, setRenameTarget] = useState<RenameTarget | null>(null);
  const [dragLabel, setDragLabel] = useState<string | null>(null);

  // Expanded state lifted out of TreeNode so keyboard navigation can
  // collapse/expand folders without per-node useState gymnastics.
  // Default: top-level expanded.
  const [expandedPaths, setExpandedPaths] = useState<Set<string>>(() => {
    const init = new Set<string>();
    for (const entry of files) if (entry.is_dir) init.add(entry.path);
    return init;
  });
  const onToggleExpand = (path: string) => {
    setExpandedPaths((prev) => {
      const next = new Set(prev);
      if (next.has(path)) next.delete(path);
      else next.add(path);
      return next;
    });
  };
  const expandPath = (path: string) =>
    setExpandedPaths((prev) => new Set(prev).add(path));
  const collapsePath = (path: string) =>
    setExpandedPaths((prev) => {
      const next = new Set(prev);
      next.delete(path);
      return next;
    });

  // The tree container ref is mutable so we can attach it from a
  // combined ref callback (also setting rootDrop.setNodeRef).
  const treeContainerRef = useRef<HTMLDivElement | null>(null);

  // Keyboard navigation. Roving focus across visible [data-tree-item]
  // elements; arrow keys move, Enter opens / toggles, F2 renames,
  // Delete deletes, ←/→ collapse/expand folders.
  const onTreeKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
    if (!treeContainerRef.current) return;
    const items = Array.from(
      treeContainerRef.current.querySelectorAll<HTMLElement>('[data-tree-item]')
    );
    if (items.length === 0) return;

    const active = document.activeElement as HTMLElement | null;
    const path = active?.getAttribute('data-tree-item') ?? null;
    const isDir = active?.getAttribute('data-tree-is-dir') === 'true';
    const idx = path ? items.findIndex((el) => el === active) : -1;

    switch (e.key) {
      case 'ArrowDown': {
        e.preventDefault();
        const next = idx < 0 ? 0 : Math.min(items.length - 1, idx + 1);
        items[next]?.focus();
        break;
      }
      case 'ArrowUp': {
        e.preventDefault();
        const prev = idx <= 0 ? 0 : idx - 1;
        items[prev]?.focus();
        break;
      }
      case 'ArrowRight': {
        if (path && isDir && !expandedPaths.has(path)) {
          e.preventDefault();
          expandPath(path);
        }
        break;
      }
      case 'ArrowLeft': {
        if (path && isDir && expandedPaths.has(path)) {
          e.preventDefault();
          collapsePath(path);
        }
        break;
      }
      case 'Enter': {
        if (active && active instanceof HTMLElement) {
          e.preventDefault();
          active.click();
        }
        break;
      }
      case 'F2': {
        if (path) {
          e.preventDefault();
          setRenameTarget({ path, isDir });
        }
        break;
      }
      case 'Delete': {
        if (path) {
          e.preventDefault();
          setDeleteTarget(path);
        }
        break;
      }
    }
  };

  // Workspace-root drop target — dragging a node onto it moves the
  // node out to the root.
  const rootDrop = useDroppable({ id: '__root__', data: { folderPath: '' } });

  // Slight activation distance so a click doesn't accidentally start
  // a drag. 5px is comfortable on desktop, still works on touch.
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } })
  );

  const onCreateFolder = async (name: string) => {
    try {
      await newFolder(name);
    } catch (err) {
      toast.error((err as Error).message ?? 'Failed to create folder');
    }
  };

  const onConfirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deleteFile(deleteTarget);
      toast.success(`Deleted ${deleteTarget}`);
    } catch (err) {
      toast.error((err as Error).message ?? 'Failed to delete file');
    } finally {
      setDeleteTarget(null);
    }
  };

  const onConfirmRename = async (newName: string) => {
    if (!renameTarget) return;
    const dir = dirname(renameTarget.path);
    const newPath = dir ? `${dir}/${newName}` : newName;
    if (newPath === renameTarget.path) return; // no-op
    try {
      await renameFile(renameTarget.path, newPath);
      toast.success(`Renamed to ${newName}`);
    } catch (err) {
      toast.error((err as Error).message ?? 'Rename failed');
    } finally {
      setRenameTarget(null);
    }
  };

  const onDragStart = (e: DragStartEvent) => {
    setDragLabel(String(e.active.id));
  };

  const onDragEnd = async (e: DragEndEvent) => {
    setDragLabel(null);
    if (!e.over) return;
    const from = String(e.active.id);
    const targetFolder = (e.over.data.current?.folderPath ?? '') as string;

    // No-op: dropped on its own current folder.
    if (dirname(from) === targetFolder) return;
    // Can't move a folder into itself or its own descendant.
    if (e.active.data.current?.isDir && isAncestor(from, targetFolder)) {
      toast.error('Cannot move a folder into itself.');
      return;
    }

    const name = basename(from);
    const to = targetFolder ? `${targetFolder}/${name}` : name;

    try {
      await renameFile(from, to);
      toast.success(`Moved to ${to}`);
    } catch (err) {
      toast.error((err as Error).message ?? 'Move failed');
    }
  };

  return (
    <DndContext
      sensors={sensors}
      onDragStart={onDragStart}
      onDragEnd={onDragEnd}
      onDragCancel={() => setDragLabel(null)}
      collisionDetection={pointerWithin}
      modifiers={[restrictToVerticalAxis, restrictToFirstScrollableAncestor]}
    >
      <div className="flex h-full flex-col">
        <div className="flex items-center justify-between gap-1 border-b border-border px-2 py-1.5">
          <span className="px-1 text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Files
          </span>
          <div className="flex items-center gap-0.5">
            <Button
              size="icon"
              variant="ghost"
              className="h-7 w-7"
              onClick={() =>
                newFile().catch((err: Error) =>
                  toast.error(err.message ?? 'Failed to create file')
                )
              }
              aria-label="New file"
              title="New file"
            >
              <Plus className="h-3.5 w-3.5" />
            </Button>
            <Button
              size="icon"
              variant="ghost"
              className="h-7 w-7"
              onClick={() => setFolderDialogOpen(true)}
              aria-label="New folder"
              title="New folder"
            >
              <FolderPlus className="h-3.5 w-3.5" />
            </Button>
            <Button
              size="icon"
              variant="ghost"
              className="h-7 w-7"
              onClick={() => void refreshTree()}
              aria-label="Refresh"
              title="Refresh"
            >
              <RefreshCw className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>

        <div
          ref={(el) => {
            rootDrop.setNodeRef(el);
            treeContainerRef.current = el;
          }}
          onKeyDown={onTreeKeyDown}
          className={cn(
            'flex-1 overflow-auto py-1',
            rootDrop.isOver && 'ring-1 ring-inset ring-foreground/20'
          )}
        >
          {!treeLoaded ? (
            <div className="px-3 py-4 text-xs text-muted-foreground">
              Loading…
            </div>
          ) : files.length === 0 ? (
            <div className="px-3 py-4 text-xs text-muted-foreground">
              No files yet. Click + to create one.
            </div>
          ) : (
            files.map((entry) => (
              <TreeNode
                key={entry.path}
                entry={entry}
                depth={0}
                expandedPaths={expandedPaths}
                onToggleExpand={onToggleExpand}
                onRequestDelete={setDeleteTarget}
                onRequestRename={setRenameTarget}
              />
            ))
          )}

          {/* Explicit "Move to workspace root" drop target — only
              visible while a drag is in progress. Sits at the bottom
              of the tree and is large enough to be a clear target. */}
          {dragLabel && <RootDropZone isOver={rootDrop.isOver} />}
        </div>

        <DragOverlay dropAnimation={null}>
          {dragLabel ? (
            <div className="pointer-events-none rounded-md border border-border bg-popover px-2 py-1 text-sm shadow-md">
              {basename(dragLabel)}
            </div>
          ) : null}
        </DragOverlay>

        <PromptDialog
          open={folderDialogOpen}
          onOpenChange={setFolderDialogOpen}
          title="New folder"
          description="Folder name, relative to the workspace root."
          label="Path"
          placeholder="notes/2026"
          submitLabel="Create"
          onSubmit={onCreateFolder}
        />

        <PromptDialog
          open={renameTarget !== null}
          onOpenChange={(open) => {
            if (!open) setRenameTarget(null);
          }}
          title={renameTarget?.isDir ? 'Rename folder' : 'Rename file'}
          label="Name"
          placeholder={renameTarget ? basename(renameTarget.path) : ''}
          defaultValue={renameTarget ? basename(renameTarget.path) : ''}
          submitLabel="Rename"
          onSubmit={onConfirmRename}
        />

        <AlertDialog
          open={deleteTarget !== null}
          onOpenChange={(open) => {
            if (!open) setDeleteTarget(null);
          }}
        >
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Delete file?</AlertDialogTitle>
              <AlertDialogDescription>
                <span className="font-mono">{deleteTarget}</span> will be
                permanently removed from the workspace. This action cannot be
                undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction
                onClick={onConfirmDelete}
                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              >
                Delete
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </DndContext>
  );
}

// ─────────────────────────────────────────────────────────────────
// RootDropZone — shown only while a drag is in progress. Gives the
// user an unambiguous place to drop a file when they want to move it
// OUT of a folder back to the workspace root. Without this, the only
// way out is to find empty space below the tree, which often doesn't
// exist when the tree fills the sidebar.
// ─────────────────────────────────────────────────────────────────

function RootDropZone({ isOver }: { isOver: boolean }) {
  const drop = useDroppable({ id: '__root_zone__', data: { folderPath: '' } });
  return (
    <div
      ref={drop.setNodeRef}
      className={cn(
        'mx-2 mt-3 flex items-center gap-2 rounded-md border border-dashed px-3 py-2 text-xs text-muted-foreground transition-colors',
        drop.isOver || isOver
          ? 'border-foreground/40 bg-accent/40 text-foreground'
          : 'border-border'
      )}
    >
      <CornerLeftUp className="h-3.5 w-3.5" />
      <span>Drop here to move to workspace root</span>
    </div>
  );
}
