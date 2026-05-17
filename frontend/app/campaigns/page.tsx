"use client";

import { useEffect, useState } from "react";
import { campaigns, templates, groups, Campaign, WATemplate, ContactGroup } from "@/lib/api";
import { Plus, Send, Eye, Trash2, Image } from "lucide-react";
import Link from "next/link";

const statusBadge = (s: string) => <span className={`badge badge-${s.toLowerCase()}`}>{s}</span>;

export default function CampaignsPage() {
  const [list, setList] = useState<Campaign[]>([]);
  const [templateList, setTemplateList] = useState<WATemplate[]>([]);
  const [groupList, setGroupList] = useState<ContactGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ name: "", templateId: "", groupId: "", imageUrl: "" });
  const [saving, setSaving] = useState(false);
  const [sending, setSending] = useState<number | null>(null);

  const load = () =>
    Promise.all([campaigns.list(), templates.list(), groups.list()])
      .then(([c, t, g]) => { setList(c); setTemplateList(t); setGroupList(g); })
      .finally(() => setLoading(false));

  useEffect(() => { load(); }, []);

  const create = async () => {
    if (!form.name || !form.templateId || !form.groupId) return alert("Fill in all required fields");
    setSaving(true);
    try {
      await campaigns.create({
        name: form.name,
        templateId: Number(form.templateId),
        groupId: Number(form.groupId),
        imageUrl: form.imageUrl,
      });
      setShowForm(false);
      setForm({ name: "", templateId: "", groupId: "", imageUrl: "" });
      load();
    } catch (e: unknown) {
      alert((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const send = async (id: number) => {
    if (!confirm("Start sending this campaign now?")) return;
    setSending(id);
    try {
      const res = await campaigns.send(id);
      alert(`Campaign started — sending to ${res.contacts} contacts`);
      load();
    } catch (e: unknown) {
      alert((e as Error).message);
    } finally {
      setSending(null);
    }
  };

  const del = async (id: number) => {
    if (!confirm("Delete this campaign and all its messages?")) return;
    await campaigns.delete(id);
    load();
  };

  const approvedTemplates = templateList.filter(t => t.MetaStatus === "APPROVED" || t.MetaStatus === "approved");

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Campaigns</h1>
          <p className="text-[#6b7280] text-sm mt-1">Create and send WhatsApp campaigns to contact groups</p>
        </div>
        <button className="btn-primary flex items-center gap-2" onClick={() => setShowForm(true)}>
          <Plus size={16} /> New Campaign
        </button>
      </div>

      {showForm && (
        <div className="card mb-6">
          <h2 className="font-semibold mb-4">Create Campaign</h2>
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label className="text-xs text-[#6b7280] mb-1 block">Campaign Name</label>
              <input className="input" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="Wedding Invites — June 2026" />
            </div>
            <div>
              <label className="text-xs text-[#6b7280] mb-1 block">Template (approved only)</label>
              <select className="input" value={form.templateId} onChange={e => setForm(f => ({ ...f, templateId: e.target.value }))}>
                <option value="">Select template…</option>
                {approvedTemplates.map(t => (
                  <option key={t.ID} value={t.ID}>{t.Name} ({t.Language})</option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-xs text-[#6b7280] mb-1 block">Contact Group</label>
              <select className="input" value={form.groupId} onChange={e => setForm(f => ({ ...f, groupId: e.target.value }))}>
                <option value="">Select group…</option>
                {groupList.map(g => (
                  <option key={g.ID} value={g.ID}>{g.Name} ({g.Contacts?.length ?? 0} contacts)</option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-xs text-[#6b7280] mb-1 block flex items-center gap-1">
                <Image size={12} /> Invitation Card URL (optional — go-invite-render)
              </label>
              <input className="input" value={form.imageUrl} onChange={e => setForm(f => ({ ...f, imageUrl: e.target.value }))} placeholder="https://your-render-service.up.railway.app/render/..." />
              <p className="text-xs text-[#6b7280] mt-1">If set, the rendered PNG is sent before the template message</p>
            </div>
          </div>
          <div className="flex gap-2">
            <button className="btn-primary" onClick={create} disabled={saving}>{saving ? "Creating…" : "Create Campaign"}</button>
            <button className="btn-secondary" onClick={() => setShowForm(false)}>Cancel</button>
          </div>
        </div>
      )}

      {loading ? (
        <p className="text-[#6b7280]">Loading…</p>
      ) : list.length === 0 ? (
        <div className="card text-center py-12 text-[#6b7280]">No campaigns yet.</div>
      ) : (
        <div className="space-y-3">
          {list.map(c => (
            <div key={c.ID} className="card flex items-center justify-between">
              <div>
                <div className="flex items-center gap-3">
                  <span className="font-semibold">{c.Name}</span>
                  {statusBadge(c.Status)}
                </div>
                <div className="text-xs text-[#6b7280] mt-1">
                  {c.Template?.Name} → {c.Group?.Name}
                  {(c.SentCount > 0 || c.FailCount > 0) && (
                    <span> · ✓ {c.SentCount} sent · ✗ {c.FailCount} failed</span>
                  )}
                  {c.ImageURL && <span> · 🖼 card attached</span>}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Link href={`/history?campaign=${c.ID}`} className="btn-secondary p-2" title="View delivery log">
                  <Eye size={14} />
                </Link>
                {(c.Status === "draft" || c.Status === "partial_failed") && (
                  <button className="btn-primary p-2" title="Send now" onClick={() => send(c.ID)} disabled={sending === c.ID}>
                    <Send size={14} />
                  </button>
                )}
                <button className="btn-secondary p-2 text-red-400" onClick={() => del(c.ID)}>
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
