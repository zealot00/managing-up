import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getExecutions, getSkills } from "../lib/api";
import ExecutionsPageClient from "../components/ExecutionsPageClient";
import { PageSkeleton } from "../components/layout/Skeleton";

function SkeletonExecutionsPage() {
  return <PageSkeleton headerActions={true} filterBar={true} content="cards" contentCount={5} />;
}

async function ExecutionsContent() {
  const [executions, skills] = await Promise.all([getExecutions(), getSkills()]);

  return (
    <main className="shell">
      <ExecutionsPageClient executions={executions} skills={skills.items} />
    </main>
  );
}

export default function ExecutionsPage() {
  return (
    <Suspense fallback={<SkeletonExecutionsPage />}>
      <ExecutionsContent />
    </Suspense>
  );
}
