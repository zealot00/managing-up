import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getTasks, getSkills } from "../lib/api";
import type { Skill } from "../lib/api";
import TaskManagerClient from "../components/TaskManagerClient";
import { PageSkeleton } from "../components/layout/Skeleton";
import { PageHeader } from "../components/layout/PageHeader";

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
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("lede")}
      />

      <TaskManagerClient tasks={tasks} skills={skills} />
    </main>
  );
}

function SkeletonTasks() {
  return <PageSkeleton headerActions={true} content="cards" contentCount={6} />;
}

export default function TasksPage() {
  return (
    <Suspense fallback={<SkeletonTasks />}>
      <TasksContent />
    </Suspense>
  );
}