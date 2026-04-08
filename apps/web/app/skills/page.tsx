import { Suspense } from "react";
import SkillsPageClient from "../components/SkillsPageClient";
import { TableSkeleton } from "../components/layout/Skeleton";

function SkeletonSkillsPage() {
  return (
    <main className="shell">
      <TableSkeleton rows={5} columns={4} />
    </main>
  );
}

export default function SkillsPage() {
  return (
    <Suspense fallback={<SkeletonSkillsPage />}>
      <SkillsPageClient />
    </Suspense>
  );
}
