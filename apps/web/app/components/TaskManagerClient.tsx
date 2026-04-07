"use client";

import { useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import { Task, Skill } from "../lib/api";
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
  const router = useRouter();
  const [showCreate, setShowCreate] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [difficultyFilter, setDifficultyFilter] = useState<string>("all");
  const [skillFilter, setSkillFilter] = useState<string>("all");

  const filteredTasks = useMemo(() => {
    return tasks.filter((task) => {
      const matchesSearch = searchQuery === "" ||
        task.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (task.description && task.description.toLowerCase().includes(searchQuery.toLowerCase()));

      const matchesDifficulty = difficultyFilter === "all" || task.difficulty === difficultyFilter;

      const matchesSkill = skillFilter === "all" || task.skill_id === skillFilter;

      return matchesSearch && matchesDifficulty && matchesSkill;
    });
  }, [tasks, searchQuery, difficultyFilter, skillFilter]);

  const uniqueDifficulties = useMemo(() => {
    const difficulties = new Set(tasks.map((task) => task.difficulty));
    return Array.from(difficulties).sort();
  }, [tasks]);

  return (
    <>
      <div className="page-header" style={{ marginBottom: "var(--space-6)", marginTop: "var(--space-4)", paddingBottom: 0, borderBottom: "none" }}>
        <div className="page-header-content">
          <p className="section-kicker" style={{ margin: 0 }}>
            {filteredTasks.length === tasks.length
              ? t("count", { count: tasks.length })
              : `${filteredTasks.length} of ${tasks.length} tasks`}
          </p>
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

      <div style={{ display: "flex", gap: "var(--space-4)", marginBottom: "var(--space-6)", flexWrap: "wrap" }}>
        <div style={{ flex: "1 1 240px", maxWidth: 320 }}>
          <input
            type="text"
            placeholder="Search tasks..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="form-input"
            style={{ width: "100%" }}
          />
        </div>

        <div style={{ flex: "0 0 auto" }}>
          <select
            value={difficultyFilter}
            onChange={(e) => setDifficultyFilter(e.target.value)}
            className="form-select"
            style={{ minWidth: 120 }}
          >
            <option value="all">All difficulties</option>
            {uniqueDifficulties.map((diff) => (
              <option key={diff} value={diff}>
                {diff.charAt(0).toUpperCase() + diff.slice(1)}
              </option>
            ))}
          </select>
        </div>

        <div style={{ flex: "0 0 auto" }}>
          <select
            value={skillFilter}
            onChange={(e) => setSkillFilter(e.target.value)}
            className="form-select"
            style={{ minWidth: 140 }}
          >
            <option value="all">All skills</option>
            {skills.map((skill) => (
              <option key={skill.id} value={skill.id}>
                {skill.name}
              </option>
            ))}
          </select>
        </div>

        {(searchQuery || difficultyFilter !== "all" || skillFilter !== "all") && (
          <button
            onClick={() => {
              setSearchQuery("");
              setDifficultyFilter("all");
              setSkillFilter("all");
            }}
            className="btn btn-ghost"
            style={{ flex: "0 0 auto" }}
          >
            Clear filters
          </button>
        )}
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
        {filteredTasks.length > 0 ? (
          <div className="eval-grid">
            {filteredTasks.map((task) => (
              <TaskCardWithActions
                key={task.id}
                task={task}
                onEdit={setEditingTask}
                onDeleted={() => router.refresh()}
              />
            ))}
          </div>
        ) : tasks.length > 0 ? (
          <div className="empty-state">
            <div className="empty-state-icon">🔍</div>
            <h3 className="empty-state-title">No matching tasks</h3>
            <p className="empty-state-description">
              Try adjusting your search or filter criteria
            </p>
            <button
              className="btn btn-secondary"
              onClick={() => {
                setSearchQuery("");
                setDifficultyFilter("all");
                setSkillFilter("all");
              }}
              style={{ marginTop: "var(--space-4)" }}
            >
              Clear filters
            </button>
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
