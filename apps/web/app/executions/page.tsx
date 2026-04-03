import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getExecutions, getSkills } from "../lib/api";
import TriggerExecutionForm from "../components/TriggerExecutionForm";

async function ExecutionsContent() {
  const t = await getTranslations("executions");
  const tc = await getTranslations("common");
  const [executions, skills] = await Promise.all([getExecutions(), getSkills()]);

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">{t("eyebrow")}</p>
        <h1>{t("title")}</h1>
        <p className="lede">
          {t("lede")}
        </p>
      </header>

      <TriggerExecutionForm skills={skills.items} />

      <section className="panel">
        <div className="panel-header">
          <p className="section-kicker">{t("runs")}</p>
          <h2 className="panel-title">{t("executionQueue")}</h2>
        </div>
        <div className="list">
          {executions.items.length === 0 ? (
            <p className="empty-note">{t("noExecutions")}</p>
          ) : (
            executions.items.map((execution) => (
              <article className="list-card" key={execution.id}>
                <div className="list-card-main">
                  <h3 className="list-card-title">{execution.skill_name}</h3>
                  <p className="list-card-meta">
                    {execution.current_step_id} · {t("triggeredBy")} {execution.triggered_by}
                  </p>
                </div>
                <div className="list-card-actions">
                  <a href={`/executions/${execution.id}`} className="trace-link">
                    {t("viewTrace")}
                  </a>
                  <span className={`badge badge-${execution.status}`}>{execution.status}</span>
                </div>
              </article>
            ))
          )}
        </div>
      </section>
    </main>
  );
}

function SkeletonExecutionsPage() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Execution Timeline</p>
        <h1>Operational runs across governed skills.</h1>
        <p className="lede">
          Monitor live and completed runs, current step progression, and operator-triggered events.
        </p>
      </header>

      <div className="form-panel">
        <div className="loading-pulse loading-pulse-short" style={{ marginBottom: 16 }} />
        <div className="form-fields">
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-long" />
        </div>
      </div>

      <div className="panel">
        <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
        <div className="skeleton-grid">
          {[1, 2, 3, 4, 5].map((i) => <div key={i} className="skeleton-card" />)}
        </div>
      </div>
    </main>
  );
}

export default function ExecutionsPage() {
  return (
    <Suspense fallback={<SkeletonExecutionsPage />}>
      <ExecutionsContent />
    </Suspense>
  );
}