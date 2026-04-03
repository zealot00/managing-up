import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getTasks, getSkills } from "../lib/api";
import type { Skill } from "../lib/api";
import TaskManagerClient from "../components/TaskManagerClient";

async function TasksContent() {
  const t = await getTranslations("tasks");
  const [tasksResp, skillsResp] = await Promise.all([
    getTasks().catch(() => null),
    getSkills().catch(() => ({ items: [] as Skill[] })),
  ]);

  const tasks = tasksResp?.items ?? [];
  const skills = skillsResp?.items ?? [];

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">{t("eyebrow")}</p>
        <h1>{t("title")}</h1>
        <p className="lede">
          {t("lede")}
        </p>
      </header>

      <TaskManagerClient tasks={tasks} skills={skills} />
    </main>
  );
}

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

export default function TasksPage() {
  return (
    <Suspense fallback={<SkeletonTasks />}>
      <TasksContent />
    </Suspense>
  );
}