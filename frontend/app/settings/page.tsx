"use client";

import { FormEvent, useEffect, useState } from "react";
import { Download, Save } from "lucide-react";
import { getLogs, getSettings, logsCSVURL, saveSettings } from "@/lib/api";
import type { ActivityLog, Settings } from "@/types/api";

export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [message, setMessage] = useState("");

  useEffect(() => {
    getSettings().then(setSettings).catch(() => setMessage("Login required"));
    getLogs().then(setLogs).catch(() => setLogs([]));
  }, []);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!settings) return;
    try {
      setSettings(await saveSettings(settings));
      setMessage("Settings saved");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Save failed");
    }
  }

  if (!settings) {
    return (
      <main className="min-h-screen bg-field px-6 py-6">
        <p className="text-slate-700">{message || "Loading"}</p>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-field px-6 py-6">
      <section className="mx-auto grid max-w-6xl gap-6 lg:grid-cols-[1fr_1fr]">
        <form onSubmit={submit} className="rounded-md border border-slate-200 bg-white p-5 shadow-sm">
          <h1 className="text-2xl font-semibold">Settings</h1>
          <label className="mt-5 block text-sm font-medium">Upload Folder</label>
          <input className="mt-2 w-full rounded-md border border-slate-300 px-3 py-2" value={settings.upload_directory} onChange={(e) => setSettings({ ...settings, upload_directory: e.target.value })} />

          <div className="mt-5 grid gap-4 sm:grid-cols-2">
            <label className="flex items-center gap-3 rounded-md border border-slate-200 p-3">
              <input type="checkbox" checked={settings.auto_organize} onChange={(e) => setSettings({ ...settings, auto_organize: e.target.checked })} />
              Auto Organize
            </label>
            <label className="flex items-center gap-3 rounded-md border border-slate-200 p-3">
              <input type="checkbox" checked={settings.auto_start_service} onChange={(e) => setSettings({ ...settings, auto_start_service: e.target.checked })} />
              Auto Start Service
            </label>
          </div>

          <label className="mt-5 block text-sm font-medium">Max Concurrent Uploads</label>
          <input className="mt-2 w-full rounded-md border border-slate-300 px-3 py-2" type="number" min={1} value={settings.max_concurrent_uploads} onChange={(e) => setSettings({ ...settings, max_concurrent_uploads: Number(e.target.value) })} />

          <label className="mt-5 block text-sm font-medium">Session Timeout (minutes)</label>
          <input className="mt-2 w-full rounded-md border border-slate-300 px-3 py-2" type="number" min={5} value={settings.session_timeout_minutes} onChange={(e) => setSettings({ ...settings, session_timeout_minutes: Number(e.target.value) })} />

          <label className="mt-5 block text-sm font-medium">Max Upload Size (bytes)</label>
          <input className="mt-2 w-full rounded-md border border-slate-300 px-3 py-2" type="number" min={0} value={settings.max_upload_size ?? ""} onChange={(e) => setSettings({ ...settings, max_upload_size: e.target.value ? Number(e.target.value) : null })} />

          <button className="mt-6 inline-flex items-center gap-2 rounded-md bg-ink px-4 py-2 font-medium text-white">
            <Save size={18} /> Save
          </button>
          {message ? <p className="mt-4 text-sm text-slate-700">{message}</p> : null}
        </form>

        <div className="rounded-md border border-slate-200 bg-white p-5 shadow-sm">
          <div className="flex items-center justify-between gap-3">
            <h2 className="text-xl font-semibold">Activity Logs</h2>
            <a className="inline-flex items-center gap-2 rounded-md bg-sky px-3 py-2 text-sm font-medium text-white" href={logsCSVURL()}>
              <Download size={16} /> CSV
            </a>
          </div>
          <div className="mt-4 max-h-[560px] overflow-auto">
            {logs.map((log) => (
              <div key={log.id} className="border-b border-slate-100 py-3">
                <p className="text-sm font-medium">{log.event_type} · {log.actor || "system"}</p>
                <p className="text-sm text-slate-600">{log.message}</p>
                <p className="text-xs text-slate-400">{new Date(log.created_at).toLocaleString()}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
