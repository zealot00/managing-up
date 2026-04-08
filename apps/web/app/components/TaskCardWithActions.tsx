"use client";

import { useState } from "react";
import { deleteTask, Task } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";
import { ConfirmDialog } from "./ui/ConfirmDialog";

type Props = {
  task: Task;
  onEdit: (task: Task) => void;
};

export default function TaskCardWithActions({ task, onEdit }: Props) {
  const t = useTranslations("tasks");
  const tc = useTranslations("common");
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  const deleteMutation = useApiMutation(deleteTask, {
    successMessage: "Task deleted",
    queryKeysToInvalidate: [["tasks"]],
  });

  function handleDelete() {
    deleteMutation.mutate(task.id);
    setShowDeleteDialog(false);
  }

  return (
    <>
      <article className="eval-card">
        <div className="eval-card-header">
          <div>
            <h3 className="eval-card-title">{task.name}</h3>
            <p className="eval-card-meta">{task.description || "No description"}</p>
          </div>
          <span className={`badge badge-${task.difficulty}`}>
            {task.difficulty}
          </span>
        </div>
        <div className="tags">
          {task.tags.map((tag) => (
            <span key={tag} className="tag">{tag}</span>
          ))}
        </div>
        <div className="eval-card-footer">
          <span>{t("testCasesCount", { count: task.test_cases.length })}</span>
          {task.skill_id && <span>{t("linkedToSkill")}</span>}
          <div className="list-card-actions">
            <button
              className="btn btn-sm btn-secondary"
              onClick={() => onEdit(task)}
            >
              {tc("edit")}
            </button>
            <button
              className="btn btn-sm btn-ghost"
              onClick={() => setShowDeleteDialog(true)}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? "..." : tc("delete")}
            </button>
          </div>
        </div>
      </article>

      <ConfirmDialog
        isOpen={showDeleteDialog}
        onClose={() => setShowDeleteDialog(false)}
        onConfirm={handleDelete}
        title={tc("deleteConfirmTitle", { name: task.name })}
        description={tc("deleteConfirmDescription")}
        confirmText={tc("delete")}
        cancelText={tc("cancel")}
        variant="danger"
      />
    </>
  );
}