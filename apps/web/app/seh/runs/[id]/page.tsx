import Link from "next/link";
import { notFound } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { getSEHRun } from "../../../lib/seh-api";
import JsonFold from "../../../components/JsonFold";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function SEHRunDetailPage({ params }: Props) {
  const t = await getTranslations("seh");
  const tc = await getTranslations("common");
  const { id } = await params;

  const run = await getSEHRun(id).catch(() => null);
  if (!run) {
    notFound();
  }

  return (
    <main className="shell">
      <section className="toprail">
        <Link href="/seh/runs" className="toprail-link">← {tc("back")} to {t("runs")}</Link>
      </section>

      <section className="hero-page hero-compact">
        <p className="eyebrow">SEH</p>
        <h1>{run.skill}</h1>
        <p className="lede">
          {run.dataset_id} · {t("score")}: {run.metrics.score.toFixed(2)} · {t("success")}: {(run.metrics.success_rate * 100).toFixed(0)}%
        </p>
      </section>

      <section className="panel-grid panel-grid-wide">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">SEH</p>
            <h2 className="panel-title">{t("runMetrics")}</h2>
          </div>
          <div className="detail-grid">
            <div className="detail-row">
              <span className="detail-label">{tc("id")}</span>
              <span className="detail-value">{run.run_id}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("datasetId")}</span>
              <span className="detail-value">{run.dataset_id}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("runtime")}</span>
              <span className="detail-value">{run.runtime}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("score")}</span>
              <span className="detail-value">{run.metrics.score.toFixed(4)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("success")}</span>
              <span className="detail-value">{(run.metrics.success_rate * 100).toFixed(2)}%</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("avgTokens")}</span>
              <span className="detail-value">{run.metrics.avg_tokens}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("p95Latency")}</span>
              <span className="detail-value">{run.metrics.p95_latency} ms</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("costUsd")}</span>
              <span className="detail-value">${run.metrics.cost_usd.toFixed(4)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("stabilityVariance")}</span>
              <span className="detail-value">{run.metrics.stability_variance.toFixed(4)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("costFactor")}</span>
              <span className="detail-value">{run.metrics.cost_factor.toFixed(4)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("classificationFactor")}</span>
              <span className="detail-value">{run.metrics.classification_factor.toFixed(4)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{tc("createdAt")}</span>
              <span className="detail-value">{new Date(run.created_at).toLocaleString()}</span>
            </div>
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">SEH</p>
            <h2 className="panel-title">{t("caseResults", { count: run.results.length })}</h2>
          </div>

          {run.results.length === 0 ? (
            <p className="empty-note">{t("noRunResults")}</p>
          ) : (
            <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-3)" }}>
              {run.results.map((result) => (
                <div key={result.case_id} style={{ borderBottom: "1px solid var(--line)", paddingBottom: "var(--space-3)" }}>
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "var(--space-2)" }}>
                    <div>
                      <span style={{ fontWeight: 600 }}>{result.case_id.slice(0, 16)}...</span>
                      <span style={{ color: "var(--muted)", fontSize: "var(--text-sm)", marginLeft: "var(--space-3)" }}>
                        {result.classification} · {result.latency_ms}ms · {result.token_usage} tokens
                      </span>
                    </div>
                    <span className={`badge badge-${result.success ? "succeeded" : "failed"}`}>
                      {result.success ? tc("success") : tc("failed")}
                    </span>
                  </div>
                  <JsonFold title={`${t("drillDown")}: ${result.case_id.slice(0, 12)}...`} data={{
                    classification: result.classification,
                    latency_ms: result.latency_ms,
                    token_usage: result.token_usage,
                    output: result.output,
                    error: result.error || null,
                  }} />
                </div>
              ))}
            </div>
          )}
        </article>
      </section>
    </main>
  );
}
