import { notFound } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { getExecution, getTraces } from "../../lib/api";
import { ExecutionDetailHeader } from "./ExecutionDetailHeader";
import Breadcrumb from "../../../components/Breadcrumb";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function ExecutionDetailPage({ params }: Props) {
  const t = await getTranslations("executions");
  const tc = await getTranslations("common");
  const { id } = await params;

  let execution;
  let traces;
  try {
    [execution, traces] = await Promise.all([getExecution(id), getTraces(id)]);
  } catch {
    notFound();
  }

  if (!execution) {
    notFound();
  }

  const formatTimestamp = (ts: string) => {
    return new Date(ts).toLocaleString();
  };

  return (
    <>
      <Breadcrumb />
      <ExecutionDetailHeader
        eyebrow={t("eyebrow")}
        execution={execution}
      />

      <section className="panel-grid panel-grid-wide">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("runs")}</p>
            <h2>{t("runDetails")}</h2>
          </div>
          <div className="detail-grid">
            <div className="detail-row">
              <span className="detail-label">{tc("id")}</span>
              <span className="detail-value">{execution.id}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("skillId")}</span>
              <span className="detail-value">{execution.skill_id}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("skillName")}</span>
              <span className="detail-value">{execution.skill_name}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{tc("status")}</span>
              <span className="detail-value">
                <span className={`badge badge-${execution.status}`}>{execution.status}</span>
              </span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("triggeredBy")}</span>
              <span className="detail-value">{execution.triggered_by}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("startedAt")}</span>
              <span className="detail-value">{execution.started_at}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("currentStep")}</span>
              <span className="detail-value">{execution.current_step_id}</span>
            </div>
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("runs")}</p>
            <h2>{t("inputPayload")}</h2>
          </div>
          <pre className="json-block">
            {Object.keys(execution.input || {}).length > 0
              ? JSON.stringify(execution.input, null, 2)
              : "{ }"}
          </pre>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("runs")}</p>
            <h2>{t("traceTimeline")}</h2>
          </div>
          {traces.length === 0 ? (
            <p className="text-muted">{t("noTraceEvents")}</p>
          ) : (
            <div className="list">
              {traces.map((event) => (
                <article className="list-card" key={event.id}>
                  <div>
                    <h3>{event.step_id}</h3>
                    <p>
                      <span className={`badge badge-${event.event_type}`}>{event.event_type}</span>
                      {" · "}
                      {formatTimestamp(event.timestamp)}
                    </p>
                    <details>
                      <summary>{t("eventData")}</summary>
                      <pre className="json-block json-small">
                        {JSON.stringify(event.event_data, null, 2)}
                      </pre>
                    </details>
                  </div>
                </article>
              ))}
            </div>
          )}
        </article>
      </section>
    </>
  );
}