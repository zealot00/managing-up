"use client";

import { useState } from "react";
import { ReplaySnapshot } from "../lib/api";
import { useTranslations } from "next-intl";

type Props = {
  onFilter: (filtered: ReplaySnapshot[]) => void;
  allSnapshots: ReplaySnapshot[];
};

export default function ReplayFilter({ onFilter, allSnapshots }: Props) {
  const t = useTranslations("replays");
  const tc = useTranslations("common");
  const [executionId, setExecutionId] = useState("");

  function handleFilter() {
    if (!executionId.trim()) {
      onFilter(allSnapshots);
      return;
    }
    const filtered = allSnapshots.filter((s) =>
      s.execution_id.toLowerCase().includes(executionId.toLowerCase())
    );
    onFilter(filtered);
  }

  return (
    <div className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("filter")}</h2>
      </div>

      <div className="form-fields">
        <label className="form-label">
          {t("executionId")}
          <input
            type="text"
            value={executionId}
            onChange={(e) => setExecutionId(e.target.value)}
            placeholder={t("executionIdPlaceholder")}
            className="form-input"
          />
        </label>
      </div>

      <button onClick={handleFilter} className="form-submit">
        {executionId ? tc("filter") : tc("showAll")}
      </button>
    </div>
  );
}