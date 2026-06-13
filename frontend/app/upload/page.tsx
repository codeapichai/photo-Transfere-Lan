"use client";

import { useState } from "react";
import { Upload } from "lucide-react";
import { uploadFile } from "@/lib/api";

export default function UploadPage() {
  const [progress, setProgress] = useState(0);
  const [status, setStatus] = useState("Choose photos or videos");

  async function onFiles(files: FileList | null) {
    if (!files?.length) return;
    const uploadToken = new URLSearchParams(window.location.search).get("token");
    for (const file of Array.from(files)) {
      setStatus(`Uploading ${file.name}`);
      const result = await uploadFile(file, navigator.userAgent, uploadToken, (sent) => setProgress(Math.round((sent / file.size) * 100)));
      setStatus(`${file.name}: ${result.status}`);
    }
  }

  return (
    <main className="min-h-screen bg-field px-5 py-6">
      <section className="mx-auto flex max-w-xl flex-col gap-5">
        <h1 className="text-3xl font-semibold text-ink">Upload to PC</h1>
        <label className="flex min-h-60 cursor-pointer flex-col items-center justify-center rounded-md border-2 border-dashed border-slate-300 bg-white p-6 text-center shadow-sm">
          <Upload size={36} />
          <span className="mt-3 text-lg font-medium">{status}</span>
          <input className="sr-only" type="file" multiple accept=".jpg,.jpeg,.png,.heic,.mov,.mp4,image/*,video/*" onChange={(e) => onFiles(e.target.files)} />
        </label>
        <div className="h-3 overflow-hidden rounded-md bg-slate-200">
          <div className="h-full bg-mint transition-all" style={{ width: `${progress}%` }} />
        </div>
      </section>
    </main>
  );
}
