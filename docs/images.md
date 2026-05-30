# Adding Image Attachments to Cases

ClinLang supports linking high-resolution clinical images (such as wound photos, ECG strips, bedside ultrasound scans, or rashes) directly to your cases. This keeps your medical notes visual and structured without slowing down the editing experience.

---

## Non-Blocking Performance

Unlike heavy word processors that freeze or bloat when you insert photos, ClinLang is built for speed:
* **Background Uploads**: When you drop an image, the file is read and uploaded asynchronously in the background. You can keep typing without interruption.
* **Lightweight Text Notes**: The note itself remains a simple, clean text file. Images are stored separately in your workspace under an `images/` directory, referencing only their filename path (e.g., `img images/wound.png`) in the text note.

---

## 1. Drag & Drop Upload (Easiest)

To attach a photo to your note:
1.  Drag the image file (PNG, JPG, JPEG, GIF, or SVG) from your computer.
2.  Drop it anywhere inside the **Editor** pane.
3.  A loader notification ("Uploading image...") will appear in the bottom right.
4.  Once uploaded, ClinLang automatically inserts the text shorthand at the exact position you dropped the file:
    ```text
    img images/1715401242312-my_photo.jpg
    ```
5.  The upload success toast is shown, and the preview panel updates immediately.

---

## 2. Manual Text Reference

If you manually copy photos into the `images/` subdirectory of your workspace, or want to reference an existing image file, you can write the shorthand command on a new line:

```text
img images/patient_rash.png
```
or
```text
image images/patient_rash.png
```

The clinical note parser will extract these image references and route them into the **Objective** block.

---

## 3. Previewing and Opening Attachments

When a note contains at least one image reference, the **SOAP Preview** pane displays an **Attachments** dock at the bottom of the screen:

* **Thumbnails**: View a quick horizontal gallery of all images referenced in the case.
* **Viewing full size**: Hovering over a thumbnail shows a "View" overlay. Clicking on any thumbnail opens the original high-resolution raw image in a new browser tab.
* **Secure Local Storage**: All files are served directly from your workspace via a secure raw binary endpoint, protecting patient privacy and ensuring compatibility with standard image viewers.
