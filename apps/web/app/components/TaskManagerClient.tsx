"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Task, Skill, createTask, updateTask, deleteTask } from "../lib/api";
import CreateTaskForm from "./CreateTaskForm";
import TaskCardWithActions from "./TaskCardWithActions";
import EditTaskForm from "./EditTaskForm";
import { useTranslations } from "next-intl";

type Props = {
  tasks: Task[];
  skills: Skill[];
};

export default function TaskManagerClient({ tasks, skills }: Props) {
  const t = useTranslations("tasks");
  const tc = useTranslations("common");
  const router = useRouter();
  const [showCreate, setShowCreate] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);

  return (
    <>
      <div className="page-header" style={{ marginBottom: "var(--space-6)", marginTop: "var(--space-4)", paddingBottom: 0, borderBottom: "none" }}>
        <div className="page-header-content">
          <p className="section-kicker" style={{ margin: 0 }}>{t("count", { count: tasks.length })}</p>
        </div>
        <div className="page-header-actions">
          <a href="/tasks/from-trace" className="btn btn-secondary">
            {t("buildFromTrace")}
          </a>
          <button
            className="btn btn-primary"
            onClick={() => { setShowCreate(true); setEditingTask(null); }}
          >
            {t("newTask")}
          </button>
        </div>
      </div>

      {showCreate && (
        <CreateTaskForm
          skills={skills}
          onCreated={() => { setShowCreate(false); router.refresh(); }}
        />
      )}

      {editingTask && (
        <EditTaskForm
          task={editingTask}
          skills={skills}
          onCancel={() => setEditingTask(null)}
          onUpdated={() => { setEditingTask(null); router.refresh(); }}
        />
      )}

      <section aria-label="Task list">
        {tasks.length > 0 ? (
          <div className="eval-grid">
            {tasks.map((task) => (
              <TaskCardWithActions
                key={task.id}
                task={task}
                onEdit={setEditingTask}
                onDeleted={() => router.refresh()}
              />
            ))}
          </div>
        ) : (
          <div className="empty-state">
            <div className="empty-state-icon">◎</div>
            <h3 className="empty-state-title">{t("noTasks")}</h3>
            <p className="empty-state-description">
              {t("noTasksDesc")}
            </p>
            <div style={{ marginTop: "var(--space-5)", display: "flex", gap: "var(--space-3)", justifyContent: "center" }}>
              <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
                {t("createTask")}
              </button>
              <a href="/tasks/from-trace" className="btn btn-secondary">
                {t("buildFromTrace")}
              </a>
            </div>
          </div>
        )}
      </section>
    </>
  );
}