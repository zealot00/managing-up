import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getReplaySnapshots } from "../lib/api";
import ReplayManager from "../components/ReplayManager";

async function ReplaysContent() {
  const t = await getTranslations("replays");
  const snapshotsResp = await getReplaySnapshots().catch(() => null);
  const snapshots = snapshotsResp?.items ?? [];

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">{t("eyebrow")}</p>
        <h1>{t("title")}</h1>
        <p className="lede">
          {t("lede")}
        </p>
      </header>

      <ReplayManager snapshots={snapshots} />
    </main>
  );
}

function SkeletonReplays() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Replay Layer</p>
        <h1>Replay Snapshots</h1>
        <p className="lede">
          Deterministic execution snapshots for replaying agent behavior.
        </p>
      </header>
      <div className="skeleton-grid">
        {[1, 2, 3].map((i) => (
          <div key={i} className="skeleton-card" />
        ))}
      </div>
    </main>
  );
}

export default function ReplaysPage() {
  return (
    <Suspense fallback={<SkeletonReplays />}>
      <ReplaysContent />
    </Suspense>
  );
}