"use client";

import { useState, useMemo } from "react";
import { Task, Skill } from "../lib/api";
import CreateTaskForm from "./CreateTaskForm";
import TaskCardWithActions from "./TaskCardWithActions";
import EditTaskForm from "./EditTaskForm";
import { useTranslations } from "next-intl";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";

type Props = {
  tasks: Task[];
  skills: Skill[];
};

export default function TaskManagerClient({ tasks, skills }: Props) {
  const t = useTranslations("tasks");
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
      <PageHeader
        eyebrow={filteredTasks.length === tasks.length
          ? t("count", { count: tasks.length })
          : `${filteredTasks.length} of ${tasks.length} tasks`}
        title=""
        actions={
          <>
            <a href="/tasks/from-trace" className="btn btn-secondary">
              {t("buildFromTrace")}
            </a>
            <button
              className="btn btn-primary"
              onClick={() => { setShowCreate(true); setEditingTask(null); }}
            >
              {t("newTask")}
            </button>
          </>
        }
      />

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
          onCreated={() => setShowCreate(false)}
        />
      )}

      {editingTask && (
        <EditTaskForm
          task={editingTask}
          skills={skills}
          onCancel={() => setEditingTask(null)}
          onUpdated={() => setEditingTask(null)}
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
              />
            ))}
          </div>
        ) : tasks.length > 0 ? (
          <EmptyState
            icon="🔍"
            title="No matching tasks"
            description="Try adjusting your search or filter criteria"
            action={
              <button
                className="btn btn-secondary"
                onClick={() => {
                  setSearchQuery("");
                  setDifficultyFilter("all");
                  setSkillFilter("all");
                }}
              >
                Clear filters
              </button>
            }
          />
        ) : (
          <EmptyState
            icon="◎"
            title={t("noTasks")}
            description={t("noTasksDesc")}
            action={
              <div style={{ display: "flex", gap: "var(--space-3)", justifyContent: "center" }}>
                <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
                  {t("createTask")}
                </button>
                <a href="/tasks/from-trace" className="btn btn-secondary">
                  {t("buildFromTrace")}
                </a>
              </div>
            }
          />
        )}
      </section>
    </>
  );
}
