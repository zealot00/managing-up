"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { TaskExecution, Task, Metric, runTaskEvaluation, createMetric } from "../lib/api";
import RunEvaluationForm from "./RunEvaluationForm";
import CreateMetricForm from "./CreateMetricForm";

type Props = {
  executions: TaskExecution[];
  tasks: Task[];
  metrics: Metric[];
};

export default function EvaluationManager({ executions, tasks, metrics }: Props) {
  const [showRunEval, setShowRunEval] = useState(false);
  const [showCreateMetric, setShowCreateMetric] = useState(false);

  return (
    <>
      <section aria-label="Available metrics" style={{ marginTop: "var(--space-6)" }}>
        <div className="page-header" style={{ marginBottom: "var(--space-3)", paddingBottom: 0, borderBottom: "none" }}>
          <div className="page-header-content">
            <h2 className="section-kicker" style={{ margin: 0 }}>
              Available Metrics ({metrics.length})
            </h2>
          </div>
          <div className="page-header-actions">
            <button
              className="btn btn-sm btn-secondary"
              onClick={() => setShowCreateMetric(!showCreateMetric)}
            >
              {showCreateMetric ? "Cancel" : "+ New Metric"}
            </button>
          </div>
        </div>

        {showCreateMetric && <CreateMetricForm onCreated={() => setShowCreateMetric(false)} />}

        <div className="tags">
          {metrics.map((metric) => (
            <span key={metric.id} className="tag">
              {metric.name} ({metric.type})
            </span>
          ))}
          {metrics.length === 0 && !showCreateMetric && (
            <span style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
              No metrics defined.
            </span>
          )}
        </div>
      </section>

      <section aria-label="Task executions" style={{ marginTop: "var(--space-8)" }}>
        <div className="page-header" style={{ marginBottom: "var(--space-3)", paddingBottom: 0, borderBottom: "none" }}>
          <div className="page-header-content">
            <h2 className="section-kicker" style={{ margin: 0 }}>
              Task Executions ({executions.length})
            </h2>
          </div>
          <div className="page-header-actions">
            <button
              className="btn btn-sm btn-primary"
              onClick={() => setShowRunEval(!showRunEval)}
            >
              {showRunEval ? "Cancel" : "Run Evaluation"}
            </button>
          </div>
        </div>

        {showRunEval && <RunEvaluationForm tasks={tasks} onCreated={() => setShowRunEval(false)} />}

        {executions.length > 0 ? (
          <div className="eval-grid">
            {executions.map((exec) => (
              <ExecutionCard key={exec.id} exec={exec} tasks={tasks} />
            ))}
          </div>
        ) : (
          <div className="empty-state">
            <div className="empty-state-icon">◎</div>
            <h3 className="empty-state-title">No executions yet</h3>
            <p className="empty-state-description">
              Run an evaluation to see agent performance metrics here.
            </p>
          </div>
        )}
      </section>

      <section aria-label="Task overview" style={{ marginTop: "var(--space-8)" }}>
        <h2 className="section-kicker" style={{ marginBottom: "var(--space-3)" }}>
          Task Overview ({tasks.length} tasks)
        </h2>
        {tasks.length > 0 ? (
          <div className="panel">
            <div className="list">
              {tasks.map((task) => (
                <article className="list-card" key={task.id}>
                  <div className="list-card-main">
                    <h3 className="list-card-title">{task.name}</h3>
                    <p className="list-card-meta">
                      {task.test_cases.length} test cases · {task.difficulty} difficulty
                    </p>
                  </div>
                  <span className={`badge badge-${task.difficulty === "easy" ? "succeeded" : task.difficulty === "medium" ? "running" : "failed"}`}>
                    {task.difficulty}
                  </span>
                </article>
              ))}
            </div>
          </div>
        ) : (
          <p style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
            No tasks defined. Create tasks via the Tasks page.
          </p>
        )}
      </section>
    </>
  );
}

function ExecutionCard({ exec, tasks }: { exec: TaskExecution; tasks: Task[] }) {
  const task = tasks.find((t) => t.id === exec.task_id);
  const taskName = task?.name || exec.task_id;

  return (
    <article className="eval-card">
      <div className="eval-card-header">
        <div>
          <h3 className="eval-card-title">{taskName}</h3>
          <p className="eval-card-meta">Agent: {exec.agent_id}</p>
        </div>
        <span className={`badge badge-${exec.status === "completed" ? "succeeded" : exec.status === "failed" ? "failed" : "running"}`}>
          {exec.status}
        </span>
      </div>
      {exec.duration_ms && (
        <p className="eval-card-body" style={{ marginTop: "var(--space-3)" }}>
          Duration: {exec.duration_ms}ms
        </p>
      )}
      <div className="eval-card-footer">
        <span>Created {new Date(exec.created_at).toLocaleString()}</span>
      </div>
    </article>
  );
}
