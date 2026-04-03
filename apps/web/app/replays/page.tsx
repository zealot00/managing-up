import { Suspense } from "react";
import { getReplaySnapshots } from "../lib/api";
import ReplayManager from "../components/ReplayManager";

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

async function ReplaysContent() {
  const snapshotsResp = await getReplaySnapshots().catch(() => null);
  const snapshots = snapshotsResp?.items ?? [];

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Replay Layer</p>
        <h1>Replay Snapshots</h1>
        <p className="lede">
          Deterministic execution snapshots for replaying agent behavior.
          Each snapshot captures state, input seeds, and deterministic RNG seeds.
        </p>
      </header>

      <ReplayManager snapshots={snapshots} />
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
