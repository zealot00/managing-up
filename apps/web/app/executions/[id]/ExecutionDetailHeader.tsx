"use client";

import { useState } from "react";
import { Sparkles } from "lucide-react";
import { PageHeader } from "../../components/layout/PageHeader";
import { TaskFromTraceDrawer } from "../../components/TaskFromTraceDrawer";
import { useTranslations } from "next-intl";
import type { Execution } from "../../lib/api";

interface ExecutionDetailHeaderProps {
  execution: Execution;
  eyebrow: string;
}

export function ExecutionDetailHeader({ execution, eyebrow }: ExecutionDetailHeaderProps) {
  const t = useTranslations("executions");
  const [showTaskDrawer, setShowTaskDrawer] = useState(false);

  return (
    <>
      <PageHeader
        eyebrow={eyebrow}
        title={execution.skill_name}
        description={
          <>
            {execution.current_step_id} · {t("triggeredBy")} {execution.triggered_by} ·{" "}
            <span className={`badge badge-${execution.status}`}>{execution.status}</span>
          </>
        }
        actions={
          <button
            className="btn btn-secondary"
            onClick={() => setShowTaskDrawer(true)}
          >
            <Sparkles size={16} />
            {t("taskBuilder.buildTask")}
          </button>
        }
      />
      <TaskFromTraceDrawer
        executionId={execution.id}
        isOpen={showTaskDrawer}
        onClose={() => setShowTaskDrawer(false)}
      />
    </>
  );
}