"use client";

import { useState } from "react";
import { Sparkles, User, Footprints } from "lucide-react";
import { TaskFromTraceDrawer } from "../../components/TaskFromTraceDrawer";
import { useTranslations } from "next-intl";
import type { Execution } from "../../lib/api";

interface ExecutionDetailHeaderProps {
  execution: Execution;
  eyebrow: string;
}

export function ExecutionDetailHeader({ execution, eyebrow }: ExecutionDetailHeaderProps) {
  const t = useTranslations("executions");
  const tt = useTranslations("tasks");
  const [showTaskDrawer, setShowTaskDrawer] = useState(false);

  return (
    <>
      <header className="detail-header">
        <div className="detail-header-main">
          <h1 className="detail-header-title">{execution.skill_name}</h1>
          <span className={`badge badge-${execution.status}`}>{execution.status}</span>
        </div>
        <div className="detail-header-chips">
          {execution.current_step_id && (
            <span className="detail-chip">
              <Footprints size={13} className="detail-chip-icon" aria-hidden="true" />
              <span>{execution.current_step_id}</span>
            </span>
          )}
          {execution.triggered_by && (
            <span className="detail-chip">
              <User size={13} className="detail-chip-icon" aria-hidden="true" />
              <span>{t("triggeredBy")} {execution.triggered_by}</span>
            </span>
          )}
        </div>
        <div style={{ marginTop: "var(--space-3)" }}>
          <button
            className="btn btn-secondary"
            onClick={() => setShowTaskDrawer(true)}
          >
            <Sparkles size={16} />
            {tt("taskBuilder.buildTask")}
          </button>
        </div>
      </header>
      <TaskFromTraceDrawer
        executionId={execution.id}
        isOpen={showTaskDrawer}
        onClose={() => setShowTaskDrawer(false)}
      />
    </>
  );
}