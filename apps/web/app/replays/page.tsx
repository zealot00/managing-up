import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getReplaySnapshots } from "../lib/api";
import ReplayManager from "../components/ReplayManager";
import { PageHeader } from "../components/layout/PageHeader";
import { PageSkeleton } from "../components/layout/Skeleton";

async function ReplaysContent() {
  const t = await getTranslations("replays");
  const snapshotsResp = await getReplaySnapshots().catch(() => null);
  const snapshots = snapshotsResp?.items ?? [];

  return (
    <main className="shell">
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("lede")}
      />

      <ReplayManager snapshots={snapshots} />
    </main>
  );
}

function SkeletonReplays() {
  return <PageSkeleton headerActions={true} content="cards" contentCount={3} />;
}

export default function ReplaysPage() {
  return (
    <Suspense fallback={<SkeletonReplays />}>
      <ReplaysContent />
    </Suspense>
  );
}