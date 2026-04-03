"use client";

import { ReplaySnapshot } from "../lib/api";
import { useTranslations } from "next-intl";

type Props = {
  snapshot: ReplaySnapshot;
  onClose: () => void;
};

export default function ReplayDetail({ snapshot, onClose }: Props) {
  const t = useTranslations("replays");
  const tc = useTranslations("common");

  return (
    <div className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("snapshotDetail")}</h2>
      </div>

      <div className="detail-grid">
        <div className="detail-row">
          <span className="detail-label">{tc("id")}</span>
          <span className="detail-value">{snapshot.id}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">{t("executionId")}</span>
          <span className="detail-value">{snapshot.execution_id}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">{t("skill").split(" ")[0]}</span>
          <span className="detail-value">{snapshot.skill_id}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">{tc("version")}</span>
          <span className="detail-value">{snapshot.skill_version}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">{t("stepIndex")}</span>
          <span className="detail-value">{snapshot.step_index}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">{t("deterministicSeed")}</span>
          <span className="detail-value" style={{ fontFamily: "monospace" }}>
            {snapshot.deterministic_seed}
          </span>
        </div>
      </div>

      <div className="panel-header" style={{ marginTop: "var(--space-6)" }}>
        <p className="section-kicker">{t("state")}</p>
        <h2>{t("stateSnapshot")}</h2>
      </div>
      <pre className="json-block">
        {JSON.stringify(snapshot.state_snapshot, null, 2)}
      </pre>

      <div className="panel-header" style={{ marginTop: "var(--space-6)" }}>
        <p className="section-kicker">{t("input")}</p>
        <h2>{t("inputSeed")}</h2>
      </div>
      <pre className="json-block">
        {JSON.stringify(snapshot.input_seed, null, 2)}
      </pre>

      <div className="form-actions" style={{ marginTop: "var(--space-6)" }}>
        <button onClick={onClose} className="btn btn-secondary" style={{ flex: 1 }}>
          {tc("close")}
        </button>
      </div>
    </div>
  );
}