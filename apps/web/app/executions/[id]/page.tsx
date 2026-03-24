import { notFound } from "next/navigation";
import { getExecution } from "../../lib/api";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function ExecutionDetailPage({ params }: Props) {
  const { id } = await params;

  let execution;
  try {
    execution = await getExecution(id);
  } catch {
    notFound();
  }

  if (!execution) {
    notFound();
  }

  return (
    <main className="shell">
      <section className="toprail">
        <a href="/executions">← Back to executions</a>
      </section>

      <section className="hero hero-compact">
        <p className="eyebrow">Execution Timeline</p>
        <h1>{execution.skill_name}</h1>
        <p className="lede">
          {execution.current_step_id} · triggered by {execution.triggered_by} ·{" "}
          <span className={`badge badge-${execution.status}`}>{execution.status}</span>
        </p>
      </section>

      <section className="panel-grid panel-grid-wide">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Execution</p>
            <h2>Run details</h2>
          </div>
          <div className="detail-grid">
            <div className="detail-row">
              <span className="detail-label">ID</span>
              <span className="detail-value">{execution.id}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Skill ID</span>
              <span className="detail-value">{execution.skill_id}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Skill name</span>
              <span className="detail-value">{execution.skill_name}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Status</span>
              <span className="detail-value">
                <span className={`badge badge-${execution.status}`}>{execution.status}</span>
              </span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Triggered by</span>
              <span className="detail-value">{execution.triggered_by}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Started at</span>
              <span className="detail-value">{execution.started_at}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Current step</span>
              <span className="detail-value">{execution.current_step_id}</span>
            </div>
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Execution</p>
            <h2>Input payload</h2>
          </div>
          <pre className="json-block">
            {Object.keys(execution.input || {}).length > 0
              ? JSON.stringify(execution.input, null, 2)
              : "{ }"}
          </pre>
        </article>
      </section>
    </main>
  );
}
