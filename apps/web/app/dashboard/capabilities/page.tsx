import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getCapabilities } from "../../lib/api";
import RadarChart from "../../../components/RadarChart";
import type { RadarChartData } from "../../../components/RadarChart";

function SkeletonCapabilitiesPage() {
  return (
    <main className="shell">
      <section className="hero-page hero-compact">
        <p className="eyebrow">Capabilities</p>
        <h1>Radar Dashboard</h1>
        <p className="lede">
          Multi-dimensional capability visualization with experiment comparison.
        </p>
      </section>

      <div style={{ display: "grid", gridTemplateColumns: "1fr", gap: "18px", marginBottom: "18px" }}>
        <div className="panel" style={{ minHeight: 520 }}>
          <div className="loading-pulse loading-pulse-medium" style={{ width: 200, marginBottom: 12 }} />
          <div className="loading-pulse" style={{ width: "100%", height: 400, marginTop: 24 }} />
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: "18px" }}>
        {[...Array(6)].map((_, i) => (
          <div key={i} className="metric" style={{ minHeight: 100 }}>
            <div className="loading-pulse loading-pulse-short" style={{ width: 100, marginBottom: 12 }} />
            <div className="loading-pulse loading-pulse-medium" style={{ width: 60 }} />
          </div>
        ))}
      </div>
    </main>
  );
}

function ErrorPanel({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <main className="error-shell">
      <div className="error-panel">
        <h2>Failed to load capabilities</h2>
        <p>{message}</p>
        <button onClick={onRetry} className="btn-retry">
          Retry
        </button>
      </div>
    </main>
  );
}

interface CapabilitiesPageProps {
  searchParams: Promise<{ experiments?: string }>;
}

async function CapabilitiesContent({ searchParams }: CapabilitiesPageProps) {
  const t = await getTranslations("dashboard");
  const params = await searchParams;
  const { data: capabilities } = await getCapabilities();

  const uniqueExperiments = Array.from(
    new Map(
      capabilities.flatMap((cap) =>
        cap.scores.map((s) => ({
          id: s.experimentId,
          name: s.experimentName,
        }))
      ).map((exp) => [exp.id, exp])
    ).values()
  );

  const selectedExperimentIds = params.experiments
    ? params.experiments.split(",")
    : uniqueExperiments.length > 0
      ? [uniqueExperiments[0].id]
      : [];

  const selectedExperiments = uniqueExperiments.filter((exp) =>
    selectedExperimentIds.includes(exp.id)
  );

  const colors = ["var(--primary)", "var(--success)", "var(--warning)", "#8b5cf6", "#ec4899", "#06b6d4"];

  const radarData: RadarChartData = {
    capabilities: capabilities.map((cap) => cap.name),
    experiments: selectedExperiments.map((exp, i) => ({
      id: exp.id,
      name: exp.name,
      color: colors[i % colors.length],
      scores: capabilities.map((cap) => {
        const score = cap.scores.find((s) => s.experimentId === exp.id);
        return score?.score ?? 0;
      }),
    })),
  };

  const experimentOptions = uniqueExperiments.map((exp, i) => ({
    ...exp,
    color: colors[i % colors.length],
  }));

  return (
    <main className="shell">
      <section className="hero-page hero-compact">
        <p className="eyebrow">{t("capabilities")}</p>
        <h1>{t("skillHealth")}</h1>
        <p className="lede">
          Multi-dimensional capability visualization with experiment comparison.
        </p>
      </section>

      <div className="panel" style={{ marginBottom: "18px" }}>
        <div style={{ marginBottom: "16px" }}>
          <p className="section-kicker">Compare Experiments</p>
          <div style={{ display: "flex", flexWrap: "wrap", gap: "8px", marginTop: "12px" }}>
            {experimentOptions.map((exp) => {
              const isSelected = selectedExperimentIds.includes(exp.id);
              return (
                <form key={exp.id} method="get" style={{ display: "inline" }}>
                  <input type="hidden" name="experiments" value={exp.id} />
                  <button
                    type="submit"
                    name="experiments"
                    value={isSelected && selectedExperimentIds.length === 1
                      ? ""
                      : [...selectedExperimentIds.filter((id) => id !== exp.id), exp.id].join(",")}
                    style={{
                      display: "inline-flex",
                      alignItems: "center",
                      gap: "6px",
                      padding: "6px 14px",
                      border: `1px solid ${isSelected ? exp.color : "var(--line-strong)"}`,
                      borderRadius: "var(--radius-full)",
                      background: isSelected ? `${exp.color}15` : "rgba(255,255,255,0.5)",
                      color: isSelected ? exp.color : "var(--ink)",
                      fontSize: "0.82rem",
                      fontWeight: 600,
                      cursor: "pointer",
                      transition: "all var(--transition-fast)",
                    }}
                  >
                    <span
                      style={{
                        width: 8,
                        height: 8,
                        borderRadius: "50%",
                        background: exp.color,
                      }}
                    />
                    {exp.name}
                  </button>
                </form>
              );
            })}
          </div>
        </div>

        <div style={{ display: "flex", justifyContent: "center", padding: "24px 0" }}>
          <RadarChart data={radarData} width={480} height={480} />
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))", gap: "18px" }}>
        {capabilities.map((cap) => {
          const latestScore = cap.scores.find(
            (s) => s.experimentId === selectedExperimentIds[0]
          );
          const scoreValue = latestScore?.score ?? cap.score;
          const scorePercent = Math.round(scoreValue);
          const scoreClass =
            scorePercent >= 80
              ? "score-fill-high"
              : scorePercent >= 50
                ? "score-fill-medium"
                : "score-fill-low";

          return (
            <article key={cap.name} className="panel">
              <p className="section-kicker">{cap.name}</p>
              <strong style={{ display: "block", fontSize: "2rem", color: "var(--ink-strong)", marginTop: "8px" }}>
                {scoreValue.toFixed(1)}
              </strong>
              <div className="score-bar">
                <div
                  className={`score-fill ${scoreClass}`}
                  style={{ width: `${scorePercent}%` }}
                />
              </div>
              <p style={{ margin: "12px 0 0", fontSize: "0.8rem", color: "var(--muted)" }}>
                Sample size: {cap.sampleSize.toLocaleString()}
              </p>
            </article>
          );
        })}
      </div>
    </main>
  );
}

export default async function CapabilitiesPage({ searchParams }: CapabilitiesPageProps) {
  return (
    <Suspense fallback={<SkeletonCapabilitiesPage />}>
      <CapabilitiesContent searchParams={searchParams} />
    </Suspense>
  );
}
