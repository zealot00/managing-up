import { Suspense } from "react";
import TaskManagerClient from "../components/TaskManagerClient";
import { CardGridSkeleton } from "../components/layout/Skeleton";

function SkeletonTasks() {
  return (
    <main className="shell">
      <CardGridSkeleton count={6} columns={3} />
    </main>
  );
}

export default function TasksPage() {
  return (
    <Suspense fallback={<SkeletonTasks />}>
      <TaskManagerClient />
    </Suspense>
  );
}