export const dynamic = "force-dynamic";

type PostListItem = {
  slug: string;
  title: string;
  excerpt: string;
  category: string;
  published_at?: string;
};

export default async function Blog() {
  const apiBase = process.env.NEXT_PUBLIC_API_BASE!;
  const res = await fetch(`${apiBase}/v1/blog`, { cache: "no-store" });
  const data = await res.json();

  const posts: PostListItem[] = data.posts || [];

  return (
    <main className="max-w-3xl mx-auto p-6 space-y-6">
      <h1 className="text-2xl font-bold">Blog</h1>

      <div className="space-y-4">
        {posts.length === 0 && <p className="text-gray-600">No posts yet.</p>}
        {posts.map((p) => (
          <a key={p.slug} className="block p-4 border rounded-xl hover:bg-gray-50" href={`/blog/${p.slug}`}>
            <div className="flex items-center justify-between gap-3">
              <h2 className="font-semibold">{p.title}</h2>
              <span className="text-xs text-gray-500">{p.category}</span>
            </div>
            <p className="text-gray-600 mt-2">{p.excerpt}</p>
          </a>
        ))}
      </div>
    </main>
  );
}
