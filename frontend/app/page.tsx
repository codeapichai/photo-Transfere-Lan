"use client";

import { useEffect, useState } from "react";
import { QRCodeCanvas } from "qrcode.react";
import { Activity, Folder, HardDrive, Wifi } from "lucide-react";
import { StatusTile } from "@/components/StatusTile";
import { createTemporaryToken, getDashboard, wsURL } from "@/lib/api";
import type { Dashboard, SocketEvent } from "@/types/api";

export default function DashboardPage() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [uploadURL, setUploadURL] = useState<string>("");
  const [current, setCurrent] = useState<string>("Idle");
  const [status, setStatus] = useState<string>("Loading");

  useEffect(() => {
    getDashboard()
      .then((data) => {
        setDashboard(data);
        setStatus(data.service_status);
        return createTemporaryToken();
      })
      .then((token) => setUploadURL(token.upload_url))
      .catch((error) => {
        setDashboard(null);
        const text = error instanceof Error ? error.message : "";
        setStatus(text.includes("login required") || text.includes("Unauthorized") ? "Login required" : "Offline");
      });
    const socket = new WebSocket(wsURL());
    socket.onmessage = (message) => {
      const event = JSON.parse(message.data) as SocketEvent;
      if (event.type === "upload_progress") {
        setCurrent(`${event.data.original_filename} ${Math.round((event.data.received_bytes / event.data.filesize) * 100)}%`);
      }
      if (event.type === "upload_complete") {
        setCurrent(`${event.data.original_filename} ${event.data.status}`);
      }
    };
    return () => socket.close();
  }, []);

  return (
    <main className="min-h-screen bg-field">
      <section className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-6">
        <header className="flex flex-col gap-2 border-b border-slate-200 pb-5 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="text-3xl font-semibold text-ink">PhotoTransfer LAN</h1>
            <p className="text-slate-600">Windows receiver for phone photos and videos on this WiFi network.</p>
          </div>
          <a className="rounded-md bg-ink px-4 py-2 text-sm font-medium text-white" href="/setup">
            Setup
          </a>
          <a className="rounded-md bg-sky px-4 py-2 text-sm font-medium text-white" href="/login">
            Login
          </a>
          <a className="rounded-md bg-mint px-4 py-2 text-sm font-medium text-white" href="/settings">
            Settings
          </a>
        </header>

        <div className="grid gap-4 md:grid-cols-4">
          <StatusTile label="Service" value={status} />
          <StatusTile label="Local IP" value={dashboard?.local_ip ?? "-"} />
          <StatusTile label="Files today" value={String(dashboard?.today_files ?? 0)} />
          <StatusTile label="Data today" value={`${Math.round((dashboard?.today_bytes ?? 0) / 1024 / 1024)} MB`} />
        </div>

        <div className="grid gap-6 lg:grid-cols-[360px_1fr]">
          <div className="rounded-md border border-slate-200 bg-white p-5 shadow-sm">
            <div className="flex items-center gap-2 text-lg font-semibold">
              <Wifi size={20} /> Connect Phone
            </div>
            <div className="mt-5 flex justify-center">
              {uploadURL ? <QRCodeCanvas value={uploadURL} size={240} /> : <div className="h-60 w-60 bg-slate-100" />}
            </div>
            <p className="mt-4 break-all rounded-md bg-field p-3 text-sm text-slate-700">{uploadURL || "Login to generate a temporary upload token"}</p>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="rounded-md border border-slate-200 bg-white p-5 shadow-sm">
              <div className="flex items-center gap-2 font-semibold"><Activity size={18} /> Current Upload</div>
              <p className="mt-4 text-2xl font-semibold text-coral">{current}</p>
            </div>
            <div className="rounded-md border border-slate-200 bg-white p-5 shadow-sm">
              <div className="flex items-center gap-2 font-semibold"><Folder size={18} /> Storage</div>
              <p className="mt-4 break-words text-slate-700">{dashboard?.storage_location ?? "-"}</p>
            </div>
            <div className="rounded-md border border-slate-200 bg-white p-5 shadow-sm md:col-span-2">
              <div className="flex items-center gap-2 font-semibold"><HardDrive size={18} /> Queue Status</div>
              <p className="mt-4 text-slate-700">Ready for chunked uploads with SHA256 verification.</p>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
