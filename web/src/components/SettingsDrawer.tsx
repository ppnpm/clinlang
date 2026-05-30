import { useEffect, useState } from 'react';
import { Settings as SettingsIcon, Folder, Palette, Pencil, Accessibility, ArrowLeft, Sliders } from 'lucide-react';
import { toast } from 'sonner';

import { api } from '@/lib/api';

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

type Section = 'workspace' | 'appearance' | 'editor' | 'accessibility' | 'customization';

interface SectionDef {
  id: Section;
  label: string;
  icon: typeof Folder;
}

const SECTIONS: SectionDef[] = [
  { id: 'workspace', label: 'Workspace', icon: Folder },
  { id: 'appearance', label: 'Appearance', icon: Palette },
  { id: 'editor', label: 'Editor', icon: Pencil },
  { id: 'accessibility', label: 'Accessibility', icon: Accessibility },
  { id: 'customization', label: 'Shorthand & Terms', icon: Sliders },
];

// SettingsDrawer — Obsidian-style two-pane settings dialog: a slim
// category rail on the left, content on the right. Minimal copy,
// generous spacing, single visible action per section.
// Mobile-responsive: displays a sliding drawer layout where categories list
// is shown first, and selecting one slides into that category's panel.
export function SettingsDrawer() {
  const [section, setSection] = useState<Section>('workspace');
  const [activeTabOpen, setActiveTabOpen] = useState(false);

  return (
    <Dialog onOpenChange={(open) => { if (!open) setActiveTabOpen(false); }}>
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
      <DialogContent className="overflow-hidden p-0 w-[95vw] sm:max-w-3xl h-[80vh] sm:h-[480px] max-h-[600px]">
        <DialogTitle className="sr-only">Settings</DialogTitle>
        <div className="flex h-full">
          {/* Category rail */}
          <nav
            className={cn(
              "w-full shrink-0 flex-col gap-0.5 border-r border-border bg-muted/30 p-2 md:w-48 transition-all duration-200",
              activeTabOpen ? "hidden md:flex" : "flex"
            )}
            aria-label="Settings categories"
          >
            <span className="px-2 py-1.5 text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Settings
            </span>
            {SECTIONS.map((s) => (
              <button
                key={s.id}
                onClick={() => {
                  setSection(s.id);
                  setActiveTabOpen(true);
                }}
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
          <div
            className={cn(
              "flex-1 overflow-y-auto p-6 transition-all duration-200",
              activeTabOpen ? "block" : "hidden md:block"
            )}
          >
            {activeTabOpen && (
              <Button
                variant="ghost"
                size="sm"
                className="mb-4 gap-1.5 md:hidden"
                onClick={() => setActiveTabOpen(false)}
              >
                <ArrowLeft className="h-4 w-4" />
                Back to categories
              </Button>
            )}
            {section === 'workspace' && <WorkspaceSection />}
            {section === 'appearance' && <AppearanceSection />}
            {section === 'editor' && <EditorSection />}
            {section === 'accessibility' && <AccessibilitySection />}
            {section === 'customization' && <CustomizationSection />}
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
// Section: Accessibility
// ─────────────────────────────────────────────────────────────────

function AccessibilitySection() {
  const autosaveOn = useStore((s) => s.autosaveOn);
  const setAutosaveOn = useStore((s) => s.setAutosaveOn);

  const fontSize = useStore((s) => s.fontSize);
  const setFontSize = useStore((s) => s.setFontSize);

  const editorFont = useStore((s) => s.editorFont);
  const setEditorFont = useStore((s) => s.setEditorFont);

  const lineSpacing = useStore((s) => s.lineSpacing);
  const setLineSpacing = useStore((s) => s.setLineSpacing);

  const highContrastFocus = useStore((s) => s.highContrastFocus);
  const setHighContrastFocus = useStore((s) => s.setHighContrastFocus);

  return (
    <SectionShell title="Accessibility & Editor Settings">
      <Row
        label="Autosave"
        hint="Automatically saves changes 1.5 seconds after you stop typing."
      >
        <Toggle
          pressed={autosaveOn}
          onPressedChange={setAutosaveOn}
          variant="outline"
        >
          {autosaveOn ? 'Autosave Enabled' : 'Autosave Disabled'}
        </Toggle>
      </Row>

      <Row
        label="Font Size"
        hint="Adjust font size for the editor and SOAP preview."
      >
        <div className="flex gap-2">
          {(['sm', 'md', 'lg', 'xl'] as const).map((sz) => (
            <Toggle
              key={sz}
              pressed={fontSize === sz}
              onPressedChange={() => setFontSize(sz)}
              variant="outline"
              className="uppercase"
            >
              {sz}
            </Toggle>
          ))}
        </div>
      </Row>

      <Row
        label="Editor Font Family"
        hint="Select a readable font. Atkinson and OpenDyslexic are optimized for legibility."
      >
        <div className="grid grid-cols-2 gap-2">
          {(
            [
              { value: 'mono', label: 'Monospace' },
              { value: 'atkinson', label: 'Atkinson Sans' },
              { value: 'dyslexic', label: 'OpenDyslexic' },
              { value: 'sans', label: 'Inter Sans' },
            ] as const
          ).map((f) => (
            <Toggle
              key={f.value}
              pressed={editorFont === f.value}
              onPressedChange={() => setEditorFont(f.value)}
              variant="outline"
              className="justify-start px-3"
            >
              {f.label}
            </Toggle>
          ))}
        </div>
      </Row>

      <Row
        label="Line Spacing"
        hint="Increase vertical space between lines of text."
      >
        <div className="flex gap-2">
          {(
            [
              { value: 'normal', label: 'Normal' },
              { value: 'relaxed', label: 'Relaxed' },
              { value: 'double', label: 'Double' },
            ] as const
          ).map((sp) => (
            <Toggle
              key={sp.value}
              pressed={lineSpacing === sp.value}
              onPressedChange={() => setLineSpacing(sp.value)}
              variant="outline"
            >
              {sp.label}
            </Toggle>
          ))}
        </div>
      </Row>

      <Row
        label="Enhanced Keyboard Focus"
        hint="Show high-contrast focus rings around active elements for keyboard navigation."
      >
        <Toggle
          pressed={highContrastFocus}
          onPressedChange={setHighContrastFocus}
          variant="outline"
        >
          {highContrastFocus ? 'Enhanced Focus On' : 'Enhanced Focus Off'}
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

// ─────────────────────────────────────────────────────────────────
// Section: Customization (Shorthand & Terms)
// ─────────────────────────────────────────────────────────────────

const CONFIG_FILES = [
  { value: 'abbreviations.json', label: 'General Abbreviations' },
  { value: 'drugs.json', label: 'Custom Drugs' },
  { value: 'reference_ranges.json', label: 'Reference Ranges' },
  { value: 'frequencies.json', label: 'Prescription Frequencies' },
  { value: 'routes.json', label: 'Prescription Routes' },
  { value: 'symptoms.json', label: 'Symptom Intensities' },
  { value: 'rad_keys.json', label: 'Radiology Keys' },
  { value: 'durations.json', label: 'Duration Units' },
];

function CustomizationSection() {
  const [selectedFile, setSelectedFile] = useState('abbreviations.json');
  const [jsonText, setJsonText] = useState('');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let active = true;
    const fetchConfig = async () => {
      setLoading(true);
      try {
        const data = await api.getConfig(selectedFile);
        if (active) {
          setJsonText(JSON.stringify(data, null, 2));
        }
      } catch (err) {
        if (active) {
          toast.error(err instanceof Error ? err.message : 'Failed to load configuration');
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    };
    fetchConfig();
    return () => {
      active = false;
    };
  }, [selectedFile]);

  const handleSave = async () => {
    try {
      JSON.parse(jsonText);
    } catch (err) {
      toast.error(`Invalid JSON syntax: ${(err as Error).message}`);
      return;
    }

    setSaving(true);
    try {
      await api.saveConfig(selectedFile, jsonText);
      toast.success(`Successfully saved ${selectedFile}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to save configuration');
    } finally {
      setSaving(false);
    }
  };

  return (
    <SectionShell title="Shorthand & Terms Customization">
      <Row
        label="Configuration File"
        hint="Select the configuration file you want to edit. Overrides will be saved to your workspace's .config/ folder."
      >
        <select
          value={selectedFile}
          onChange={(e) => setSelectedFile(e.target.value)}
          className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
        >
          {CONFIG_FILES.map((f) => (
            <option key={f.value} value={f.value} className="bg-popover text-popover-foreground">
              {f.label}
            </option>
          ))}
        </select>
      </Row>

      <div className="flex flex-col gap-2 mt-4">
        <label className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
          JSON Content
        </label>
        {loading ? (
          <div className="flex h-[240px] items-center justify-center rounded-md border border-border bg-muted/10">
            <span className="text-xs text-muted-foreground animate-pulse">Loading configuration...</span>
          </div>
        ) : (
          <textarea
            value={jsonText}
            onChange={(e) => setJsonText(e.target.value)}
            className="font-mono text-xs p-3 rounded-md border border-border bg-muted/10 h-[240px] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring w-full resize-none leading-relaxed"
            placeholder="{}"
            spellCheck={false}
          />
        )}
      </div>

      <div className="flex justify-end gap-2 mt-6">
        <Button
          onClick={handleSave}
          disabled={loading || saving}
          className="gap-2"
        >
          {saving ? 'Saving...' : 'Save Customization'}
        </Button>
      </div>
    </SectionShell>
  );
}
