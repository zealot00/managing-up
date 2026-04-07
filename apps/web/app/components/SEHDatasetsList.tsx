"use client";

import Link from "next/link";
import { useState, useMemo } from "react";
import { useTranslations } from "next-intl";

type Dataset = { dataset_id: string; name: string; version: string; owner: string; case_count: number };

type Props = {
  datasets: Dataset[];
  total: number;
  hasMore: boolean;
};

const PAGE_SIZE = 20;

export default function SEHDatasetsList({ datasets, total, hasMore }: Props) {
  const t = useTranslations("seh");
  const tc = useTranslations("common");
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);

  const filtered = useMemo(() => {
    if (!searchQuery) return datasets;
    const q = searchQuery.toLowerCase();
    return datasets.filter(d =>
      d.name.toLowerCase().includes(q) ||
      d.owner.toLowerCase().includes(q) ||
      d.dataset_id.toLowerCase().includes(q)
    );
  }, [datasets, searchQuery]);

  const totalPages = Math.ceil(filtered.length / PAGE_SIZE);
  const start = (page - 1) * PAGE_SIZE;
  const paged = filtered.slice(start, start + PAGE_SIZE);

  return (
    <main className="shell">
      <section className="toprail">
        <Link href="/seh" className="toprail-link">← {tc("back")} to SEH</Link>
      </section>

      <header className="hero-page hero-compact">
        <p className="eyebrow">SEH</p>
        <h1>{t("datasets")}</h1>
        <p className="lede">{t("datasetsPageLede")}</p>
      </header>

      <section className="panel">
        <div className="panel-header">
          <p className="section-kicker">SEH</p>
          <h2 className="panel-title">{t("allDatasets", { count: filtered.length })}</h2>
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

          <div className="table-wrapper">
            {paged.length === 0 ? (
              <p className="empty-note">{searchQuery ? "No matching datasets" : t("noDatasets")}</p>
            ) : (
              <table className="table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Version</th>
                    <th>Owner</th>
                    <th>Cases</th>
                    <th>Dataset ID</th>
                  </tr>
                </thead>
                <tbody>
                  {paged.map((dataset) => (
                    <tr
                      key={dataset.dataset_id}
                      onClick={() => window.location.href = `/seh/datasets/${dataset.dataset_id}`}
                      style={{ cursor: "pointer" }}
                    >
                      <td>{dataset.name}</td>
                      <td>{dataset.version}</td>
                      <td>{dataset.owner}</td>
                      <td>{dataset.case_count}</td>
                      <td>
                        <span className="badge badge-muted">{dataset.dataset_id}</span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
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
