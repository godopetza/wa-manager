"use client";

import { useEffect, useState } from "react";
import { templates, WATemplate } from "@/lib/api";
import { Plus, RefreshCw, Trash2, Upload } from "lucide-react";

const statusBadge = (s: string) => (
  <span className={`badge badge-${s.toLowerCase()}`}>{s}</span>
);

export default function TemplatesPage() {
  const [list, setList] = useState<WATemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ name: "", language: "en_US", category: "MARKETING", components: "[]", submitNow: false });
  const [saving, setSaving] = useState(false);

  const load = () => templates.list().then(setList).finally(() => setLoading(false));
  useEffect(() => { load(); }, []);

  const create = async () => {
    setSaving(true);
    try {
      await templates.create({ ...form, components: JSON.parse(form.components) });
      setShowForm(false);
      setForm({ name: "", language: "en_US", category: "MARKETING", components: "[]", submitNow: false });
      load();
    } catch (e: unknown) {
      alert((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const sync = async (id: number) => {
    await templates.sync(id);
    load();
  };

  const del = async (id: number) => {
    if (!confirm("Delete this template?")) return;
    await templates.delete(id);
    load();
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Templates</h1>
          <p className="text-[#6b7280] text-sm mt-1">Manage your WhatsApp message templates</p>
        </div>
        <button className="btn-primary flex items-center gap-2" onClick={() => setShowForm(true)}>
          <Plus size={16} /> New Template
        </button>
      </div>

      {showForm && (
        <div className="card mb-6">
          <h2 className="font-semibold mb-4">Create Template</h2>
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label className="text-xs text-[#6b7280] mb-1 block">Name</label>
              <input className="input" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="my_template_name" />
            </div>
            <div>
              <label className="text-xs text-[#6b7280] mb-1 block">Language</label>
              <select className="input" value={form.language} onChange={e => setForm(f => ({ ...f, language: e.target.value }))}>
                <option value="en_US">English</option>
                <option value="sw">Swahili</option>
                <option value="fr">French</option>
                <option value="pt_BR">Portuguese</option>
              </select>
            </div>
            <div>
              <label className="text-xs text-[#6b7280] mb-1 block">Category</label>
              <select className="input" value={form.category} onChange={e => setForm(f => ({ ...f, category: e.target.value }))}>
                <option>MARKETING</option>
                <option>UTILITY</option>
                <option>AUTHENTICATION</option>
              </select>
            </div>
          </div>
          <div className="mb-4">
            <label className="text-xs text-[#6b7280] mb-1 block">Components JSON</label>
            <textarea
              className="input font-mono text-sm"
              rows={8}
              value={form.components}
              onChange={e => setForm(f => ({ ...f, components: e.target.value }))}
              placeholder='[{"type":"BODY","text":"Hello {{1}}!","example":{"body_text":[["World"]]}}]'
            />
          </div>
          <label className="flex items-center gap-2 text-sm mb-4 cursor-pointer">
            <input type="checkbox" checked={form.submitNow} onChange={e => setForm(f => ({ ...f, submitNow: e.target.checked }))} />
            Submit to Meta for approval immediately
          </label>
          <div className="flex gap-2">
            <button className="btn-primary" onClick={create} disabled={saving}>{saving ? "Saving…" : "Create"}</button>
            <button className="btn-secondary" onClick={() => setShowForm(false)}>Cancel</button>
          </div>
        </div>
      )}

      {loading ? (
        <p className="text-[#6b7280]">Loading…</p>
      ) : list.length === 0 ? (
        <div className="card text-center py-12 text-[#6b7280]">No templates yet. Create your first one above.</div>
      ) : (
        <div className="space-y-3">
          {list.map(t => (
            <div key={t.ID} className="card flex items-center justify-between">
              <div>
                <div className="flex items-center gap-3">
                  <span className="font-mono text-sm font-semibold">{t.Name}</span>
                  {statusBadge(t.MetaStatus)}
                </div>
                <div className="text-xs text-[#6b7280] mt-1">{t.Language} · {t.Category}{t.RejectReason ? ` · Reason: ${t.RejectReason}` : ""}</div>
              </div>
              <div className="flex items-center gap-2">
                <button className="btn-secondary p-2" title="Sync status from Meta" onClick={() => sync(t.ID)}>
                  <RefreshCw size={14} />
                </button>
                {t.MetaStatus === "draft" && (
                  <button className="btn-secondary p-2" title="Submit to Meta" onClick={() => templates.submit(t.ID).then(load)}>
                    <Upload size={14} />
                  </button>
                )}
                <button className="btn-secondary p-2 text-red-400" title="Delete" onClick={() => del(t.ID)}>
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
