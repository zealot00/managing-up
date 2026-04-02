import { Suspense } from "react";
import { getTasks } from "../lib/api";
import type { Task } from "../lib/api";

function SkeletonTasks() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Task Registry</p>
        <h1>Evaluation Tasks</h1>
        <p className="lede">
          Reusable tasks for measuring agent performance. Each task defines inputs, expected outputs, and difficulty ratings.
        </p>
      </header>
      <div className="skeleton-grid">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div key={i} className="skeleton-card" />
        ))}
      </div>
    </main>
  );
}

function TaskCard({ task }: { task: Task }) {
  return (
    <article className="eval-card">
      <div className="eval-card-header">
        <div>
          <h3 className="eval-card-title">{task.name}</h3>
          <p className="eval-card-meta">{task.description || "No description"}</p>
        </div>
        <span className={`badge badge-${task.difficulty === "easy" ? "succeeded" : task.difficulty === "medium" ? "running" : "failed"}`}>
          {task.difficulty}
        </span>
      </div>
      <div className="tags">
        {task.tags.map((tag) => (
          <span key={tag} className="tag">{tag}</span>
        ))}
      </div>
      <div className="eval-card-footer">
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
      <header className="hero-page hero-compact">
        <p className="eyebrow">Task Registry</p>
        <h1>Evaluation Tasks</h1>
        <p className="lede">
          Reusable tasks for measuring agent performance. Each task defines inputs, expected outputs, and difficulty ratings.
        </p>
      </header>

      <section aria-label="Task list">
        {(tasks?.items ?? []).length > 0 ? (
          <div className="eval-grid">
            {tasks?.items.map((task) => (
              <TaskCard key={task.id} task={task} />
            ))}
          </div>
        ) : (
          <div className="empty-state">
            <div className="empty-state-icon">◎</div>
            <h3 className="empty-state-title">No tasks yet</h3>
            <p className="empty-state-description">
              Tasks will appear here once created via the API. Use POST /api/v1/tasks to create evaluation tasks.
            </p>
          </div>
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
