"use client";

import { useState } from "react";
import { Experiment, Task } from "../lib/api";
import CreateExperimentForm from "./CreateExperimentForm";
import CompareExperimentsForm from "./CompareExperimentsForm";
import ExperimentCardWithActions from "./ExperimentCardWithActions";
import { useTranslations } from "next-intl";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";

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
      <PageHeader
        eyebrow={t("count", { count: experiments.length })}
        title=""
        actions={
          <>
            <button className="btn btn-secondary" onClick={() => { setShowCompare(!showCompare); setShowCreate(false); }}>
              {showCompare ? tc("cancel") : t("compare")}
            </button>
            <button className="btn btn-primary" onClick={() => { setShowCreate(!showCreate); setShowCompare(false); }}>
              {showCreate ? tc("cancel") : t("newExperiment")}
            </button>
          </>
        }
      />

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
          <EmptyState
            icon="◎"
            title={t("noExperiments")}
            description={t("noExperimentsDesc")}
            action={
              <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
                {t("createExperiment")}
              </button>
            }
          />
        )}
      </section>
    </>
  );
}