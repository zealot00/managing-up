"use client";

import Link from "next/link";
import { useState, useMemo } from "react";
import { useTranslations } from "next-intl";

type Run = { run_id: string; skill: string; dataset_id: string; metrics: { score: number; success_rate: number } };

type Props = {
  runs: Run[];
  total: number;
  hasMore: boolean;
};

const PAGE_SIZE = 20;

export default function SEHRunsList({ runs, total, hasMore }: Props) {
  const t = useTranslations("seh");
  const tc = useTranslations("common");
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);

  const filtered = useMemo(() => {
    if (!searchQuery) return runs;
    const q = searchQuery.toLowerCase();
    return runs.filter(r =>
      r.skill.toLowerCase().includes(q) ||
      r.run_id.toLowerCase().includes(q) ||
      r.dataset_id.toLowerCase().includes(q)
    );
  }, [runs, searchQuery]);

  const totalPages = Math.ceil(filtered.length / PAGE_SIZE);
  const start = (page - 1) * PAGE_SIZE;
  const paged = filtered.slice(start, start + PAGE_SIZE);

  return (
    <main className="shell">
      <section className="toprail">
        <Link href="/seh">{tc("back")} to SEH</Link>
      </section>

      <header className="hero-page hero-compact">
        <p className="eyebrow">SEH</p>
        <h1>{t("runs")}</h1>
        <p className="lede">{t("runsPageLede")}</p>
      </header>

      <section className="panel">
        <div className="panel-header">
          <p className="section-kicker">SEH</p>
          <h2 className="panel-title">{t("allRuns", { count: filtered.length })}</h2>
        </div>

        <div style={{ padding: "var(--space-4)" }}>
          <div className="search-bar">
            <input
              type="text"
              placeholder={t("searchPlaceholder")}
              value={searchQuery}
              onChange={(e) => { setSearchQuery(e.target.value); setPage(1); }}
            />
          </div>

          <div className="list">
            {paged.length === 0 ? (
              <p className="empty-note">{searchQuery ? "No matching runs" : t("noRuns")}</p>
            ) : (
              paged.map((run) => (
                <Link
                  key={run.run_id}
                  href={`/seh/runs/${run.run_id}`}
                  className="list-card"
                  style={{ textDecoration: "none" }}
                >
                  <div className="list-card-main">
                    <h3 className="list-card-title">{run.skill}</h3>
                    <p className="list-card-meta">
                      {t("score")}: {run.metrics.score.toFixed(2)} · {t("success")}: {(run.metrics.success_rate * 100).toFixed(0)}%
                    </p>
                  </div>
                  <span className={`badge badge-${run.metrics.score >= 0.75 ? "succeeded" : "failed"}`}>
                    {run.run_id}
                  </span>
                </Link>
              ))
            )}
          </div>

          {totalPages > 1 && (
            <div className="pagination-bar">
              <span>{t("showing", { from: start + 1, to: Math.min(start + PAGE_SIZE, filtered.length), total: filtered.length })}</span>
              <div style={{ display: "flex", gap: "var(--space-2)" }}>
                <button
                  className="pagination-btn"
                  onClick={() => setPage(p => Math.max(1, p - 1))}
                  disabled={page === 1}
                >
                  {t("prev")}
                </button>
                <button
                  className="pagination-btn"
                  onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                >
                  {t("next")}
                </button>
              </div>
            </div>
          )}
        </div>
      </section>
    </main>
  );
}
