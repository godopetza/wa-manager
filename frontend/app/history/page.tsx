"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { campaigns, CampaignMessage, Campaign } from "@/lib/api";
import { CheckCheck, Clock, XCircle, Mail } from "lucide-react";

const statusIcon = (s: string) => {
  switch (s) {
    case "read": return <CheckCheck size={14} className="text-[#25d366]" />;
    case "delivered": return <CheckCheck size={14} className="text-[#6b7280]" />;
    case "sent": return <Mail size={14} className="text-blue-400" />;
    case "failed": return <XCircle size={14} className="text-red-400" />;
    default: return <Clock size={14} className="text-[#6b7280]" />;
  }
};

function HistoryInner() {
  const params = useSearchParams();
  const campaignId = params.get("campaign");
  const [list, setList] = useState<Campaign[]>([]);
  const [selected, setSelected] = useState<number | null>(campaignId ? Number(campaignId) : null);
  const [messages, setMessages] = useState<CampaignMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMsgs, setLoadingMsgs] = useState(false);

  useEffect(() => {
    campaigns.list().then(setList).finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (!selected) return;
    setLoadingMsgs(true);
    campaigns.messages(selected).then(setMessages).finally(() => setLoadingMsgs(false));
  }, [selected]);

  const stats = {
    sent: messages.filter(m => ["sent","delivered","read"].includes(m.Status)).length,
    delivered: messages.filter(m => ["delivered","read"].includes(m.Status)).length,
    read: messages.filter(m => m.Status === "read").length,
    failed: messages.filter(m => m.Status === "failed").length,
  };

  return (
    <div className="grid grid-cols-5 gap-6">
      {/* Campaign list */}
      <div className="col-span-2">
        <h1 className="text-2xl font-bold mb-4">History</h1>
        {loading ? <p className="text-[#6b7280]">Loading…</p> : (
          <div className="space-y-2 max-h-[80vh] overflow-y-auto">
            {list.filter(c => c.Status !== "draft").map(c => (
              <button
                key={c.ID}
                onClick={() => setSelected(c.ID)}
                className={`card w-full text-left transition-colors ${selected === c.ID ? "border-[#25d366]/50" : ""}`}
              >
                <div className="font-medium">{c.Name}</div>
                <div className="text-xs text-[#6b7280] mt-1 flex items-center gap-2">
                  <span className={`badge badge-${c.Status.toLowerCase()}`}>{c.Status}</span>
                  ✓ {c.SentCount} · ✗ {c.FailCount}
                </div>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Messages panel */}
      <div className="col-span-3">
        {!selected ? (
          <div className="card h-full flex items-center justify-center text-[#6b7280]">Select a campaign to view delivery log</div>
        ) : (
          <div>
            {/* Stats bar */}
            <div className="grid grid-cols-4 gap-3 mb-4">
              {[
                { label: "Sent", value: stats.sent, color: "text-blue-400" },
                { label: "Delivered", value: stats.delivered, color: "text-[#6b7280]" },
                { label: "Read", value: stats.read, color: "text-[#25d366]" },
                { label: "Failed", value: stats.failed, color: "text-red-400" },
              ].map(s => (
                <div key={s.label} className="card text-center py-3">
                  <div className={`text-2xl font-bold ${s.color}`}>{s.value}</div>
                  <div className="text-xs text-[#6b7280]">{s.label}</div>
                </div>
              ))}
            </div>

            {loadingMsgs ? <p className="text-[#6b7280]">Loading…</p> : (
              <div className="space-y-2 max-h-[65vh] overflow-y-auto">
                {messages.map(m => (
                  <div key={m.ID} className="card flex items-center justify-between py-3">
                    <div>
                      <div className="flex items-center gap-2">
                        {statusIcon(m.Status)}
                        <span className="font-mono text-sm">+{m.Phone}</span>
                        {m.Contact?.Name && <span className="text-xs text-[#6b7280]">{m.Contact.Name}</span>}
                      </div>
                      {m.FailReason && <div className="text-xs text-red-400 mt-0.5">{m.FailReason}</div>}
                    </div>
                    <div className="text-xs text-[#6b7280] text-right">
                      {m.ReadAt ? `Read ${new Date(m.ReadAt).toLocaleTimeString()}` :
                       m.DeliveredAt ? `Delivered ${new Date(m.DeliveredAt).toLocaleTimeString()}` :
                       m.SentAt ? `Sent ${new Date(m.SentAt).toLocaleTimeString()}` : m.Status}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default function HistoryPage() {
  return (
    <Suspense fallback={<p className="text-[#6b7280]">Loading…</p>}>
      <HistoryInner />
    </Suspense>
  );
}
