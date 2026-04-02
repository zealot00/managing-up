import { Suspense } from "react";
import { getReplaySnapshots } from "../lib/api";
import type { ReplaySnapshot } from "../lib/api";

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

function ReplayCard({ snap }: { snap: ReplaySnapshot }) {
  return (
    <article className="eval-card">
      <div className="eval-card-header">
        <div>
          <h3 className="eval-card-title">Snapshot {snap.id.slice(0, 12)}</h3>
          <p className="eval-card-meta">
            Execution: {snap.execution_id.slice(0, 16)}...
          </p>
        </div>
        <span className="badge badge-muted">
          Step {snap.step_index}
        </span>
      </div>
      <div className="detail-grid">
        <div className="detail-row">
          <span className="detail-label">Skill ID</span>
          <span className="detail-value">{snap.skill_id}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Version</span>
          <span className="detail-value">{snap.skill_version}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Seed</span>
          <span className="detail-value" style={{ fontFamily: "monospace", fontSize: "var(--text-xs)" }}>
            {snap.deterministic_seed}
          </span>
        </div>
      </div>
      <div className="eval-card-footer">
        <span>Created: {new Date(snap.created_at).toLocaleString()}</span>
      </div>
    </article>
  );
}

async function ReplaysContent() {
  let snapshots: { items: ReplaySnapshot[] } | null = null;

  try {
    snapshots = await getReplaySnapshots();
  } catch {
    snapshots = null;
  }

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

      <section aria-label="Snapshot list">
        {(snapshots?.items ?? []).length > 0 ? (
          <div className="eval-grid">
            {snapshots?.items.map((snap) => (
              <ReplayCard key={snap.id} snap={snap} />
            ))}
          </div>
        ) : (
          <div className="empty-state">
            <div className="empty-state-icon">◎</div>
            <h3 className="empty-state-title">No snapshots yet</h3>
            <p className="empty-state-description">
              Replay snapshots capture execution state for deterministic replay. They are created when enabled during skill execution.
            </p>
          </div>
        )}
      </section>
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
