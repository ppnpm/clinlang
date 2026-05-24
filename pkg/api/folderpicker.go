package api

import (
	"bytes"
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

// errPickerUnavailable is returned when no folder-picker tool is
// available on the current OS (e.g. Linux without zenity or kdialog).
// HTTP layer maps this to 501 so the frontend falls back to manual entry.
var errPickerUnavailable = errors.New("native folder picker not available on this system")

// openFolderPicker spawns the OS-native folder selection dialog and
// returns the chosen absolute path. Conventions:
//
//	("path", nil)  → user selected a folder
//	("",     nil)  → user cancelled the dialog
//	("", err)      → real error (tool missing, command failed)
//
// Only called from local-mode requests because spawning a UI process
// from a hosted server makes no sense.
func openFolderPicker() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return openFolderPickerWindows()
	case "darwin":
		return openFolderPickerMacOS()
	case "linux":
		return openFolderPickerLinux()
	default:
		return "", errPickerUnavailable
	}
}

func openFolderPickerWindows() (string, error) {
	// PowerShell single-threaded apartment (-STA) is required for
	// Windows Forms. We use FolderBrowserDialog from the .NET BCL,
	// which is bundled with Windows; no extra dependency to install.
	script := `
		Add-Type -AssemblyName System.Windows.Forms | Out-Null
		$dlg = New-Object System.Windows.Forms.FolderBrowserDialog
		$dlg.Description = 'Choose your ClinLang workspace'
		$dlg.ShowNewFolderButton = $true
		if ($dlg.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
		    Write-Output $dlg.SelectedPath
		}
	`
	cmd := exec.Command("powershell.exe", "-NoProfile", "-STA", "-Command", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		// Some Windows installs hide PowerShell behind a slightly
		// different path; the error here is "executable not found"
		// or the script itself failed. Treat as unavailable.
		return "", errPickerUnavailable
	}
	return strings.TrimSpace(out.String()), nil
}

func openFolderPickerMacOS() (string, error) {
	// osascript returns the POSIX path with a trailing slash and a
	// trailing newline. User cancel produces a non-zero exit with
	// "execution error" on stderr — we detect that and report cancel
	// rather than error so the UX is clean.
	cmd := exec.Command("osascript",
		"-e", `tell application "System Events" to activate`,
		"-e", `POSIX path of (choose folder with prompt "Choose your ClinLang workspace")`,
	)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		if strings.Contains(errOut.String(), "User canceled") ||
			strings.Contains(errOut.String(), "-128") {
			return "", nil
		}
		return "", errPickerUnavailable
	}
	path := strings.TrimSpace(out.String())
	path = strings.TrimSuffix(path, "/")
	return path, nil
}

func openFolderPickerLinux() (string, error) {
	// Try zenity first (GNOME, GTK), then kdialog (KDE). On a
	// headless server or a minimal install, neither is present —
	// caller falls back to the manual-entry input.
	if path, err := tryLinuxTool("zenity",
		"--file-selection", "--directory",
		"--title=Choose your ClinLang workspace",
	); err != errPickerUnavailable {
		return path, err
	}
	if path, err := tryLinuxTool("kdialog",
		"--getexistingdirectory",
		"--title", "Choose your ClinLang workspace",
		".",
	); err != errPickerUnavailable {
		return path, err
	}
	return "", errPickerUnavailable
}

// tryLinuxTool runs name with args. Returns errPickerUnavailable if
// the binary isn't on PATH; (path, nil) on success; ("", nil) on cancel.
func tryLinuxTool(name string, args ...string) (string, error) {
	if _, err := exec.LookPath(name); err != nil {
		return "", errPickerUnavailable
	}
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		// Both zenity and kdialog return non-zero on cancel without
		// any useful distinguishing signal. Treat empty stdout as
		// cancel, anything else as success.
		if out.Len() == 0 {
			return "", nil
		}
	}
	return strings.TrimSpace(out.String()), nil
}
