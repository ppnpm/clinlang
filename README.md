# ClinLang

ClinLang is a lightweight, high-performance shorthand clinical documentation and note-taking tool.

By typing brief, zero-punctuation shorthand, ClinLang automatically translates your text, resolves abbreviations, checks reference ranges, and organizes findings into a clean, standardized **SOAP (Subjective, Objective, Assessment, Plan)** format.

Documentation Website: Visit the clinician guide at [https://ppnpm.github.io/clinlang/](https://ppnpm.github.io/clinlang/)

![ClinLang Web Interface](https://i.ibb.co/QvQJ7ynv/73999-A19-42-FE-489-F-86-A4-66944-F0-CD043.png)

> [!IMPORTANT]
> **Clinical Disclaimer**
> ClinLang is a personal shorthand translation and note-taking utility. It is **not** a medical device. It does not provide clinical decision support, diagnostic assistance, dosing recommendations, or treatment guidelines. All clinical judgments, interpretations, and choices remain the sole responsibility of the clinician. See [DISCLAIMER.md](DISCLAIMER.md) at the repository root.

---

## Features

* **Zero-Punctuation Shorthand:** Type as fast as you think (`pt 34M cc fever 3d vitals hr110 bp130/80`).
* **Medical Abbreviations Expansion:** Automatically expands shorthands like `dm2` to "Type 2 Diabetes Mellitus" and prescription routes/frequencies like `tds po` to "Orally, Three times daily".
* **Interactive Live Preview:** Type shorthand on the left and see the formatted SOAP note render instantly on the right.
* **Drag-and-Drop Image Attachments:** Drag photos, scans, or wound logs directly onto the editor. Images upload asynchronously in the background and display in the preview attachments gallery.
* **Specialty Support:** Built-in rules for pediatric ages (`d`, `w`, `m` durations) and a dedicated `@profile obgyn` plugin for obstetric metrics (gestational age, LMP, EDD, GPAL, fetal heart rate, and contractions).
* **Workspace Customization:** Customize your clinic's abbreviations, drug lists, symptom terms, and reference ranges via `.config/` JSON files directly inside the settings drawer in the web UI.
* **Self-Contained Executable:** The production Go binary embeds the React frontend assets (`go:embed`), allowing you to deploy the entire tool as a single, zero-dependency executable file.

---

## Project Structure

* `pkg/engine/` - Core parsing logic, plugins, and formats (Go library).
* `pkg/api/` - HTTP JSON API server, file handlers, and embedded frontend distribution.
* `cmd/clinlang/` - CLI entry points and server daemon.
* `web/` - Single Page Application React client (Vite, TypeScript, TailwindCSS, CodeMirror 6).
* `docs/` - Comprehensive clinician-focused guides (Quickstart, Fast Writing, Commands, Reference Ranges, Specialty Workflows, Customization, Images, and FAQs).

---

## Running Locally

To run the application locally on your machine, follow either the development setup or the production build instructions.

### 🛠️ 1. Development Mode (Separated Front/Back Servers)

Run the backend and frontend separately to enjoy live reloading and hot module replacement:

#### Step A: Start the Go API Server
Open a terminal in the project root directory and run:
```bash
go run ./cmd/clinlang server --port 8080
```
This starts the Go backend API at `http://localhost:8080`.

#### Step B: Start the Vite Frontend Server
Open a second terminal, navigate into the `web` directory, install dependencies, and run:
```bash
cd web
npm install
npm run dev
```
This starts the frontend development server at `http://localhost:5173`.
*Vite is configured to automatically proxy requests from `http://localhost:5173/api/*` to the Go backend on port `8080`.*

---

### 📦 2. Production Mode (Embedded Single-Binary Build)

To build a production release where the React frontend assets are compiled directly inside the Go binary:

#### Step A: Compile the Frontend
Build the frontend bundle:
```bash
cd web
npm install
npm run build
```
This outputs compiled, optimized assets to the `web/dist` folder.

#### Step B: Embed & Copy Assets
Copy the compiled frontend assets to the Go embedding folder:

*   **On Windows (PowerShell):**
    ```powershell
    Remove-Item -Recurse -Force pkg/api/web-dist
    New-Item -ItemType Directory -Force pkg/api/web-dist
    Copy-Item -Recurse web/dist/* pkg/api/web-dist/
    ```
*   **On macOS/Linux:**
    ```bash
    rm -rf pkg/api/web-dist
    mkdir -p pkg/api/web-dist
    cp -R web/dist/. pkg/api/web-dist/
    ```

#### Step C: Build the Go Binary
Compile the executable:
```bash
go build -o clinlang ./cmd/clinlang
```

#### Step D: Run the Application
Start the unified application:
```bash
# On Windows
./clinlang.exe server --port 8080

# On macOS/Linux
./clinlang server --port 8080
```
Now, navigate to `http://localhost:8080` in your web browser. The Go server will serve the SPA directly, communicate with the API, and handle file operations all through a single port.

---

## CLI Command Usage

If you prefer to run ClinLang in the terminal:

```bash
# Print a plain-text clinical note
clinlang run examples/mi.cln

# Print a SOAP-formatted note (optional --markers flags out-of-range values)
clinlang soap examples/mi.cln --markers

# Export structured JSON data from a case
clinlang json examples/mi.cln

# Run parser lint checks (returns warnings and reference range violations)
clinlang lint examples/mi.cln
```

---

## License

MIT License. See `LICENSE` for details.
