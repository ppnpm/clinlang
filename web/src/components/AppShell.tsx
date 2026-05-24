import { ReactNode, useEffect } from 'react';
import { PanelLeft, PanelRight } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { FileTree } from '@/components/FileTree';
import { StatusBar } from '@/components/StatusBar';
import { DisclaimerFooter } from '@/components/DisclaimerFooter';
import { SettingsDrawer } from '@/components/SettingsDrawer';
import { ThemeToggle } from '@/components/ThemeToggle';
import { ModeBadge } from '@/components/ModeBadge';
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

  useEffect(() => {
    void init();
  }, [init]);

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
            className="h-8 w-8"
            onClick={togglePreview}
            aria-label="Toggle preview"
            title="Toggle preview"
          >
            <PanelRight className="h-4 w-4" />
          </Button>
          <ThemeToggle />
          <SettingsDrawer />
        </div>
      </header>

      {/* Body: sidebar + main */}
      <div className="flex flex-1 overflow-hidden">
        <aside
          className={cn(
            'shrink-0 overflow-hidden border-r border-border bg-muted/20 transition-[width] duration-150',
            sidebarOpen ? 'w-60' : 'w-0'
          )}
        >
          <FileTree />
        </aside>

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
