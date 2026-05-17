"use client";

import { useEffect, useState } from "react";
import { groups, ContactGroup } from "@/lib/api";
import { Plus, Users, Trash2, Upload, UserPlus } from "lucide-react";

export default function GroupsPage() {
  const [list, setList] = useState<ContactGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<ContactGroup | null>(null);
  const [newName, setNewName] = useState("");
  const [newPhone, setNewPhone] = useState("");
  const [newContactName, setNewContactName] = useState("");
  const [importing, setImporting] = useState(false);

  const load = () => groups.list().then(setList).finally(() => setLoading(false));
  useEffect(() => { load(); }, []);

  const createGroup = async () => {
    if (!newName.trim()) return;
    await groups.create(newName.trim());
    setNewName("");
    load();
  };

  const addContact = async () => {
    if (!selected || !newPhone.trim()) return;
    await groups.addContact(selected.ID, newPhone.trim(), newContactName.trim());
    setNewPhone(""); setNewContactName("");
    const updated = await groups.get(selected.ID);
    setSelected(updated);
    load();
  };

  const removeContact = async (contactId: number) => {
    if (!selected) return;
    await groups.removeContact(selected.ID, contactId);
    const updated = await groups.get(selected.ID);
    setSelected(updated);
    load();
  };

  const importCSV = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!selected || !e.target.files?.[0]) return;
    setImporting(true);
    const result = await groups.importCSV(selected.ID, e.target.files[0]);
    alert(`Imported ${result.imported} contacts`);
    const updated = await groups.get(selected.ID);
    setSelected(updated);
    setImporting(false);
    load();
  };

  return (
    <div className="grid grid-cols-5 gap-6 h-full">
      {/* Group list */}
      <div className="col-span-2">
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-2xl font-bold">Groups</h1>
        </div>
        <div className="flex gap-2 mb-4">
          <input className="input" placeholder="New group name" value={newName} onChange={e => setNewName(e.target.value)} onKeyDown={e => e.key === "Enter" && createGroup()} />
          <button className="btn-primary flex items-center gap-1" onClick={createGroup}><Plus size={16} /></button>
        </div>
        {loading ? (
          <p className="text-[#6b7280]">Loading…</p>
        ) : (
          <div className="space-y-2">
            {list.map(g => (
              <button
                key={g.ID}
                onClick={() => setSelected(g)}
                className={`card w-full text-left flex items-center justify-between transition-colors ${selected?.ID === g.ID ? "border-[#25d366]/50" : ""}`}
              >
                <div>
                  <div className="font-medium">{g.Name}</div>
                  <div className="text-xs text-[#6b7280] flex items-center gap-1 mt-0.5">
                    <Users size={11} /> {g.Contacts?.length ?? 0} contacts
                  </div>
                </div>
                <button className="text-red-400 hover:text-red-300 p-1" onClick={async e => { e.stopPropagation(); if (confirm("Delete group?")) { await groups.delete(g.ID); load(); if (selected?.ID === g.ID) setSelected(null); }}}>
                  <Trash2 size={14} />
                </button>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Contacts panel */}
      <div className="col-span-3">
        {!selected ? (
          <div className="card h-full flex items-center justify-center text-[#6b7280]">
            Select a group to manage contacts
          </div>
        ) : (
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold">{selected.Name}</h2>
              <label className={`btn-secondary flex items-center gap-2 cursor-pointer ${importing ? "opacity-50" : ""}`}>
                <Upload size={14} /> {importing ? "Importing…" : "Import CSV"}
                <input type="file" accept=".csv" className="hidden" onChange={importCSV} disabled={importing} />
              </label>
            </div>

            {/* Add single contact */}
            <div className="flex gap-2 mb-4">
              <input className="input" placeholder="Phone (+255...)" value={newPhone} onChange={e => setNewPhone(e.target.value)} />
              <input className="input" placeholder="Name (optional)" value={newContactName} onChange={e => setNewContactName(e.target.value)} />
              <button className="btn-primary flex items-center gap-1 whitespace-nowrap" onClick={addContact}><UserPlus size={16} /></button>
            </div>

            <p className="text-xs text-[#6b7280] mb-3">CSV format: phone, name (header row required)</p>

            <div className="space-y-2 max-h-[60vh] overflow-y-auto">
              {selected.Contacts?.length === 0 ? (
                <div className="card text-center py-8 text-[#6b7280]">No contacts yet</div>
              ) : (
                selected.Contacts?.map(ct => (
                  <div key={ct.ID} className="card flex items-center justify-between py-3">
                    <div>
                      <div className="font-mono text-sm">+{ct.Phone}</div>
                      {ct.Name && <div className="text-xs text-[#6b7280]">{ct.Name}</div>}
                    </div>
                    <button className="text-red-400 hover:text-red-300" onClick={() => removeContact(ct.ID)}>
                      <Trash2 size={14} />
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
