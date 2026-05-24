import { ReactNode, useEffect, useState } from 'react';
import { toast } from 'sonner';
import {
  ArrowLeft,
  ArrowRight,
  FileText,
  HardDrive,
  Sparkles,
  Folder,
} from 'lucide-react';

import {
  Dialog,
  DialogContent,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { WorkspacePicker } from '@/components/WorkspacePicker';
import { cn } from '@/lib/utils';
import { useStore } from '@/lib/store';

// WelcomeDialog — 4-step onboarding carousel. Three intro slides
// introduce the tool, then the last step collects the workspace
// folder via the native OS picker. Non-dismissible by Escape or
// backdrop click.
export function WelcomeDialog() {
  const open = useStore((s) => s.welcomeOpen);
  const workspace = useStore((s) => s.workspace);
  const setWorkspace = useStore((s) => s.setWorkspace);

  const [step, setStep] = useState(0);
  const [path, setPath] = useState('');
  const [busy, setBusy] = useState(false);

  // Pre-fill the picker with the backend's suggested path the first
  // time the dialog opens.
  useEffect(() => {
    if (open && workspace?.suggested && !path) {
      setPath(workspace.suggested);
    }
  }, [open, workspace?.suggested, path]);

  // Reset to first step every time the dialog reopens (e.g. after
  // user reset their config). Keeping `path` is fine.
  useEffect(() => {
    if (open) setStep(0);
  }, [open]);

  const onContinue = async () => {
    if (!path.trim()) {
      toast.error('Please choose a folder.');
      return;
    }
    setBusy(true);
    try {
      await setWorkspace(path.trim());
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to set workspace';
      toast.error(msg);
    } finally {
      setBusy(false);
    }
  };

  const last = step === STEPS.length - 1;
  const current = STEPS[step];

  return (
    <Dialog open={open} onOpenChange={() => undefined}>
      <DialogContent
        className="overflow-hidden p-0 sm:max-w-md"
        onPointerDownOutside={(e) => e.preventDefault()}
        onEscapeKeyDown={(e) => e.preventDefault()}
      >
        <DialogTitle className="sr-only">Welcome to ClinLang</DialogTitle>

        <div className="flex h-[420px] flex-col">
          {/* Slide content */}
          <div className="flex-1 overflow-y-auto px-6 pt-8">
            {step === 0 && <IntroSlide />}
            {step === 1 && <StorageSlide />}
            {step === 2 && <SyntaxSlide />}
            {step === 3 && (
              <WorkspaceSlide value={path} onChange={setPath} />
            )}
          </div>

          {/* Footer: step dots + nav */}
          <div className="flex items-center justify-between gap-2 border-t border-border px-4 py-3">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setStep((s) => Math.max(0, s - 1))}
              disabled={step === 0}
              className="gap-1.5"
            >
              <ArrowLeft className="h-4 w-4" />
              Back
            </Button>

            <Dots count={STEPS.length} active={step} />

            {last ? (
              <Button
                type="button"
                size="sm"
                onClick={onContinue}
                disabled={busy || !path.trim()}
              >
                Continue
              </Button>
            ) : (
              <Button
                type="button"
                size="sm"
                onClick={() => setStep((s) => Math.min(STEPS.length - 1, s + 1))}
                className="gap-1.5"
              >
                Next
                <ArrowRight className="h-4 w-4" />
              </Button>
            )}
          </div>

          {/* Step title hint (sr/a11y) */}
          <span className="sr-only">{current.label}</span>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// ─────────────────────────────────────────────────────────────────
// Slides
// ─────────────────────────────────────────────────────────────────

function SlideShell({
  icon,
  title,
  children,
}: {
  icon: ReactNode;
  title: string;
  children: ReactNode;
}) {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex h-10 w-10 items-center justify-center rounded-md bg-muted text-foreground">
        {icon}
      </div>
      <h2 className="text-lg font-semibold tracking-tight">{title}</h2>
      <div className="text-sm leading-relaxed text-muted-foreground">
        {children}
      </div>
    </div>
  );
}

function IntroSlide() {
  return (
    <SlideShell icon={<Sparkles className="h-5 w-5" />} title="Welcome to ClinLang">
      <p>
        A fast shorthand notebook for your clinical notes. Type a few
        keystrokes; ClinLang renders SOAP, Markdown, or JSON for you.
      </p>
      <p className="mt-3 text-xs text-muted-foreground/80">
        ClinLang is a personal note-taking tool — not a medical device.
        No diagnosis, dosing, or decision support.
      </p>
    </SlideShell>
  );
}

function StorageSlide() {
  return (
    <SlideShell icon={<HardDrive className="h-5 w-5" />} title="Your notes are plain files">
      <p>
        Every note is a <span className="font-mono">.cln</span> text file
        in a folder you pick. No proprietary database. You can back up,
        sync, version-control, or open them in any editor.
      </p>
      <p className="mt-3 text-xs text-muted-foreground/80">
        Nothing leaves your device unless you put the folder inside a
        sync service like iCloud, OneDrive, or Dropbox.
      </p>
    </SlideShell>
  );
}

function SyntaxSlide() {
  return (
    <SlideShell icon={<FileText className="h-5 w-5" />} title="A taste of the shorthand">
      <p>You type something like:</p>
      <pre className="mt-2 rounded-md border border-border bg-muted/40 p-3 font-mono text-xs leading-relaxed text-foreground">
{`pt 40M wt70
cc fever, cough
vitals bp130/85 hr96 spo295 temp99.5F
rx paracetamol 500mg tds po`}
      </pre>
      <p className="mt-3">
        ClinLang formats it as a clean SOAP note in the preview pane.
        Abbreviations expand only at display time — your raw input is
        preserved verbatim.
      </p>
    </SlideShell>
  );
}

function WorkspaceSlide({
  value,
  onChange,
}: {
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <SlideShell icon={<Folder className="h-5 w-5" />} title="Choose your workspace">
      <p>
        Pick a folder on this device where your notes will live. You
        can change this any time from Settings.
      </p>
      <div className="mt-4">
        <WorkspacePicker value={value} onChange={onChange} />
      </div>
    </SlideShell>
  );
}

// ─────────────────────────────────────────────────────────────────
// Step dots
// ─────────────────────────────────────────────────────────────────

function Dots({ count, active }: { count: number; active: number }) {
  return (
    <div className="flex items-center gap-1.5">
      {Array.from({ length: count }, (_, i) => (
        <span
          key={i}
          className={cn(
            'h-1.5 w-1.5 rounded-full transition-colors',
            i === active ? 'bg-foreground' : 'bg-muted-foreground/30'
          )}
          aria-hidden
        />
      ))}
    </div>
  );
}

const STEPS = [
  { label: 'Welcome' },
  { label: 'Storage' },
  { label: 'Syntax' },
  { label: 'Workspace' },
] as const;
