import { Suspense } from "react";
import { getExecution, getTraces } from "../../../lib/api";
import type { Execution, TraceEvent } from "../../../lib/api";
import { SkeletonPanel } from "../../../components/SkeletonPanel";

type Props = {
  params: Promise<{ id: string }>;
};

function SkeletonTracePage() {
  return (
    <main className="shell">
      <section className="toprail">
        <div className="loading-pulse" style={{ width: 160, height: 44, borderRadius: 999 }} />
      </section>
      <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 8 }} />
      <div className="loading-pulse loading-pulse-short" />
      <div className="skeleton-grid" style={{ marginTop: 32 }}>
        {[1, 2, 3, 4].map((i) => (
          <div className="skeleton-card" key={i} />
        ))}
      </div>
    </main>
  );
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
          <p style={{ margin: "6px 0 0", fontSize: "0.82rem", color: "var(--muted)" }}>
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
  let execution: Execution | null = null;
  let traces: TraceEvent[] = [];

  try {
    execution = await getExecution(id);
    traces = await getTraces(id);
  } catch {
    execution = null;
  }

  return (
    <main className="shell">
      <section className="toprail">
        <a className="toprail-link" href="/executions">
          ← Executions
        </a>
        <a className="toprail-link" href="/">
          Dashboard
        </a>
      </section>

      {execution ? (
        <>
          <section className="hero-page hero-compact" style={{ marginBottom: 24 }}>
            <p className="eyebrow">Execution Trace</p>
            <h1>{execution.skill_name}</h1>
            <div style={{ marginTop: 12, display: "flex", gap: 12, alignItems: "center" }}>
              <span className={`badge badge-${execution.status}`}>{execution.status}</span>
              <span style={{ color: "var(--muted)", fontSize: "0.85rem" }}>
                ID: {execution.id}
              </span>
            </div>
          </section>

          <article className="panel">
            <div className="panel-header">
              <p className="section-kicker">Timeline</p>
              <h2>Execution Events</h2>
            </div>
            {traces.length > 0 ? (
              <div className="trace-timeline">
                {traces.map((event) => (
                  <TraceEventCard key={event.id} event={event} />
                ))}
              </div>
            ) : (
              <p style={{ color: "var(--muted)", marginTop: 16 }}>
                No trace events recorded yet. Traces are captured when execution tracing is enabled.
              </p>
            )}
          </article>

          <article className="panel" style={{ marginTop: 18 }}>
            <div className="panel-header">
              <h2>Execution Details</h2>
            </div>
            <div className="detail-grid">
              <div className="detail-row">
                <span className="detail-label">Execution ID</span>
                <span className="detail-value">{execution.id}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Skill ID</span>
                <span className="detail-value">{execution.skill_id}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Status</span>
                <span className="detail-value">{execution.status}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Triggered By</span>
                <span className="detail-value">{execution.triggered_by}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Started At</span>
                <span className="detail-value">{new Date(execution.started_at).toLocaleString()}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Current Step</span>
                <span className="detail-value">{execution.current_step_id}</span>
              </div>
            </div>
          </article>
        </>
      ) : (
        <article className="panel" style={{ marginTop: 24 }}>
          <h2 style={{ color: "var(--ink-strong)" }}>Execution not found</h2>
          <p style={{ color: "var(--muted)", marginTop: 8 }}>
            Could not load execution data. Make sure the backend is running.
          </p>
        </article>
      )}
    </main>
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
