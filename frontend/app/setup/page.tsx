"use client";

import { FormEvent, useState } from "react";
import { setup } from "@/lib/api";

export default function SetupPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    try {
      await setup(username, password);
      setMessage("Setup complete. Return to dashboard.");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Setup failed");
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-field px-6">
      <form onSubmit={submit} className="w-full max-w-md rounded-md border border-slate-200 bg-white p-6 shadow-sm">
        <h1 className="text-2xl font-semibold">First Launch Setup</h1>
        <label className="mt-5 block text-sm font-medium">Username</label>
        <input className="mt-2 w-full rounded-md border border-slate-300 px-3 py-2" value={username} onChange={(e) => setUsername(e.target.value)} />
        <label className="mt-4 block text-sm font-medium">Password</label>
        <input className="mt-2 w-full rounded-md border border-slate-300 px-3 py-2" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
        <button className="mt-6 w-full rounded-md bg-ink px-4 py-2 font-medium text-white">Create Account</button>
        {message ? <p className="mt-4 text-sm text-slate-700">{message}</p> : null}
      </form>
    </main>
  );
}

