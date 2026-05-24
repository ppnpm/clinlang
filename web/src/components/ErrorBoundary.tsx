import { Component, type ReactNode } from 'react';
import { AlertCircle, RefreshCw, Copy } from 'lucide-react';

import { Button } from '@/components/ui/button';

interface State {
  hasError: boolean;
  error: Error | null;
}

// ErrorBoundary — last-resort UI when something throws in render or
// in a lifecycle. Keeps the user on a usable page instead of a blank
// screen and gives them a way to recover (reload) or report the issue.
//
// The medicolegal posture is preserved: even the error screen shows
// the disclaimer.
export class ErrorBoundary extends Component<
  { children: ReactNode },
  State
> {
  override state: State = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  override componentDidCatch(error: Error, info: React.ErrorInfo) {
    // eslint-disable-next-line no-console
    console.error('[ClinLang ErrorBoundary]', error, info);
  }

  private reload = () => {
    window.location.reload();
  };

  private copy = async () => {
    const err = this.state.error;
    if (!err) return;
    const text = `${err.name}: ${err.message}\n\n${err.stack ?? ''}`;
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      // Clipboard not available (insecure context, etc.) — fall back
      // to a textarea selection.
      const ta = document.createElement('textarea');
      ta.value = text;
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
    }
  };

  override render() {
    if (!this.state.hasError) return this.props.children;

    const err = this.state.error;
    return (
      <div className="flex min-h-dvh items-center justify-center bg-background p-6">
        <div className="max-w-md space-y-4 rounded-lg border border-border bg-card p-6 shadow-sm">
          <div className="flex items-center gap-2 text-foreground">
            <AlertCircle className="h-5 w-5" />
            <h1 className="text-base font-semibold">Something went wrong</h1>
          </div>

          <p className="text-sm text-muted-foreground">
            ClinLang hit an unexpected error and couldn't render this view.
            Your notes on disk are unaffected. Reloading usually resolves it.
          </p>

          {err && (
            <details className="rounded-md border border-border bg-muted/30 p-3">
              <summary className="cursor-pointer text-xs font-medium text-muted-foreground">
                Error details
              </summary>
              <pre className="mt-2 max-h-48 overflow-auto whitespace-pre-wrap font-mono text-[11px] text-foreground/80">
                {err.name}: {err.message}
                {'\n\n'}
                {err.stack}
              </pre>
            </details>
          )}

          <div className="flex flex-wrap items-center gap-2">
            <Button size="sm" onClick={this.reload} className="gap-1.5">
              <RefreshCw className="h-3.5 w-3.5" />
              Reload
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={this.copy}
              className="gap-1.5"
            >
              <Copy className="h-3.5 w-3.5" />
              Copy error
            </Button>
          </div>

          <p className="text-[10px] text-muted-foreground/80">
            ClinLang is a personal note-taking and templating tool — not a
            medical device. No diagnosis, dosing, or decision support.
          </p>
        </div>
      </div>
    );
  }
}
