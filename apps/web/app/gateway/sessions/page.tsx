import { Suspense } from "react";
import SessionHistoryClient from "../../components/SessionHistoryClient";
import { ListSkeleton } from "../../components/layout/Skeleton";

function SkeletonSessionPage() {
  return (
    <main className="shell">
      <ListSkeleton rows={5} />
    </main>
  );
}

export default function SessionHistoryPage() {
  return (
    <Suspense fallback={<SkeletonSessionPage />}>
      <SessionHistoryClient />
    </Suspense>
  );
}