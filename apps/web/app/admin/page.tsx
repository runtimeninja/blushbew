"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function AdminLogin() {
  const apiBase = process.env.NEXT_PUBLIC_API_BASE!;
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function login() {
    setErr(null);
    setLoading(true);
    try {
      const res = await fetch(`${apiBase}/v1/admin/login`, {
        method: "POST",
        credentials: "include",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      const data = await res.json();
      if (!res.ok) {
        setErr(data?.error || "login_failed");
        return;
      }
      router.push("/admin/dashboard");
    } catch {
      setErr("network_error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="max-w-md mx-auto p-6 space-y-4">
      <h1 className="text-2xl font-bold">Admin Login</h1>

      <input className="w-full border rounded-lg p-2" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
      <input className="w-full border rounded-lg p-2" placeholder="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />

      <button className="px-4 py-2 rounded-lg border disabled:opacity-50" disabled={loading} onClick={login}>
        {loading ? "Signing in..." : "Sign in"}
      </button>

      {err && <p className="text-sm text-red-600">Error: {err}</p>}

      <p className="text-sm text-gray-500">
        Uses admin credentials from API env.
      </p>
    </main>
  );
}
