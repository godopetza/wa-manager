"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { login } from "@/lib/api";
import { Zap } from "lucide-react";

export default function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const router = useRouter();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await login(username, password);
      router.push("/campaigns");
    } catch {
      setError("Invalid credentials");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#0f1117] flex items-center justify-center">
      <div className="w-full max-w-sm">
        <div className="flex items-center gap-3 mb-8 justify-center">
          <div className="w-10 h-10 rounded-xl bg-[#25d366] flex items-center justify-center">
            <Zap size={20} className="text-black" />
          </div>
          <span className="text-2xl font-bold">WA Manager</span>
        </div>

        <div className="card">
          <h1 className="text-lg font-semibold mb-6">Sign in</h1>
          <form onSubmit={submit} className="space-y-4">
            <div>
              <label className="text-xs text-[#6b7280] mb-1 block">Username</label>
              <input className="input" value={username} onChange={e => setUsername(e.target.value)} autoFocus />
            </div>
            <div>
              <label className="text-xs text-[#6b7280] mb-1 block">Password</label>
              <input className="input" type="password" value={password} onChange={e => setPassword(e.target.value)} />
            </div>
            {error && <p className="text-red-400 text-sm">{error}</p>}
            <button className="btn-primary w-full" disabled={loading}>
              {loading ? "Signing in…" : "Sign in"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
