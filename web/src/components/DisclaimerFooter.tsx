import { useEffect, useState } from 'react';
import { api } from '@/lib/api';

const FALLBACK =
  'ClinLang is a personal note-taking and templating tool — not a medical device. No diagnosis, dosing, or decision support.';

// DisclaimerFooter — single muted line, always visible. The
// medicolegal posture requires the disclaimer to be present in the UI
// at all times, not buried in a settings panel.
export function DisclaimerFooter() {
  const [text, setText] = useState(FALLBACK);

  useEffect(() => {
    let alive = true;
    api
      .health()
      .then((h) => {
        if (alive && h.disclaimer) setText(h.disclaimer);
      })
      .catch(() => {
        /* fallback */
      });
    return () => {
      alive = false;
    };
  }, []);

  return (
    <div
      className="truncate border-t border-border/60 px-3 py-1 text-center text-[10px] text-muted-foreground/80"
      title={text}
    >
      {text}
    </div>
  );
}
