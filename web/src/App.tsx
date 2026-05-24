import { useEffect } from 'react';
import { BrowserRouter, Route, Routes } from 'react-router-dom';

import { AppShell } from '@/components/AppShell';
import { Toaster } from '@/components/ui/sonner';
import { Workspace } from '@/routes/Workspace';
import { ThemeProvider } from '@/components/theme-provider';
import { PWAUpdatePrompt } from '@/components/PWAUpdatePrompt';
import { WelcomeDialog } from '@/components/WelcomeDialog';
import { ConflictDialog } from '@/components/ConflictDialog';
import { ErrorBoundary } from '@/components/ErrorBoundary';
import { hasDirtyFiles, useStore } from '@/lib/store';

// DirtyExitGuard hooks the browser's beforeunload event. If any open
// file has unsaved edits, the browser shows its native "leave site?"
// prompt so the user can cancel and save first.
function DirtyExitGuard() {
  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      const open = useStore.getState().open;
      if (hasDirtyFiles(open)) {
        e.preventDefault();
        // Most browsers ignore the returnValue text and show their
        // own generic message; setting it is still required to
        // trigger the prompt.
        e.returnValue = '';
      }
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, []);
  return null;
}

export function App() {
  return (
    <ErrorBoundary>
      <ThemeProvider defaultTheme="system">
        <BrowserRouter>
          <AppShell>
            <Routes>
              <Route path="/" element={<Workspace />} />
              <Route path="*" element={<Workspace />} />
            </Routes>
          </AppShell>
          <Toaster position="bottom-right" />
          <WelcomeDialog />
          <ConflictDialog />
          <DirtyExitGuard />
          {/* SW + update toast only in production. */}
          {import.meta.env.PROD && <PWAUpdatePrompt />}
        </BrowserRouter>
      </ThemeProvider>
    </ErrorBoundary>
  );
}
