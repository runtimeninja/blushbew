export const dynamic = "force-dynamic";

type Post = {
  title: string;
  excerpt: string;
  content_md: string;
  category: string;
};

export default async function BlogDetail({ params }: { params: { slug: string } }) {
  const apiBase = process.env.NEXT_PUBLIC_API_BASE!;
  const res = await fetch(`${apiBase}/v1/blog/${params.slug}`, { cache: "no-store" });
  if (!res.ok) {
    return (
      <main className="max-w-3xl mx-auto p-6">
        <h1 className="text-2xl font-bold">Not found</h1>
      </main>
    );
  }
  const p: Post = await res.json();

  return (
    <main className="max-w-3xl mx-auto p-6 space-y-4">
      <h1 className="text-3xl font-bold">{p.title}</h1>
      <p className="text-gray-600">{p.excerpt}</p>
      <div className="p-4 border rounded-xl">
        <pre className="whitespace-pre-wrap text-gray-800">{p.content_md}</pre>
      </div>
    </main>
  );
}
