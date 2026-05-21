"use client";

import { useState, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { Task, Skill, getTasks, getSkills, deleteTask } from "../lib/api";
import { useApiMutation } from "../lib/use-mutations";
import { useDebounce } from "../hooks/use-debounce";
import CreateTaskForm from "./CreateTaskForm";
import TaskFromTraceWizard from "../../components/TaskFromTraceForm";
import TaskCardWithActions from "./TaskCardWithActions";
import EditTaskForm from "./EditTaskForm";
import { useTranslations } from "next-intl";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { QueryError } from "./layout/QueryError";
import { CardGridSkeleton } from "./layout/Skeleton";
import { BulkActionBar } from "./ui/BulkActionBar";
import { ConfirmDialog } from "./ui/ConfirmDialog";
import { RefreshIndicator } from "./ui/RefreshIndicator";
import { SelectableCard } from "./ui/SelectableCard";
import { LoadMore } from "./ui/LoadMore";
import { Drawer } from "./ui/Drawer";
import { Trash2, Search, ListChecks } from "lucide-react";

const PAGE_SIZE = 20;

export default function TaskManagerClient() {
  const t = useTranslations("tasks");
  const [showCreateDrawer, setShowCreateDrawer] = useState(false);
  const [showBuildDrawer, setShowBuildDrawer] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const debouncedSearchQuery = useDebounce(searchQuery, 300);
  const [difficultyFilter, setDifficultyFilter] = useState<string>("all");
  const [skillFilter, setSkillFilter] = useState<string>("all");
  const [selectedTaskIds, setSelectedTaskIds] = useState<Set<string>>(new Set());
  const [displayCount, setDisplayCount] = useState(PAGE_SIZE);
  const [showBulkDeleteConfirm, setShowBulkDeleteConfirm] = useState(false);

  function toggleTaskSelection(taskId: string) {
    setSelectedTaskIds((prev) => {
      const next = new Set(prev);
      if (next.has(taskId)) {
        next.delete(taskId);
      } else {
        next.add(taskId);
      }
      return next;
    });
  }

  function clearSelection() {
    setSelectedTaskIds(new Set());
  }

  const deleteTaskMutation = useApiMutation(deleteTask, {
    successMessage: "Task deleted",
    queryKeysToInvalidate: [["tasks"]],
    onSuccess: () => clearSelection(),
  });

  async function handleBulkDelete() {
    for (const id of selectedTaskIds) {
      await deleteTaskMutation.mutateAsync(id);
    }
    setShowBulkDeleteConfirm(false);
  }

  const { data: tasksData, isLoading, isFetching, isError, refetch } = useQuery({
    queryKey: ["tasks"],
    queryFn: getTasks,
    placeholderData: (previousData) => previousData,
    refetchInterval: 30_000,
  });

  const tasks = tasksData?.items ?? [];

  const { data: skillsData } = useQuery({
    queryKey: ["skills"],
    queryFn: getSkills,
    placeholderData: (previousData) => previousData,
  });

  const skills = skillsData?.items ?? [];

  const filteredTasks = useMemo(() => {
    return tasks.filter((task) => {
      const matchesSearch = debouncedSearchQuery === "" ||
        task.name.toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
        (task.description && task.description.toLowerCase().includes(debouncedSearchQuery.toLowerCase()));

      const matchesDifficulty = difficultyFilter === "all" || task.difficulty === difficultyFilter;

      const matchesSkill = skillFilter === "all" || task.skill_id === skillFilter;

      return matchesSearch && matchesDifficulty && matchesSkill;
    });
  }, [tasks, debouncedSearchQuery, difficultyFilter, skillFilter]);

  const displayedTasks = filteredTasks.slice(0, displayCount);
  const hasMore = displayCount < filteredTasks.length;

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
            <RefreshIndicator isFetching={isFetching} isLoading={isLoading} />
            <button
              className="btn btn-secondary"
              onClick={() => setShowBuildDrawer(true)}
            >
              {t("buildFromTrace")}
            </button>
            <button
              className="btn btn-primary"
              onClick={() => { setShowCreateDrawer(true); setEditingTask(null); }}
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
            onChange={(e) => {
              setSearchQuery(e.target.value);
              setDisplayCount(PAGE_SIZE);
            }}
            className="form-input"
            style={{ width: "100%" }}
          />
        </div>

        <div style={{ flex: "0 0 auto" }}>
          <select
            value={difficultyFilter}
            onChange={(e) => {
              setDifficultyFilter(e.target.value);
              setDisplayCount(PAGE_SIZE);
            }}
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
            onChange={(e) => {
              setSkillFilter(e.target.value);
              setDisplayCount(PAGE_SIZE);
            }}
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
              setDisplayCount(PAGE_SIZE);
            }}
            className="btn btn-ghost"
            style={{ flex: "0 0 auto" }}
          >
            Clear filters
          </button>
        )}
      </div>

      <Drawer isOpen={showCreateDrawer} onClose={() => { setShowCreateDrawer(false); setEditingTask(null); }} title={editingTask ? t("editTask", { name: editingTask.name }) : t("newTask")}>
        <CreateTaskForm
          skills={skills}
          onCreated={() => setShowCreateDrawer(false)}
        />
      </Drawer>

      <Drawer isOpen={showBuildDrawer} onClose={() => setShowBuildDrawer(false)} title={t("buildFromTrace")}>
        <TaskFromTraceWizard
          onTaskCreated={() => setShowBuildDrawer(false)}
          hideHeader
        />
      </Drawer>

      {editingTask && (
        <EditTaskForm
          task={editingTask}
          skills={skills}
          onCancel={() => setEditingTask(null)}
          onUpdated={() => setEditingTask(null)}
        />
      )}

      <section aria-label="Task list">
        {isLoading ? (
          <CardGridSkeleton count={6} columns={3} />
        ) : isError ? (
          <QueryError onRetry={() => refetch()} />
        ) : (
          <div style={{ opacity: isFetching && !isLoading ? 0.5 : 1, transition: "opacity 0.2s" }}>
            {displayedTasks.length > 0 ? (
              <>
                <div className="eval-grid">
                  {displayedTasks.map((task) => (
                    <SelectableCard
                      key={task.id}
                      isSelected={selectedTaskIds.has(task.id)}
                      onToggle={() => toggleTaskSelection(task.id)}
                    >
                      <TaskCardWithActions
                        task={task}
                        onEdit={setEditingTask}
                      />
                    </SelectableCard>
                  ))}
                </div>
                <LoadMore
                  hasMore={hasMore}
                  isLoading={isFetching}
                  onLoadMore={() => setDisplayCount((c) => c + PAGE_SIZE)}
                  label="Load more tasks"
                />
              </>
            ) : tasks.length > 0 ? (
              <EmptyState
                icon={<Search size={48} aria-hidden="true" />}
                title="No matching tasks"
                description="Try adjusting your search or filter criteria"
                action={
                  <button
                    className="btn btn-secondary"
                    onClick={() => {
                      setSearchQuery("");
                      setDifficultyFilter("all");
                      setSkillFilter("all");
                      setDisplayCount(PAGE_SIZE);
                    }}
                  >
                    Clear filters
                  </button>
                }
              />
            ) : (
              <EmptyState
                icon={<ListChecks size={48} aria-hidden="true" />}
                title={t("noTasks")}
                description={t("noTasksDesc")}
                action={
                  <div style={{ display: "flex", gap: "var(--space-3)", justifyContent: "center" }}>
                    <button className="btn btn-primary" onClick={() => setShowCreateDrawer(true)}>
                      {t("createTask")}
                    </button>
                    <button className="btn btn-secondary" onClick={() => setShowBuildDrawer(true)}>
                      {t("buildFromTrace")}
                    </button>
                  </div>
                }
              />
            )}
          </div>
        )}
      </section>

      <BulkActionBar
        selectedCount={selectedTaskIds.size}
        onClear={clearSelection}
        actions={[
          {
            label: `Delete (${selectedTaskIds.size})`,
            icon: <Trash2 size={16} />,
            variant: "danger",
            onClick: () => setShowBulkDeleteConfirm(true),
          },
        ]}
      />

      <ConfirmDialog
        isOpen={showBulkDeleteConfirm}
        onClose={() => setShowBulkDeleteConfirm(false)}
        onConfirm={handleBulkDelete}
        title="Delete tasks"
        description={`Are you sure you want to delete ${selectedTaskIds.size} task${selectedTaskIds.size > 1 ? "s" : ""}? This action cannot be undone.`}
        confirmText="Delete"
        cancelText="Cancel"
        variant="danger"
      />
    </>
  );
}