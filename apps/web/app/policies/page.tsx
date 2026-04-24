import { Suspense } from "react";
import PoliciesPageClient from "./PoliciesPageClient";
import { ListSkeleton } from "../components/layout/Skeleton";

function SkeletonPoliciesPage() {
  return (
    <main className="shell">
      <ListSkeleton rows={5} />
    </main>
  );
}

export default function PoliciesPage() {
  return (
    <Suspense fallback={<SkeletonPoliciesPage />}>
      <PoliciesPageClient />
    </Suspense>
  );
}