"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Drawer } from "./ui/Drawer";
import TaskFromTraceForm from "../../components/TaskFromTraceForm";
import { useTranslations } from "next-intl";
import type { Task } from "../lib/api";

interface TaskFromTraceDrawerProps {
  executionId: string;
  isOpen: boolean;
  onClose: () => void;
}

export function TaskFromTraceDrawer({ executionId, isOpen, onClose }: TaskFromTraceDrawerProps) {
  const t = useTranslations("tasks");
  const router = useRouter();

  function handleTaskCreated(task: Task) {
    onClose();
    setTimeout(() => {
      router.push(`/tasks?highlight=${task.id}`);
    }, 500);
  }

  return (
    <Drawer
      isOpen={isOpen}
      onClose={onClose}
      title={t("taskBuilder.title")}
      description={t("taskBuilder.drawerDescription")}
    >
      <div style={{ maxWidth: 480 }}>
        <TaskFromTraceForm
          initialExecutionId={executionId}
          onTaskCreated={handleTaskCreated}
        />
      </div>
    </Drawer>
  );
}