"use client";

import { useState } from "react";
import { Experiment, Task } from "../lib/api";
import CreateExperimentForm from "./CreateExperimentForm";
import CompareExperimentsForm from "./CompareExperimentsForm";
import ExperimentCardWithActions from "./ExperimentCardWithActions";
import { useTranslations } from "next-intl";

type Props = {
  experiments: Experiment[];
  tasks: Task[];
};

export default function ExperimentManager({ experiments, tasks }: Props) {
  const t = useTranslations("experiments");
  const tc = useTranslations("common");
  const [showCreate, setShowCreate] = useState(false);
  const [showCompare, setShowCompare] = useState(false);

  return (
    <>
      <div className="page-header" style={{ marginBottom: "var(--space-6)", marginTop: "var(--space-4)", paddingBottom: 0, borderBottom: "none" }}>
        <div className="page-header-content">
          <p className="section-kicker" style={{ margin: 0 }}>
            {t("count", { count: experiments.length })}
          </p>
        </div>
        <div className="page-header-actions">
          <button className="btn btn-secondary" onClick={() => { setShowCompare(!showCompare); setShowCreate(false); }}>
            {showCompare ? tc("cancel") : t("compare")}
          </button>
          <button className="btn btn-primary" onClick={() => { setShowCreate(!showCreate); setShowCompare(false); }}>
            {showCreate ? tc("cancel") : t("newExperiment")}
          </button>
        </div>
      </div>

      {showCreate && (
        <CreateExperimentForm
          tasks={tasks}
          onCreated={() => setShowCreate(false)}
        />
      )}

      {showCompare && (
        <CompareExperimentsForm experiments={experiments} />
      )}

      <section aria-label="Experiment list">
        {experiments.length > 0 ? (
          <div className="eval-grid">
            {experiments.map((exp) => (
              <ExperimentCardWithActions key={exp.id} exp={exp} />
            ))}
          </div>
        ) : (
          <div className="empty-state">
            <div className="empty-state-icon">◎</div>
            <h3 className="empty-state-title">{t("noExperiments")}</h3>
            <p className="empty-state-description">
              {t("noExperimentsDesc")}
            </p>
            <div style={{ marginTop: "var(--space-5)" }}>
              <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
                {t("createExperiment")}
              </button>
            </div>
          </div>
        )}
      </section>
    </>
  );
}