import { refresh } from "next/cache";
import { connection } from "next/server";

type Site = {
  id: number;
  url: string;
  up: boolean | null;
  status_code: number | null;
  latency_ms: number | null;
  checked_at: string | null;
};

// Live status page: render on every request, wait for the backend (no static shell).
export const instant = false;

// Read at request time on the server, so one image works in every environment.
const api = () => process.env.API_URL ?? "http://localhost:8080";

async function addSite(formData: FormData) {
  "use server";
  await fetch(`${api()}/sites`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ url: formData.get("url") }),
  });
  refresh();
}

export default async function Home() {
  await connection();
  let sites: Site[] = [];
  let error = "";
  try {
    const res = await fetch(`${api()}/sites`, { cache: "no-store" });
    if (!res.ok) throw new Error(`backend returned ${res.status}`);
    sites = await res.json();
  } catch (e) {
    error = `Cannot reach backend at ${api()}: ${(e as Error).message}`;
  }

  return (
    <main>
      <meta httpEquiv="refresh" content="30" />
      <h1>Uptime Watcher</h1>

      <form action={addSite}>
        <input name="url" type="url" placeholder="https://example.com" required />
        <button type="submit">Add site</button>
      </form>

      {error && <p className="error">{error}</p>}

      <ul>
        {sites.map((s) => (
          <li key={s.id}>
            <span className={`dot ${s.up === null ? "unknown" : s.up ? "up" : "down"}`} />
            <span className="url">{s.url}</span>
            <span className="meta">
              {s.checked_at
                ? `${s.status_code ?? "no response"} · ${s.latency_ms} ms · ${new Date(s.checked_at).toLocaleTimeString()}`
                : "not checked yet"}
            </span>
          </li>
        ))}
      </ul>
    </main>
  );
}
