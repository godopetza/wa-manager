import type { Metadata } from "next";
import "./globals.css";
import { Sidebar } from "@/components/sidebar";

export const metadata: Metadata = {
  title: "WA Manager",
  description: "WhatsApp campaign manager",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-[#0f1117] text-white min-h-screen flex">
        <Sidebar />
        <main className="flex-1 ml-64 p-8">{children}</main>
      </body>
    </html>
  );
}
