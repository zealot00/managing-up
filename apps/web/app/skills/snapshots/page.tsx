import { Suspense } from "react";
import SnapshotHistoryClient from "../../components/SnapshotHistoryClient";
import { ListSkeleton } from "../../components/layout/Skeleton";

function SkeletonSnapshotPage() {
  return (
    <main className="shell">
      <ListSkeleton rows={5} />
    </main>
  );
}

export default function SnapshotHistoryPage() {
  return (
    <Suspense fallback={<SkeletonSnapshotPage />}>
      <SnapshotHistoryClient />
    </Suspense>
  );
}