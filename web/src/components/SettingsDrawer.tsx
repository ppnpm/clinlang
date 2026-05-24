import { useEffect, useState } from 'react';
import { Settings as SettingsIcon, Folder, Palette, Pencil } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Toggle } from '@/components/ui/toggle';
import { WorkspacePicker } from '@/components/WorkspacePicker';
import { cn } from '@/lib/utils';
import { useStore } from '@/lib/store';
import { useTheme } from '@/components/theme-provider';

type Section = 'workspace' | 'appearance' | 'editor';

interface SectionDef {
  id: Section;
  label: string;
  icon: typeof Folder;
}

const SECTIONS: SectionDef[] = [
  { id: 'workspace', label: 'Workspace', icon: Folder },
  { id: 'appearance', label: 'Appearance', icon: Palette },
  { id: 'editor', label: 'Editor', icon: Pencil },
];

// SettingsDrawer — Obsidian-style two-pane settings dialog: a slim
// category rail on the left, content on the right. Minimal copy,
// generous spacing, single visible action per section.
export function SettingsDrawer() {
  const [section, setSection] = useState<Section>('workspace');

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          aria-label="Settings"
          title="Settings"
          className="h-8 w-8"
        >
          <SettingsIcon className="h-4 w-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="overflow-hidden p-0 sm:max-w-3xl">
        <DialogTitle className="sr-only">Settings</DialogTitle>
        <div className="flex h-[480px]">
          {/* Category rail */}
          <nav
            className="flex w-48 shrink-0 flex-col gap-0.5 border-r border-border bg-muted/30 p-2"
            aria-label="Settings categories"
          >
            <span className="px-2 py-1.5 text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Settings
            </span>
            {SECTIONS.map((s) => (
              <button
                key={s.id}
                onClick={() => setSection(s.id)}
                className={cn(
                  'flex items-center gap-2 rounded-md px-3 py-2 text-left text-sm transition-colors',
                  section === s.id
                    ? 'bg-accent text-accent-foreground'
                    : 'text-muted-foreground hover:bg-accent/60 hover:text-accent-foreground'
                )}
              >
                <s.icon className="h-4 w-4" />
                {s.label}
              </button>
            ))}
          </nav>

          {/* Content panel */}
          <div className="flex-1 overflow-y-auto p-6">
            {section === 'workspace' && <WorkspaceSection />}
            {section === 'appearance' && <AppearanceSection />}
            {section === 'editor' && <EditorSection />}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// ─────────────────────────────────────────────────────────────────
// Section: Workspace
// ─────────────────────────────────────────────────────────────────

function WorkspaceSection() {
  const workspace = useStore((s) => s.workspace);
  const setWorkspace = useStore((s) => s.setWorkspace);

  const [pathInput, setPathInput] = useState(workspace?.path ?? '');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    setPathInput(workspace?.path ?? '');
  }, [workspace?.path]);

  const onSwitch = async () => {
    if (!pathInput.trim()) {
      toast.error('Workspace path cannot be empty.');
      return;
    }
    if (pathInput.trim() === workspace?.path) {
      toast.message('Already using this workspace.');
      return;
    }
    setBusy(true);
    try {
      await setWorkspace(pathInput.trim());
      toast.success('Workspace switched.');
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to switch workspace';
      toast.error(msg);
    } finally {
      setBusy(false);
    }
  };

  if (workspace?.mode === 'hosted') {
    return (
      <SectionShell title="Workspace">
        <Row label="Folder" hint="Set by the operator. Read-only.">
          <Input value={workspace?.path ?? ''} readOnly disabled />
        </Row>
        <Row label="Mode">
          <div className="flex h-9 items-center text-sm capitalize text-muted-foreground">
            {workspace?.mode}
          </div>
        </Row>
      </SectionShell>
    );
  }

  return (
    <SectionShell title="Workspace">
      <Row
        label="Folder"
        hint="The folder is created if missing. Open tabs close on switch."
      >
        <WorkspacePicker value={pathInput} onChange={setPathInput} />
        <div className="mt-2">
          <Button
            size="sm"
            onClick={onSwitch}
            disabled={busy || !pathInput.trim() || pathInput.trim() === workspace?.path}
          >
            Switch
          </Button>
        </div>
      </Row>

      <Row label="Mode">
        <div className="flex h-9 items-center text-sm capitalize text-muted-foreground">
          {workspace?.mode ?? 'unknown'}
        </div>
      </Row>
    </SectionShell>
  );
}

// ─────────────────────────────────────────────────────────────────
// Section: Appearance
// ─────────────────────────────────────────────────────────────────

function AppearanceSection() {
  const { theme, setTheme } = useTheme();

  return (
    <SectionShell title="Appearance">
      <Row label="Theme" hint="Follows your OS when set to System.">
        <div className="flex gap-2">
          {(['light', 'dark', 'system'] as const).map((t) => (
            <Toggle
              key={t}
              pressed={theme === t}
              onPressedChange={() => setTheme(t)}
              variant="outline"
              className="capitalize"
            >
              {t}
            </Toggle>
          ))}
        </div>
      </Row>
    </SectionShell>
  );
}

// ─────────────────────────────────────────────────────────────────
// Section: Editor
// ─────────────────────────────────────────────────────────────────

function EditorSection() {
  const markersOn = useStore((s) => s.markersOn);
  const setMarkers = useStore((s) => s.setMarkers);

  return (
    <SectionShell title="Editor">
      <Row
        label="Out-of-range markers"
        hint="Transcription aid only. Not clinical decision support."
      >
        <Toggle
          pressed={markersOn}
          onPressedChange={setMarkers}
          variant="outline"
        >
          {markersOn ? 'Shown in preview' : 'Hidden in preview'}
        </Toggle>
      </Row>
    </SectionShell>
  );
}

// ─────────────────────────────────────────────────────────────────
// Layout helpers
// ─────────────────────────────────────────────────────────────────

function SectionShell({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="grid gap-6">
      <h2 className="text-base font-semibold tracking-tight">{title}</h2>
      <div className="grid gap-5">{children}</div>
    </div>
  );
}

function Row({
  label,
  hint,
  children,
}: {
  label: string;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="grid gap-1.5">
      <div className="text-sm font-medium">{label}</div>
      {children}
      {hint && <div className="text-xs text-muted-foreground">{hint}</div>}
    </div>
  );
}
