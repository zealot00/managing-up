import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getSkills } from "../lib/api";
import SkillsPageClient from "../components/SkillsPageClient";
import { PageSkeleton } from "../components/layout/Skeleton";

function SkeletonSkillsPage() {
  return <PageSkeleton headerActions={true} content="cards" contentCount={5} />;
}

async function SkillsContent() {
  const skills = await getSkills();

  return (
    <main className="shell">
      <SkillsPageClient skills={skills} />
    </main>
  );
}

export default function SkillsPage() {
  return (
    <Suspense fallback={<SkeletonSkillsPage />}>
      <SkillsContent />
    </Suspense>
  );
}
