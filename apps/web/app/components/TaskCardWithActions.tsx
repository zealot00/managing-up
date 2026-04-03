"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { deleteTask, Task } from "../lib/api";

type Props = {
  task: Task;
  onEdit: (task: Task) => void;
  onDeleted: () => void;
};

export default function TaskCardWithActions({ task, onEdit, onDeleted }: Props) {
  const router = useRouter();
  const [deleting, setDeleting] = useState(false);

  async function handleDelete() {
    if (!confirm(`Delete task "${task.name}"?`)) return;
    setDeleting(true);
    try {
      await deleteTask(task.id);
      onDeleted();
      router.refresh();
    } catch {
      setDeleting(false);
    }
  }

  return (
    <article className="eval-card">
      <div className="eval-card-header">
        <div>
          <h3 className="eval-card-title">{task.name}</h3>
          <p className="eval-card-meta">{task.description || "No description"}</p>
        </div>
        <span className={`badge badge-${task.difficulty === "easy" ? "succeeded" : task.difficulty === "medium" ? "running" : "failed"}`}>
          {task.difficulty}
        </span>
      </div>
      <div className="tags">
        {task.tags.map((tag) => (
          <span key={tag} className="tag">{tag}</span>
        ))}
      </div>
      <div className="eval-card-footer">
        <span>{task.test_cases.length} test cases</span>
        {task.skill_id && <span>Linked to skill</span>}
        <div className="list-card-actions">
          <button
            className="btn btn-sm btn-secondary"
            onClick={() => onEdit(task)}
          >
            Edit
          </button>
          <button
            className="btn btn-sm btn-ghost"
            onClick={handleDelete}
            disabled={deleting}
          >
            {deleting ? "..." : "Delete"}
          </button>
        </div>
      </div>
    </article>
  );
}
