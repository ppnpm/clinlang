import { ReactNode, useEffect } from 'react';
import { PanelLeft, Eye, EyeOff } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { FileTree } from '@/components/FileTree';
import { StatusBar } from '@/components/StatusBar';
import { DisclaimerFooter } from '@/components/DisclaimerFooter';
import { SettingsDrawer } from '@/components/SettingsDrawer';
import { ThemeToggle } from '@/components/ThemeToggle';
import { ModeBadge } from '@/components/ModeBadge';
import { ShortcutsHelp } from '@/components/ShortcutsHelp';
import { useStore } from '@/lib/store';
import { cn } from '@/lib/utils';

// AppShell — Obsidian-style three-zone layout:
//   - thin top bar (title, mode badge, panel toggles, theme, settings)
//   - sidebar (file tree) + main (children) + optional right pane
//   - status bar
//
// Sidebar collapses on small screens; user can toggle manually too.
export function AppShell({ children }: { children: ReactNode }) {
  const init = useStore((s) => s.init);
  const sidebarOpen = useStore((s) => s.sidebarOpen);
  const previewOpen = useStore((s) => s.previewOpen);
  const toggleSidebar = useStore((s) => s.toggleSidebar);
  const togglePreview = useStore((s) => s.togglePreview);
  const highContrastFocus = useStore((s) => s.highContrastFocus);

  useEffect(() => {
    void init();
  }, [init]);

  useEffect(() => {
    const root = window.document.documentElement;
    if (highContrastFocus) {
      root.classList.add('high-contrast-focus');
    } else {
      root.classList.remove('high-contrast-focus');
    }
  }, [highContrastFocus]);

  return (
    <div className="flex h-dvh flex-col bg-background">
      {/* Top bar */}
      <header className="flex h-10 shrink-0 items-center justify-between gap-2 border-b border-border bg-background px-2">
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={toggleSidebar}
            aria-label="Toggle sidebar"
            title="Toggle sidebar"
          >
            <PanelLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm font-medium tracking-tight">ClinLang</span>
          <ModeBadge />
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            className={cn(
              "h-8 w-8 transition-colors",
              previewOpen ? "bg-muted text-foreground" : "text-muted-foreground"
            )}
            onClick={togglePreview}
            aria-label={previewOpen ? "Hide preview" : "Show preview"}
            title={previewOpen ? "Hide preview" : "Show preview"}
          >
            {previewOpen ? (
              <Eye className="h-4 w-4" />
            ) : (
              <EyeOff className="h-4 w-4" />
            )}
          </Button>
          <ThemeToggle />
          <ShortcutsHelp />
          <SettingsDrawer />
        </div>
      </header>

      {/* Body: sidebar + main */}
      <div className="flex flex-1 overflow-hidden relative">
        {/* Sidebar */}
        <aside
          className={cn(
            'shrink-0 overflow-hidden border-border bg-muted/20 transition-all duration-150',
            'md:static md:translate-x-0 md:h-full md:border-r',
            sidebarOpen
              ? 'w-60 fixed top-10 bottom-0 left-0 z-40 bg-background border-r shadow-lg translate-x-0'
              : 'w-0 fixed top-10 bottom-0 left-0 -translate-x-full md:translate-x-0 md:border-r-0'
          )}
        >
          <FileTree />
        </aside>

        {/* Mobile Sidebar backdrop */}
        {sidebarOpen && (
          <div
            className="fixed inset-0 top-10 z-30 bg-background/80 backdrop-blur-sm md:hidden"
            onClick={toggleSidebar}
            aria-hidden="true"
          />
        )}

        <main className="flex flex-1 flex-col overflow-hidden">
          {children}
        </main>
      </div>

      <StatusBar />
      <DisclaimerFooter />

      {/* Pass preview visibility through CSS var so Workspace can read
          it without prop drilling. (Workspace reads from the store
          directly; this attribute is just for any deeper component.) */}
      <span data-preview-open={previewOpen} hidden />
    </div>
  );
}
