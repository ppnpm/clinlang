import { useEffect } from 'react';
import { useRegisterSW } from 'virtual:pwa-register/react';
import { toast } from 'sonner';

// PWAUpdatePrompt registers the service worker and notifies the user
// when a new version of the app is available. Only rendered in
// production builds (see App.tsx).
export function PWAUpdatePrompt() {
  const {
    needRefresh: [needRefresh, setNeedRefresh],
    updateServiceWorker,
  } = useRegisterSW();

  useEffect(() => {
    if (!needRefresh) return;
    toast('A new version is available.', {
      action: {
        label: 'Reload',
        onClick: () => updateServiceWorker(true),
      },
      onDismiss: () => setNeedRefresh(false),
      duration: Infinity,
    });
  }, [needRefresh, setNeedRefresh, updateServiceWorker]);

  return null;
}
