"use client";

import { useState } from "react";
import { Experiment, Task } from "../lib/api";
import CreateExperimentForm from "./CreateExperimentForm";
import CompareExperimentsForm from "./CompareExperimentsForm";
import ExperimentCardWithActions from "./ExperimentCardWithActions";
import { Drawer } from "./ui/Drawer";
import { useTranslations } from "next-intl";
import { FlaskConical } from "lucide-react";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";

type Props = {
  experiments: Experiment[];
  tasks: Task[];
};

export default function ExperimentManager({ experiments, tasks }: Props) {
  const t = useTranslations("experiments");
  const tc = useTranslations("common");
  const [showCreateDrawer, setShowCreateDrawer] = useState(false);
  const [showCompare, setShowCompare] = useState(false);

  return (
    <>
      <PageHeader
        eyebrow={t("count", { count: experiments.length })}
        title=""
        actions={
          <>
            <button className="btn btn-secondary" onClick={() => { setShowCompare(!showCompare); setShowCreateDrawer(false); }}>
              {showCompare ? tc("cancel") : t("compare")}
            </button>
            <button className="btn btn-primary" onClick={() => { setShowCreateDrawer(!showCreateDrawer); setShowCompare(false); }}>
              {showCreateDrawer ? tc("cancel") : t("newExperiment")}
            </button>
          </>
        }
      />

      <Drawer isOpen={showCreateDrawer} onClose={() => setShowCreateDrawer(false)} title={t("createExperiment")}>
        <CreateExperimentForm
          tasks={tasks}
          onCreated={() => setShowCreateDrawer(false)}
        />
      </Drawer>

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
            icon={<FlaskConical size={48} aria-hidden="true" />}
            title={t("noExperiments")}
            description={t("noExperimentsDesc")}
            action={
              <button className="btn btn-primary" onClick={() => setShowCreateDrawer(true)}>
                {t("createExperiment")}
              </button>
            }
          />
        )}
      </section>
    </>
  );
}