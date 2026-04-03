"use client";

import { useState } from "react";
import { ReplaySnapshot } from "../lib/api";
import ReplayFilter from "./ReplayFilter";
import ReplayDetail from "./ReplayDetail";
import { useTranslations } from "next-intl";

type Props = {
  snapshots: ReplaySnapshot[];
};

export default function ReplayManager({ snapshots }: Props) {
  const t = useTranslations("replays");
  const tc = useTranslations("common");
  const [filteredSnapshots, setFilteredSnapshots] = useState<ReplaySnapshot[]>(snapshots);
  const [selectedSnapshot, setSelectedSnapshot] = useState<ReplaySnapshot | null>(null);

  const groupedByExecution = filteredSnapshots.reduce((acc: Record<string, ReplaySnapshot[]>, snap: ReplaySnapshot) => {
    const key = snap.execution_id;
    if (!acc[key]) acc[key] = [];
    acc[key].push(snap);
    return acc;
  }, {} as Record<string, ReplaySnapshot[]>);

  return (
    <>
      <div className="page-header" style={{ marginBottom: "var(--space-6)", marginTop: "var(--space-4)", paddingBottom: 0, borderBottom: "none" }}>
        <div className="page-header-content">
          <p className="section-kicker" style={{ margin: 0 }}>
            {t("count", { count: filteredSnapshots.length, execCount: Object.keys(groupedByExecution).length })}
          </p>
        </div>
      </div>

      <section aria-label="Filter" style={{ marginTop: "var(--space-4)" }}>
        <ReplayFilter
          onFilter={(filtered: ReplaySnapshot[]) => setFilteredSnapshots(filtered)}
          allSnapshots={snapshots}
        />
      </section>

      {selectedSnapshot && (
        <ReplayDetail
          snapshot={selectedSnapshot}
          onClose={() => setSelectedSnapshot(null)}
        />
      )}

      <section aria-label="Snapshot list" style={{ marginTop: "var(--space-6)" }}>
        {Object.keys(groupedByExecution).length > 0 ? (
          Object.entries(groupedByExecution).map(([execId, snaps]) => (
            <div key={execId} style={{ marginBottom: "var(--space-8)" }}>
              <div className="panel">
                <div className="panel-header">
                  <p className="section-kicker">{t("executions").split(" ").slice(0, 1).join("")}</p>
                  <h2 className="panel-title">{execId.slice(0, 32)}... ({snaps.length} {t("title").split(" ").slice(1).join("")})</h2>
                </div>
                <div className="eval-grid">
                  {snaps.map((snap) => (
                    <article
                      key={snap.id}
                      className="eval-card"
                      style={{ cursor: "pointer" }}
                      onClick={() => setSelectedSnapshot(snap)}
                    >
                      <div className="eval-card-header">
                        <div>
                          <h3 className="eval-card-title">{t("snapshotDetail").split(" ")[0]} {snap.id.slice(0, 12)}</h3>
                          <p className="eval-card-meta">
                            {t("stepIndex").split(" ")[0]} {snap.step_index}
                          </p>
                        </div>
                        <span className="badge badge-muted">
                          {t("stepIndex").split(" ")[0]} {snap.step_index}
                        </span>
                      </div>
                      <div className="detail-grid">
                        <div className="detail-row">
                          <span className="detail-label">{t("skill").split(" ")[0]}</span>
                          <span className="detail-value">{snap.skill_id}</span>
                        </div>
                        <div className="detail-row">
                          <span className="detail-label">{tc("version")}</span>
                          <span className="detail-value">{snap.skill_version}</span>
                        </div>
                        <div className="detail-row">
                          <span className="detail-label">{t("deterministicSeed").split(" ")[0]}</span>
                          <span className="detail-value" style={{ fontFamily: "monospace", fontSize: "var(--text-xs)" }}>
                            {snap.deterministic_seed}
                          </span>
                        </div>
                      </div>
                      <div className="eval-card-footer">
                        <span>{tc("createdAt")}: {new Date(snap.created_at).toLocaleString()}</span>
                      </div>
                    </article>
                  ))}
                </div>
              </div>
            </div>
          ))
        ) : (
          <div className="empty-state">
            <div className="empty-state-icon">◎</div>
            <h3 className="empty-state-title">{t("noSnapshots")}</h3>
            <p className="empty-state-description">
              {t("noSnapshotsDesc")}
            </p>
          </div>
        )}
      </section>
    </>
  );
}