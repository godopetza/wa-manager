const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function getToken() {
  if (typeof window === "undefined") return "";
  return localStorage.getItem("wam_token") ?? "";
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}/api${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getToken()}`,
      ...init?.headers,
    },
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? "request failed");
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

// ── Auth ─────────────────────────────────────────────────────────────────────
export async function login(username: string, password: string) {
  const { token } = await request<{ token: string }>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  });
  localStorage.setItem("wam_token", token);
  return token;
}

export function logout() {
  localStorage.removeItem("wam_token");
}

// ── Templates ────────────────────────────────────────────────────────────────
export type WATemplate = {
  ID: number;
  CreatedAt: string;
  Name: string;
  Language: string;
  Category: string;
  Components: unknown;
  MetaStatus: string;
  MetaID: string;
  RejectReason: string;
};

export const templates = {
  list: () => request<WATemplate[]>("/templates"),
  get: (id: number) => request<WATemplate>(`/templates/${id}`),
  create: (body: object) => request<WATemplate>("/templates", { method: "POST", body: JSON.stringify(body) }),
  submit: (id: number) => request<WATemplate>(`/templates/${id}/submit`, { method: "POST" }),
  sync: (id: number) => request<WATemplate>(`/templates/${id}/sync`, { method: "POST" }),
  uploadImage: (id: number, imageUrl: string) =>
    request<{ handle: string }>(`/templates/${id}/upload-image`, {
      method: "POST",
      body: JSON.stringify({ imageUrl }),
    }),
  delete: (id: number) => request<void>(`/templates/${id}`, { method: "DELETE" }),
};

// ── Groups ───────────────────────────────────────────────────────────────────
export type Contact = { ID: number; Phone: string; Name: string; GroupID: number };
export type ContactGroup = { ID: number; Name: string; Contacts: Contact[] };

export const groups = {
  list: () => request<ContactGroup[]>("/groups"),
  get: (id: number) => request<ContactGroup>(`/groups/${id}`),
  create: (name: string) => request<ContactGroup>("/groups", { method: "POST", body: JSON.stringify({ name }) }),
  delete: (id: number) => request<void>(`/groups/${id}`, { method: "DELETE" }),
  addContact: (groupId: number, phone: string, name: string) =>
    request<Contact>(`/groups/${groupId}/contacts`, {
      method: "POST",
      body: JSON.stringify({ phone, name }),
    }),
  removeContact: (groupId: number, contactId: number) =>
    request<void>(`/groups/${groupId}/contacts/${contactId}`, { method: "DELETE" }),
  importCSV: (groupId: number, file: File) => {
    const form = new FormData();
    form.append("file", file);
    return fetch(`${BASE}/api/groups/${groupId}/import`, {
      method: "POST",
      headers: { Authorization: `Bearer ${getToken()}` },
      body: form,
    }).then((r) => r.json());
  },
};

// ── Campaigns ────────────────────────────────────────────────────────────────
export type Campaign = {
  ID: number;
  CreatedAt: string;
  Name: string;
  TemplateID: number;
  Template: WATemplate;
  GroupID: number;
  Group: ContactGroup;
  ImageURL: string;
  Status: string;
  SentCount: number;
  FailCount: number;
  StartedAt: string | null;
  CompletedAt: string | null;
};

export type CampaignMessage = {
  ID: number;
  Phone: string;
  Status: string;
  WAMessageID: string;
  FailReason: string;
  SentAt: string | null;
  DeliveredAt: string | null;
  ReadAt: string | null;
  Contact: Contact;
};

export const campaigns = {
  list: () => request<Campaign[]>("/campaigns"),
  get: (id: number) => request<Campaign>(`/campaigns/${id}`),
  create: (body: object) => request<Campaign>("/campaigns", { method: "POST", body: JSON.stringify(body) }),
  send: (id: number) => request<{ message: string; contacts: number }>(`/campaigns/${id}/send`, { method: "POST" }),
  messages: (id: number) => request<CampaignMessage[]>(`/campaigns/${id}/messages`),
  delete: (id: number) => request<void>(`/campaigns/${id}`, { method: "DELETE" }),
};
