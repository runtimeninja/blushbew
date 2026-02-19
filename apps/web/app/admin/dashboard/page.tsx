"use client";

import { useEffect, useState } from "react";

type Post = {
  id: number;
  slug: string;
  title: string;
  excerpt: string;
  content_md: string;
  category: string;
  status: string;
};

export default function Dashboard() {
  const apiBase = process.env.NEXT_PUBLIC_API_BASE!;
  const [posts, setPosts] = useState<Post[]>([]);
  const [err, setErr] = useState<string | null>(null);

  const [slug, setSlug] = useState("");
  const [title, setTitle] = useState("");
  const [excerpt, setExcerpt] = useState("");
  const [category, setCategory] = useState("general");
  const [content, setContent] = useState("");
  const [status, setStatus] = useState("draft");

  async function load() {
    setErr(null);
    const res = await fetch(`${apiBase}/v1/admin/posts`, { credentials: "include" });
    const data = await res.json();
    if (!res.ok) {
      setErr(data?.error || "unauthorized");
      return;
    }
    setPosts(data.posts || []);
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function create() {
    setErr(null);
    const res = await fetch(`${apiBase}/v1/admin/posts`, {
      method: "POST",
      credentials: "include",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        slug,
        title,
        excerpt,
        category,
        content_md: content,
        status,
      }),
    });
    const data = await res.json();
    if (!res.ok) {
      setErr(data?.error || "create_failed");
      return;
    }
    setSlug(""); setTitle(""); setExcerpt(""); setCategory("general"); setContent(""); setStatus("draft");
    await load();
  }

  async function publish(id: number) {
    setErr(null);
    const res = await fetch(`${apiBase}/v1/admin/posts/${id}/publish`, {
      method: "POST",
      credentials: "include",
    });
    const data = await res.json();
    if (!res.ok) {
      setErr(data?.error || "publish_failed");
      return;
    }
    await load();
  }

  return (
    <main className="max-w-4xl mx-auto p-6 space-y-6">
      <h1 className="text-2xl font-bold">Admin Dashboard</h1>

      {err && <p className="text-sm text-red-600">Error: {err}</p>}

      <section className="p-4 border rounded-xl space-y-3">
        <h2 className="font-semibold">Create new post</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <input className="border rounded-lg p-2" placeholder="slug (unique)" value={slug} onChange={(e) => setSlug(e.target.value)} />
          <input className="border rounded-lg p-2" placeholder="title" value={title} onChange={(e) => setTitle(e.target.value)} />
        </div>
        <input className="w-full border rounded-lg p-2" placeholder="excerpt" value={excerpt} onChange={(e) => setExcerpt(e.target.value)} />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <input className="border rounded-lg p-2" placeholder="category" value={category} onChange={(e) => setCategory(e.target.value)} />
          <select className="border rounded-lg p-2" value={status} onChange={(e) => setStatus(e.target.value)}>
            <option value="draft">draft</option>
            <option value="published">published</option>
          </select>
        </div>
        <textarea className="w-full border rounded-lg p-2 min-h-[150px]" placeholder="content (markdown)" value={content} onChange={(e) => setContent(e.target.value)} />
        <button className="px-4 py-2 rounded-lg border" onClick={create}>Create</button>
      </section>

      <section className="p-4 border rounded-xl space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="font-semibold">Posts</h2>
          <button className="px-3 py-1 rounded-lg border" onClick={load}>Refresh</button>
        </div>

        <div className="space-y-2">
          {posts.map((p) => (
            <div key={p.id} className="p-3 border rounded-lg flex items-center justify-between gap-3">
              <div className="min-w-0">
                <div className="font-medium truncate">{p.title}</div>
                <div className="text-xs text-gray-500 truncate">{p.slug} • {p.status}</div>
              </div>
              <div className="flex gap-2">
                {p.status !== "published" && (
                  <button className="px-3 py-1 rounded-lg border" onClick={() => publish(p.id)}>
                    Publish
                  </button>
                )}
                <a className="px-3 py-1 rounded-lg border" href={`/blog/${p.slug}`} target="_blank">
                  View
                </a>
              </div>
            </div>
          ))}
        </div>
      </section>
    </main>
  );
}
