# Frequently Asked Questions

---

### Do I need computer programming or coding skills to use ClinLang?
**No, absolutely not.** 
Think of writing ClinLang notes like texting another doctor or typing quick notes on a pad. You just write short shorthand labels, vital numbers, and drug commands. ClinLang does the tedious work of formatting and expanding them into a professional note.

---

### What happens if I make a typo or use a command that doesn't exist?
**It is completely safe. Your work will never be lost or deleted.**
If you type something ClinLang doesn't recognize (like a custom command `pupils:equal`), the system won't crash or trigger an error. It will simply create a separate category in the note (e.g. `PUPILS: equal`) so you can copy it as-is.

---

### Can I write commands in any order?
**Yes.**
You can type `vitals` before `pt`, or put `dx` at the very top. ClinLang automatically reorganizes everything into a standardized SOAP format (Subjective, Objective, Assessment, Plan) regardless of the order you type it in the editor.

---

### Is my patient data sent to external AI servers or the internet?
**No. ClinLang is completely private and local.**
ClinLang uses a local, deterministic rule-based shorthand engine. It does **not** send your clinical notes, patient names, or demographics to any third-party AI models (like OpenAI or Google Gemini). Your data stays entirely on your computer or local hospital network, keeping you compliant with medical privacy laws (like HIPAA or GDPR).

---

### How do I write complex labs or blood test results?
For simple numbers, just type them together (e.g., `hb12.5`, `wbc14000`).
For complex test results, serology, or descriptive reports, use a colon (`:`) to separate the test name from the result:
*   `ix widal:1/160` (titer)
*   `ix abg:pco2:45 po2:90` (arterial blood gas)
*   `ix covid:negative` (serological screen)

---

### Can I customize abbreviations or add new drugs?
**Yes.**
You can customize everything to match your specialty. By opening the **Settings Drawer** (gear icon) in the web UI, you can select files like `abbreviations.json` or `drugs.json` to add your own custom mappings. See **[Customization Guide](customization.md)** for details.

---

### How do I attach clinical photos or wound logs?
You can drag and drop image files directly onto the editor. ClinLang will upload them to your local project workspace and insert a reference in your text. You can also type manual references like:
`img images/wound_day1.jpg`
A scrollable **Attachments** dock will appear in the SOAP preview. See **[Image Attachments Guide](images.md)** for more details.
