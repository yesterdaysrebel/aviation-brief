import { useEffect, useState } from "react";

function formatTime(iso) {
  try {
    return new Intl.DateTimeFormat(undefined, {
      dateStyle: "medium",
      timeStyle: "short",
    }).format(new Date(iso));
  } catch {
    return iso;
  }
}

export default function App() {
  const [headlines, setHeadlines] = useState([]);
  const [wx, setWx] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const [hRes, wRes] = await Promise.all([
          fetch("/api/headlines"),
          fetch("/api/metar?icao=KSFO"),
        ]);
        if (!hRes.ok) throw new Error("Headlines request failed");
        if (!wRes.ok) throw new Error("Weather request failed");
        const [hData, wData] = await Promise.all([hRes.json(), wRes.json()]);
        if (!cancelled) {
          setHeadlines(hData.items ?? []);
          setWx(wData);
        }
      } catch (e) {
        if (!cancelled) setError(e.message ?? "Something went wrong");
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="shell">
      <header className="hero">
        <div className="hero__badge">live brief</div>
        <h1 className="hero__title">Aviation Brief</h1>
        <p className="hero__sub">
          Headlines and a sample METAR-style readout. One page, dark glass UI,
          Go APIs behind the scenes.
        </p>
      </header>

      <section className="panel wx" aria-label="Sample weather">
        <div className="panel__head">
          <h2>Weather snapshot</h2>
          <span className="icao">{wx?.icao ?? "—"}</span>
        </div>
        {loading && <p className="muted">Loading…</p>}
        {error && <p className="err">{error}</p>}
        {!loading && !error && wx && (
          <div className="wx__body">
            <p className="wx__raw">{wx.raw}</p>
            <dl className="wx__meta">
              <div>
                <dt>Wind</dt>
                <dd>{wx.wind}</dd>
              </div>
              <div>
                <dt>Visibility</dt>
                <dd>{wx.visibility}</dd>
              </div>
              <div>
                <dt>Sky</dt>
                <dd>{wx.sky}</dd>
              </div>
            </dl>
          </div>
        )}
      </section>

      <section className="panel feed" aria-label="Headlines">
        <div className="panel__head">
          <h2>Headlines</h2>
          <span className="muted small">Aviation & ops</span>
        </div>
        {loading && <p className="muted">Loading…</p>}
        {!loading && !error && (
          <ul className="cards">
            {headlines.map((item) => (
              <li key={item.id} className="card">
                <div className="card__top">
                  <span className="card__src">{item.source}</span>
                  <time dateTime={item.publishedAt}>
                    {formatTime(item.publishedAt)}
                  </time>
                </div>
                <h3 className="card__title">{item.title}</h3>
                <p className="card__sum">{item.summary}</p>
              </li>
            ))}
          </ul>
        )}
      </section>

      <footer className="foot">
        <span>Aviation Brief</span>
        <span className="muted small">Kubernetes-ready · GHCR-friendly</span>
      </footer>
    </div>
  );
}
