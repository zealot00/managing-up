import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getExecution, getTraces } from "../../../lib/api";
import type { Execution, TraceEvent } from "../../../lib/api";
import { PageSkeleton } from "../../../components/layout/Skeleton";
import Breadcrumb from "../../../../components/Breadcrumb";
import { PageHeader } from "../../../components/layout/PageHeader";

type Props = {
  params: Promise<{ id: string }>;
};

function SkeletonTracePage() {
  return <PageSkeleton headerActions={false} content="cards" contentCount={4} />;
}

function getMarkerClass(eventType: string): string {
  switch (eventType) {
    case "execution_started":
      return "trace-marker trace-marker-start";
    case "execution_succeeded":
      return "trace-marker trace-marker-end";
    case "execution_failed":
      return "trace-marker trace-marker-error";
    case "step_started":
    case "step_succeeded":
    case "step_failed":
      return "trace-marker trace-marker-step";
    case "approval_requested":
    case "approval_resolved":
      return "trace-marker trace-marker-approval";
    default:
      return "trace-marker";
  }
}

function formatEventType(eventType: string): string {
  return eventType.replace(/_/g, " ");
}

function formatTime(timestamp: string): string {
  try {
    const date = new Date(timestamp);
    return date.toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    });
  } catch {
    return timestamp;
  }
}

function TraceEventCard({ event }: { event: TraceEvent }) {
  const eventData = event.event_data || {};

  return (
    <div className="trace-event">
      <div className={getMarkerClass(event.event_type)} />
      <div className="trace-content">
        <p className="trace-type">{formatEventType(event.event_type)}</p>
        <p className="trace-time">{formatTime(event.timestamp)}</p>
        {event.step_id && (
          <p className="text-muted" style={{ margin: "6px 0 0", fontSize: "0.82rem" }}>
            Step: {event.step_id}
          </p>
        )}
        {Object.keys(eventData).length > 0 && (
          <div className="trace-data">
            {Object.entries(eventData).slice(0, 4).map(([key, value]) => (
              <div key={key} style={{ marginBottom: 4 }}>
                <span style={{ color: "var(--primary)", fontWeight: 600 }}>{key}:</span>{" "}
                {typeof value === "object" ? JSON.stringify(value) : String(value)}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

async function TraceContent({ id }: { id: string }) {
  const t = await getTranslations("executions");
  const tc = await getTranslations("common");
  let execution: Execution | null = null;
  let traces: TraceEvent[] = [];

  try {
    execution = await getExecution(id);
    traces = await getTraces(id);
  } catch {
    execution = null;
  }

  return (
    <>
      <Breadcrumb />

      {execution ? (
        <>
          <header className="detail-header">
            <div className="detail-header-main">
              <h1 className="detail-header-title">{execution.skill_name}</h1>
              <span className={`badge badge-${execution.status}`}>{execution.status}</span>
            </div>
            <div className="detail-header-chips">
              <span className="detail-chip">
                <span className="detail-chip-dot" style={{ background: "var(--muted)" }} aria-hidden="true" />
                <span style={{ fontFamily: "monospace", fontSize: "var(--text-xs)" }}>{execution.id.slice(0, 8)}…</span>
              </span>
            </div>
          </header>

          <article className="panel">
            <div className="panel-header">
              <p className="section-kicker">{t("traceTimeline")}</p>
              <h2>{t("runDetails")}</h2>
            </div>
            {traces.length > 0 ? (
              <div className="trace-timeline">
                {traces.map((event) => (
                  <TraceEventCard key={event.id} event={event} />
                ))}
              </div>
            ) : (
              <p className="empty-note">{t("noTraceEvents")}</p>
            )}
          </article>

          <article className="panel">
            <div className="panel-header">
              <h2>{t("runDetails")}</h2>
            </div>
            <div className="detail-grid">
              <div className="detail-row">
                <span className="detail-label">{t("skillId")}</span>
                <span className="detail-value">{execution.skill_id}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">{tc("status")}</span>
                <span className="detail-value">{execution.status}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">{t("triggeredBy")}</span>
                <span className="detail-value">{execution.triggered_by}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">{t("startedAt")}</span>
                <span className="detail-value">{new Date(execution.started_at).toLocaleString()}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">{t("currentStep")}</span>
                <span className="detail-value">{execution.current_step_id}</span>
              </div>
            </div>
          </article>
        </>
      ) : (
        <article className="panel">
          <h2>{tc("notFound")}</h2>
          <p className="empty-note">{t("noTraceEvents")}</p>
        </article>
      )}
    </>
  );
}

export default async function ExecutionTracePage({ params }: Props) {
  const { id } = await params;
  return (
    <Suspense fallback={<SkeletonTracePage />}>
      <TraceContent id={id} />
    </Suspense>
  );
}