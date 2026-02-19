"use client";

import { useState } from "react";

type Output = {
  likely_causes: string[];
  fix_steps: string[];
  do: string[];
  dont: string[];
  suggested_product_types: string[];
};

export default function FixMyMakeup() {
  const apiBase = process.env.NEXT_PUBLIC_API_BASE!;
  const [problemText, setProblemText] = useState("");
  const [skinType, setSkinType] = useState("");
  const [climate, setClimate] = useState("");
  const [eventType, setEventType] = useState("");
  const [loading, setLoading] = useState(false);
  const [out, setOut] = useState<Output | null>(null);
  const [err, setErr] = useState<string | null>(null);

  async function submit() {
    setErr(null);
    setOut(null);
    setLoading(true);
    try {
      const res = await fetch(`${apiBase}/v1/diagnosis`, {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({
          problem_text: problemText,
          skin_type: skinType,
          climate,
          event_type: eventType,
        }),
      });
      const data = await res.json();
      if (!res.ok) {
        setErr(data?.error || "request_failed");
        return;
      }
      setOut(data);
    } catch (e: any) {
      setErr("network_error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="max-w-3xl mx-auto p-6 space-y-6">
      <h1 className="text-2xl font-bold">Fix My Makeup (Free AI)</h1>

      <div className="p-4 border rounded-xl space-y-3">
        <label className="block text-sm font-medium">Describe your problem</label>
        <textarea
          className="w-full border rounded-lg p-3 min-h-[110px]"
          placeholder='Example: "My foundation melts after 2 hours in humid weather"'
          value={problemText}
          onChange={(e) => setProblemText(e.target.value)}
        />

        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          <input
            className="border rounded-lg p-2"
            placeholder="Skin type (optional)"
            value={skinType}
            onChange={(e) => setSkinType(e.target.value)}
          />
          <input
            className="border rounded-lg p-2"
            placeholder="Climate (optional)"
            value={climate}
            onChange={(e) => setClimate(e.target.value)}
          />
          <input
            className="border rounded-lg p-2"
            placeholder="Event (optional)"
            value={eventType}
            onChange={(e) => setEventType(e.target.value)}
          />
        </div>

        <button
          className="px-4 py-2 rounded-lg border disabled:opacity-50"
          disabled={loading || problemText.trim().length < 5}
          onClick={submit}
        >
          {loading ? "Analyzing..." : "Analyze"}
        </button>

        {err && <p className="text-sm text-red-600">Error: {err}</p>}
      </div>

      {out && (
        <div className="space-y-4">
          <Section title="Likely causes" items={out.likely_causes} />
          <Section title="Fix steps" items={out.fix_steps} />
          <Section title="Do" items={out.do} />
          <Section title="Don't" items={out.dont} />
          <Section title="Suggested product types" items={out.suggested_product_types} />
        </div>
      )}

      <p className="text-sm text-gray-500">
        Disclaimer: This is general guidance, not medical advice. Patch test products.
      </p>
    </main>
  );
}

function Section({ title, items }: { title: string; items: string[] }) {
  return (
    <section className="p-4 border rounded-xl space-y-2">
      <h2 className="font-semibold">{title}</h2>
      <ul className="list-disc ml-5 space-y-1">
        {items.map((x, i) => (
          <li key={i} className="text-gray-700">{x}</li>
        ))}
      </ul>
    </section>
  );
}
