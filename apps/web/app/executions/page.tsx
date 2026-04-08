import { Suspense } from "react";
import ExecutionsPageClient from "../components/ExecutionsPageClient";
import { ListSkeleton } from "../components/layout/Skeleton";

function SkeletonExecutionsPage() {
  return (
    <main className="shell">
      <ListSkeleton rows={5} />
    </main>
  );
}

export default function ExecutionsPage() {
  return (
    <Suspense fallback={<SkeletonExecutionsPage />}>
      <ExecutionsPageClient />
    </Suspense>
  );
}
