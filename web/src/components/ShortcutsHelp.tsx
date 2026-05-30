import { HelpCircle } from 'lucide-react';

import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogTrigger,
  DialogHeader,
  DialogDescription,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

export function ShortcutsHelp() {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          aria-label="Keyboard shortcuts & shorthand guide"
          title="Keyboard shortcuts & shorthand guide"
          className="h-8 w-8"
        >
          <HelpCircle className="h-4 w-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="overflow-hidden p-6 w-[95vw] sm:max-w-2xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-lg font-semibold tracking-tight">ClinLang Guide & Shortcuts</DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            Learn the quick commands and keyboard shortcuts to navigate ClinLang like a pro.
          </DialogDescription>
        </DialogHeader>

        <div className="mt-4 space-y-6">
          {/* Shorthand Syntax guide */}
          <section className="space-y-2">
            <h3 className="text-sm font-semibold tracking-tight text-foreground">ClinLang Shorthand Cheat Sheet</h3>
            <div className="overflow-x-auto rounded-md border border-border bg-muted/20">
              <table className="w-full text-left border-collapse text-xs">
                <thead>
                  <tr className="border-b border-border bg-muted/40">
                    <th className="p-2 font-medium">Command</th>
                    <th className="p-2 font-medium">Example</th>
                    <th className="p-2 font-medium">Output Explanation</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  <tr>
                    <td className="p-2 font-mono font-semibold">pt</td>
                    <td className="p-2 font-mono">pt 42F wt70 ht168</td>
                    <td className="p-2 text-muted-foreground">Patient 42 years old, Female, Weight 70kg, Height 168cm (calculates BMI & BSA)</td>
                  </tr>
                  <tr>
                    <td className="p-2 font-mono font-semibold">cc</td>
                    <td className="p-2 font-mono">cc fever, sore throat</td>
                    <td className="p-2 text-muted-foreground">Chief Complaint: fever, sore throat</td>
                  </tr>
                  <tr>
                    <td className="p-2 font-mono font-semibold">vitals</td>
                    <td className="p-2 font-mono">vitals bp120/80 hr76 spo298 temp98.6f rr16</td>
                    <td className="p-2 text-muted-foreground">Blood pressure, Heart rate, SpO2, Temp (units auto-parsed), Respiratory rate</td>
                  </tr>
                  <tr>
                    <td className="p-2 font-mono font-semibold">rx</td>
                    <td className="p-2 font-mono">rx amoxicillin 500mg tds po for 7d</td>
                    <td className="p-2 text-muted-foreground">Prescription: Amoxicillin 500mg Oral three times daily for 7 days</td>
                  </tr>
                  <tr>
                    <td className="p-2 font-mono font-semibold">labs</td>
                    <td className="p-2 font-mono">labs wbc12.5 hgb11.2</td>
                    <td className="p-2 text-muted-foreground">Labs: WBC (highlighted if out of range), Hgb, etc.</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          {/* Keyboard shortcuts */}
          <section className="space-y-2">
            <h3 className="text-sm font-semibold tracking-tight text-foreground">Editor & Navigation Keyboard Shortcuts</h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
              <div className="flex justify-between items-center p-2 rounded-md bg-muted/30 border border-border">
                <span className="text-muted-foreground">Save Active File</span>
                <kbd className="px-1.5 py-0.5 rounded bg-muted border border-border text-[10px] font-mono shadow-sm">Ctrl + S</kbd>
              </div>
              <div className="flex justify-between items-center p-2 rounded-md bg-muted/30 border border-border">
                <span className="text-muted-foreground">Toggle Preview Pane</span>
                <kbd className="px-1.5 py-0.5 rounded bg-muted border border-border text-[10px] font-mono shadow-sm">Eye Icon / Toggle</kbd>
              </div>
              <div className="flex justify-between items-center p-2 rounded-md bg-muted/30 border border-border">
                <span className="text-muted-foreground">Rename Selected File</span>
                <kbd className="px-1.5 py-0.5 rounded bg-muted border border-border text-[10px] font-mono shadow-sm">F2</kbd>
              </div>
              <div className="flex justify-between items-center p-2 rounded-md bg-muted/30 border border-border">
                <span className="text-muted-foreground">Delete Selected File</span>
                <kbd className="px-1.5 py-0.5 rounded bg-muted border border-border text-[10px] font-mono shadow-sm">Delete</kbd>
              </div>
              <div className="flex justify-between items-center p-2 rounded-md bg-muted/30 border border-border">
                <span className="text-muted-foreground">Navigate Sidebar Files</span>
                <kbd className="px-1.5 py-0.5 rounded bg-muted border border-border text-[10px] font-mono shadow-sm">↑ / ↓ Keys</kbd>
              </div>
              <div className="flex justify-between items-center p-2 rounded-md bg-muted/30 border border-border">
                <span className="text-muted-foreground">Open File / Expand Folder</span>
                <kbd className="px-1.5 py-0.5 rounded bg-muted border border-border text-[10px] font-mono shadow-sm">Enter Key</kbd>
              </div>
            </div>
          </section>
        </div>
      </DialogContent>
    </Dialog>
  );
}
