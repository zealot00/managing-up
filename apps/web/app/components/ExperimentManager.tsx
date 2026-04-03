"use client";

import { useState } from "react";
import { Experiment, Task } from "../lib/api";
import CreateExperimentForm from "./CreateExperimentForm";
import CompareExperimentsForm from "./CompareExperimentsForm";
import ExperimentCardWithActions from "./ExperimentCardWithActions";

type Props = {
  experiments: Experiment[];
  tasks: Task[];
};

export default function ExperimentManager({ experiments, tasks }: Props) {
  const [showCreate, setShowCreate] = useState(false);
  const [showCompare, setShowCompare] = useState(false);

  return (
    <>
      <div className="page-header" style={{ marginBottom: "var(--space-6)", marginTop: "var(--space-4)", paddingBottom: 0, borderBottom: "none" }}>
        <div className="page-header-content">
          <p className="section-kicker" style={{ margin: 0 }}>
            {experiments.length} experiments defined
          </p>
        </div>
        <div className="page-header-actions">
          <button className="btn btn-secondary" onClick={() => { setShowCompare(!showCompare); setShowCreate(false); }}>
            {showCompare ? "Cancel" : "Compare"}
          </button>
          <button className="btn btn-primary" onClick={() => { setShowCreate(!showCreate); setShowCompare(false); }}>
            {showCreate ? "Cancel" : "+ New Experiment"}
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
            <h3 className="empty-state-title">No experiments yet</h3>
            <p className="empty-state-description">
              Create your first experiment to start comparing agent performance across tasks.
            </p>
            <div style={{ marginTop: "var(--space-5)" }}>
              <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
                Create Experiment
              </button>
            </div>
          </div>
        )}
      </section>
    </>
  );
}
