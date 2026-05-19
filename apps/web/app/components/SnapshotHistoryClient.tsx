"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { listSnapshots, getSnapshot, SkillCapabilitySnapshot } from "../lib/gateway-api";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { ListSkeleton } from "./layout/Skeleton";
import { Badge } from "./ui/Badge";
import { Search } from "lucide-react";

const PAGE_SIZE = 20;

function SnapshotCard({ snapshot, t }: { snapshot: SkillCapabilitySnapshot; t: ReturnType<typeof useTranslations<"snapshots">> }) {
  return (
    <div className="list-card" style={{ flexDirection: "column", alignItems: "stretch" }}>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "var(--space-2)" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "var(--space-2)" }}>
          <Badge variant={snapshot.passed ? "succeeded" : "failed"}>
            {snapshot.passed ? t("passed") : t("failed")}
          </Badge>
          <Badge variant="draft">{snapshot.snapshot_type}</Badge>
        </div>
        <span className="text-muted">
          {new Date(snapshot.evaluated_at).toLocaleString()}
        </span>
      </div>
      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: "var(--space-2)", fontSize: "var(--text-sm)", marginBottom: "var(--space-2)" }}>
        <div>
          <span className="text-muted">{t("skill")}:</span>{" "}
          <span className="font-mono">{snapshot.skill_id}</span>
        </div>
        <div>
          <span className="text-muted">{t("version")}:</span> <span>{snapshot.version}</span>
        </div>
        <div>
          <span className="text-muted">{t("score")}:</span>{" "}
          <span style={{ fontWeight: 600 }}>{snapshot.overall_score.toFixed(2)}</span>
        </div>
      </div>
      {snapshot.dataset_id && (
        <div className="text-muted" style={{ marginBottom: "var(--space-2)" }}>
          {t("dataset")}: <span className="font-mono">{snapshot.dataset_id}</span>
        </div>
      )}
      <div style={{ borderTop: "1px solid var(--line)", paddingTop: "var(--space-2)" }}>
        <div className="text-muted" style={{ marginBottom: "var(--space-1)" }}>{t("metrics")}:</div>
        <div className="tags">
          {Object.entries(snapshot.metrics).map(([key, value]) => (
            <span key={key} className="tag">
              {key}: {typeof value === "number" ? value.toFixed(3) : value}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}

function SnapshotChecker({ t }: { t: ReturnType<typeof useTranslations<"snapshots">> }) {
  const [skillId, setSkillId] = useState("");
  const [version, setVersion] = useState("");

  const { data: snapshotResult, isLoading, refetch } = useQuery({
    queryKey: ["snapshot-check", skillId, version],
    queryFn: () => getSnapshot({ skill_id: skillId, version }),
    enabled: false,
  });

  const handleCheck = (e: React.FormEvent) => {
    e.preventDefault();
    if (skillId && version) {
      refetch();
    }
  };

  return (
    <div className="panel" style={{ marginBottom: "var(--space-5)" }}>
      <h3 style={{ fontSize: "var(--text-sm)", fontWeight: 600, marginBottom: "var(--space-3)" }}>{t("checkSnapshot")}</h3>
      <form onSubmit={handleCheck} style={{ display: "flex", alignItems: "flex-end", gap: "var(--space-3)" }}>
        <div className="form-group">
          <label className="form-label" htmlFor="check-skill-id">{t("skillId")}</label>
          <input
            id="check-skill-id"
            type="text"
            value={skillId}
            onChange={(e) => setSkillId(e.target.value)}
            placeholder="skill_xxx"
            className="form-input"
          />
        </div>
        <div className="form-group">
          <label className="form-label" htmlFor="check-version">{t("version")}</label>
          <input
            id="check-version"
            type="text"
            value={version}
            onChange={(e) => setVersion(e.target.value)}
            placeholder="1.0.0"
            className="form-input"
          />
        </div>
        <button
          type="submit"
          disabled={!skillId || !version}
          className="btn btn-primary btn-sm"
        >
          {t("check")}
        </button>
      </form>
      {isLoading && <div className="loading-pulse loading-pulse-medium" style={{ width: 80, height: 14, marginTop: "var(--space-2)" }} />}
      {snapshotResult && (
        <div style={{ marginTop: "var(--space-2)" }}>
          {snapshotResult.found ? (
            snapshotResult.snapshot?.passed ? (
              <div style={{ fontSize: "var(--text-sm)", color: "var(--success)" }}>{t("snapshotPassed")}</div>
            ) : (
              <div style={{ fontSize: "var(--text-sm)", color: "var(--danger)" }}>
                {t("snapshotFailed", { score: (snapshotResult.snapshot?.overall_score ?? 0).toFixed(2) })}
              </div>
            )
          ) : (
            <div className="text-muted">{t("noSnapshotFound")}</div>
          )}
        </div>
      )}
    </div>
  );
}

export default function SnapshotHistoryClient() {
  const t = useTranslations("snapshots");
  const [skillFilter, setSkillFilter] = useState("");
  const [displayCount, setDisplayCount] = useState(PAGE_SIZE);

  const { data: snapshotsData, isLoading, isError } = useQuery({
    queryKey: ["snapshots", skillFilter],
    queryFn: () => listSnapshots({ skill_id: skillFilter, limit: 100 }),
    enabled: !!skillFilter,
  });

  const snapshots = Array.isArray(snapshotsData) ? snapshotsData : snapshotsData?.items ?? [];

  const displayedSnapshots = snapshots.slice(0, displayCount);
  const hasMore = displayCount < snapshots.length;

  return (
    <>
      <PageHeader
        title={t("title")}
        description={t("description")}
      />

      <SnapshotChecker t={t} />

      <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
        <div className="form-group" style={{ flex: 1, maxWidth: 400 }}>
          <div style={{ position: "relative" }}>
            <Search size={16} className="search-icon" />
            <input
              type="text"
              placeholder={t("filterBySkillId")}
              value={skillFilter}
              onChange={(e) => setSkillFilter(e.target.value)}
              className="search-input"
            />
          </div>
        </div>
      </div>

      {!skillFilter ? (
        <EmptyState
          icon={<Search size={32} />}
          title={t("enterSkillId")}
          description={t("enterSkillIdDesc")}
        />
      ) : isLoading ? (
        <ListSkeleton rows={5} />
      ) : isError ? (
        <div className="panel" role="alert">
          <p className="form-error">{t("noSnapshots")}</p>
        </div>
      ) : snapshots.length === 0 ? (
        <EmptyState
          icon={<Search size={32} />}
          title={t("noSnapshots")}
          description={t("noSnapshotsDesc")}
        />
      ) : (
        <>
          <div className="list" style={{ marginTop: "var(--space-5)" }}>
            {displayedSnapshots.map((snapshot) => (
              <SnapshotCard key={snapshot.id} snapshot={snapshot} t={t} />
            ))}
          </div>
          {hasMore && (
            <div style={{ display: "flex", justifyContent: "center", marginTop: "var(--space-5)" }}>
              <button
                onClick={() => setDisplayCount((c) => c + PAGE_SIZE)}
                className="btn btn-ghost btn-sm"
              >
                {t("loadMore")}
              </button>
            </div>
          )}
        </>
      )}
    </>
  );
}
