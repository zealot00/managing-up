"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { listSnapshots, getSnapshot, SkillCapabilitySnapshot } from "../lib/gateway-api";
import { PageHeader } from "./layout/PageHeader";
import { ListSkeleton } from "./layout/Skeleton";
import { Badge } from "./ui/Badge";

const PAGE_SIZE = 20;

function SnapshotCard({ snapshot }: { snapshot: SkillCapabilitySnapshot }) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <Badge variant={snapshot.passed ? "succeeded" : "failed"}>
            {snapshot.passed ? "PASSED" : "FAILED"}
          </Badge>
          <Badge variant="outline" className="bg-blue-50 text-blue-700">
            {snapshot.snapshot_type}
          </Badge>
        </div>
        <span className="text-sm text-gray-500">
          {new Date(snapshot.evaluated_at).toLocaleString()}
        </span>
      </div>
      <div className="grid grid-cols-3 gap-2 text-sm mb-2">
        <div>
          <span className="text-gray-500">Skill:</span>{" "}
          <span className="font-mono">{snapshot.skill_id}</span>
        </div>
        <div>
          <span className="text-gray-500">Version:</span> <span>{snapshot.version}</span>
        </div>
        <div>
          <span className="text-gray-500">Score:</span>{" "}
          <span className="font-semibold">{snapshot.overall_score.toFixed(2)}</span>
        </div>
      </div>
      {snapshot.dataset_id && (
        <div className="text-sm text-gray-500 mb-2">
          Dataset: <span className="font-mono">{snapshot.dataset_id}</span>
        </div>
      )}
      <div className="border-t border-gray-100 pt-2">
        <div className="text-sm text-gray-500 mb-1">Metrics:</div>
        <div className="flex flex-wrap gap-2">
          {Object.entries(snapshot.metrics).map(([key, value]) => (
            <span key={key} className="text-xs bg-gray-100 px-2 py-1 rounded">
              {key}: {typeof value === "number" ? value.toFixed(3) : value}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}

function SnapshotChecker() {
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
    <div className="bg-white rounded-lg border border-gray-200 p-4 mb-4">
      <h3 className="text-sm font-medium mb-2">Check Skill Snapshot</h3>
      <form onSubmit={handleCheck} className="flex items-end gap-2">
        <div>
          <label className="block text-xs text-gray-500 mb-1">Skill ID</label>
          <input
            type="text"
            value={skillId}
            onChange={(e) => setSkillId(e.target.value)}
            placeholder="skill_xxx"
            className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div>
          <label className="block text-xs text-gray-500 mb-1">Version</label>
          <input
            type="text"
            value={version}
            onChange={(e) => setVersion(e.target.value)}
            placeholder="1.0.0"
            className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <button
          type="submit"
          disabled={!skillId || !version}
          className="px-4 py-2 bg-blue-600 text-white rounded-md text-sm hover:bg-blue-700 disabled:opacity-50"
        >
          Check
        </button>
      </form>
      {isLoading && <div className="mt-2 text-sm text-gray-500">Loading...</div>}
      {snapshotResult && (
        <div className="mt-2">
          {snapshotResult.found ? (
            snapshotResult.snapshot?.passed ? (
              <div className="text-sm text-green-600">Snapshot exists and PASSED</div>
            ) : (
              <div className="text-sm text-red-600">
                Snapshot exists but FAILED (score:{" "}
                {snapshotResult.snapshot?.overall_score.toFixed(2)})
              </div>
            )
          ) : (
            <div className="text-sm text-gray-500">No passed snapshot found for this version</div>
          )}
        </div>
      )}
    </div>
  );
}

export default function SnapshotHistoryClient() {
  const [skillFilter, setSkillFilter] = useState("");
  const [displayCount, setDisplayCount] = useState(PAGE_SIZE);

  const { data: snapshotsData, isLoading } = useQuery({
    queryKey: ["snapshots", skillFilter],
    queryFn: () => listSnapshots({ skill_id: skillFilter, limit: 100 }),
    enabled: !!skillFilter,
  });

  const snapshots = snapshotsData?.items ?? [];

  const displayedSnapshots = snapshots.slice(0, displayCount);
  const hasMore = displayCount < snapshots.length;

  return (
    <div className="space-y-4">
      <PageHeader
        title="Skill Capability Snapshots"
        description="View skill evaluation snapshots and regression gate results"
      />

      <SnapshotChecker />

      <div className="flex items-center gap-4">
        <input
          type="text"
          placeholder="Filter by skill ID..."
          value={skillFilter}
          onChange={(e) => setSkillFilter(e.target.value)}
          className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      {!skillFilter ? (
        <div className="text-center py-12 text-gray-500">
          Enter a skill ID above to view snapshots
        </div>
      ) : isLoading ? (
        <ListSkeleton rows={5} />
      ) : snapshots.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          No snapshots found for this skill
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {displayedSnapshots.map((snapshot) => (
              <SnapshotCard key={snapshot.id} snapshot={snapshot} />
            ))}
          </div>
          {hasMore && (
            <div className="flex justify-center">
              <button
                onClick={() => setDisplayCount((c) => c + PAGE_SIZE)}
                className="px-4 py-2 text-sm text-blue-600 hover:text-blue-800"
              >
                Load more
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}