import { Suspense } from "react";
import { getTasks } from "../lib/api";
import type { Task } from "../lib/api";
import { SkeletonPanel } from "../components/SkeletonPanel";

function SkeletonTasks() {
  return (
    <main className="shell">
      <section className="toprail" aria-label="Tasks navigation">
        <div className="loading-pulse" style={{ width: 180, height: 44, borderRadius: 999 }} />
      </section>
      <section className="hero-page hero-compact">
        <p className="eyebrow">Task Registry</p>
        <h1>Evaluation Tasks</h1>
      </section>
      <div className="skeleton-grid">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div className="skeleton-card" key={i} />
        ))}
      </div>
    </main>
  );
}

function TaskCard({ task }: { task: Task }) {
  return (
    <article className="card">
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 12 }}>
        <div>
          <h2 style={{ margin: "0 0 8px", fontSize: "1.2rem" }}>{task.name}</h2>
          <p style={{ margin: 0, color: "var(--muted)", fontSize: "0.9rem" }}>{task.description || "No description"}</p>
        </div>
        <span className={`badge badge-${task.difficulty === "easy" ? "succeeded" : task.difficulty === "medium" ? "running" : "failed"}`}>
          {task.difficulty}
        </span>
      </div>
      <div style={{ marginTop: 16, display: "flex", gap: 8, flexWrap: "wrap" }}>
        {task.tags.map((tag) => (
          <span key={tag} style={{ padding: "4px 10px", borderRadius: 999, background: "rgba(36, 49, 64, 0.06)", fontSize: "0.78rem", color: "var(--ink)" }}>
            {tag}
          </span>
        ))}
      </div>
      <div style={{ marginTop: 12, paddingTop: 12, borderTop: "1px solid var(--line)", display: "flex", gap: 16, fontSize: "0.82rem", color: "var(--muted)" }}>
        <span>{task.test_cases.length} test cases</span>
        {task.skill_id && <span>Linked to skill</span>}
      </div>
    </article>
  );
}

async function TasksContent() {
  let tasks: { items: Task[] } | null = null;

  try {
    tasks = await getTasks();
  } catch {
    tasks = null;
  }

  return (
    <main className="shell">
      <section className="toprail" aria-label="Tasks navigation">
        <a className="toprail-link" href="/">
          Dashboard
        </a>
        <a className="toprail-link" href="/tasks">
          Tasks
        </a>
        <a className="toprail-link" href="/evaluations">
          Evaluations
        </a>
      </section>

      <section className="hero-page hero-compact">
        <p className="eyebrow">Task Registry</p>
        <h1>Evaluation Tasks</h1>
        <p className="lede">
          Reusable tasks for measuring agent performance. Each task defines inputs,
          expected outputs, and difficulty ratings.
        </p>
      </section>

      <section aria-label="Task list">
        {(tasks?.items ?? []).length > 0 ? (
          <div className="eval-grid">
            {tasks?.items.map((task) => (
              <TaskCard key={task.id} task={task} />
            ))}
          </div>
        ) : (
          <article className="panel" style={{ marginTop: 24 }}>
            <div className="panel-header">
              <h2>No tasks yet</h2>
            </div>
            <p style={{ color: "var(--muted)", marginTop: 12 }}>
              Tasks will appear here once created via the API. Use POST /api/v1/tasks to create evaluation tasks.
            </p>
          </article>
        )}
      </section>
    </main>
  );
}

export default function TasksPage() {
  return (
    <Suspense fallback={<SkeletonTasks />}>
      <TasksContent />
    </Suspense>
  );
}
