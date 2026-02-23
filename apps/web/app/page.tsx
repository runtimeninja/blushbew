import Link from "next/link";

export default function Home() {
  return (
    <main className="max-w-3xl mx-auto p-6 space-y-6">
      <header className="space-y-2">
        <h1 className="text-3xl font-bold">Blushbew</h1>
        <p className="text-gray-600">
          Daily beauty tips + a free AI tool to debug why your makeup breaks.
        </p>
      </header>

      <section className="p-4 border rounded-xl space-y-3">
        <h2 className="text-xl font-semibold">Try the AI Makeup Debugger</h2>
        <p className="text-gray-600">
          Example: “My foundation cracks after 2 hours”
        </p>
        <link className="inline-block px-4 py-2 rounded-lg border" href="/tool/fix-my-makeup">
          Open Tool →
        </link>
      </section>

      <section className="p-4 border rounded-xl space-y-3">
        <h2 className="text-xl font-semibold">Read the latest posts</h2>
        <link className="inline-block px-4 py-2 rounded-lg border" href="/blog">
          Go to Blog →
        </link>
      </section>

      <footer className="text-sm text-gray-500">
        Not medical advice. Patch test products.
      </footer>
    </main>
  );
}
