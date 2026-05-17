"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { MessageSquare, Users, Send, History, LogOut, Zap } from "lucide-react";
import { logout } from "@/lib/api";
import { useRouter } from "next/navigation";

const nav = [
  { href: "/templates", label: "Templates", icon: MessageSquare },
  { href: "/groups", label: "Groups", icon: Users },
  { href: "/campaigns", label: "Campaigns", icon: Send },
  { href: "/history", label: "History", icon: History },
];

export function Sidebar() {
  const path = usePathname();
  const router = useRouter();

  return (
    <aside className="fixed left-0 top-0 h-full w-64 bg-[#1a1d27] border-r border-[#2a2d3a] flex flex-col">
      {/* Logo */}
      <div className="p-6 border-b border-[#2a2d3a]">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-[#25d366] flex items-center justify-center">
            <Zap size={16} className="text-black" />
          </div>
          <span className="font-bold text-lg">WA Manager</span>
        </div>
        <p className="text-xs text-[#6b7280] mt-1">WhatsApp Campaign Tool</p>
      </div>

      {/* Nav */}
      <nav className="flex-1 p-4 space-y-1">
        {nav.map(({ href, label, icon: Icon }) => {
          const active = path.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                active
                  ? "bg-[#25d366]/10 text-[#25d366] border border-[#25d366]/20"
                  : "text-[#9ca3af] hover:text-white hover:bg-[#0f1117]"
              }`}
            >
              <Icon size={18} />
              {label}
            </Link>
          );
        })}
      </nav>

      {/* Logout */}
      <div className="p-4 border-t border-[#2a2d3a]">
        <button
          onClick={() => { logout(); router.push("/login"); }}
          className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-[#9ca3af] hover:text-white hover:bg-[#0f1117] w-full transition-colors"
        >
          <LogOut size={18} />
          Logout
        </button>
      </div>
    </aside>
  );
}
