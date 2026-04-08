"use client";

import { useState, useMemo } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { Skill, createExecution, getExecutions, getSkills } from "../lib/api";
import { useApiMutation } from "../lib/use-mutations";
import { useTranslations } from "next-intl";
import { useToast } from "../../components/ToastProvider";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { FormModal } from "./ui/FormModal";
import { ListSkeleton } from "./layout/Skeleton";

export default function ExecutionsPageClient() {
  const t = useTranslations("executions");
  const tc = useTranslations("common");
  const toast = useToast();
  const [showModal, setShowModal] = useState(false);
  const [skillId, setSkillId] = useState("");
  const [triggeredBy, setTriggeredBy] = useState("");
  const [input, setInput] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");

  const { data: executionsData, isLoading, isFetching } = useQuery({
    queryKey: ["executions"],
    queryFn: getExecutions,
    placeholderData: (previousData) => previousData,
  });

  const executions = executionsData ?? { items: [] };

  const { data: skillsData } = useQuery({
    queryKey: ["skills"],
    queryFn: getSkills,
    placeholderData: (previousData) => previousData,
  });

  const skills = skillsData?.items ?? [];

  const createExecutionMutation = useApiMutation(createExecution, {
    queryKeysToInvalidate: [["executions"]],
    successMessage: tc("success") + ": Execution triggered",
    onSuccess: () => {
      setSkillId("");
      setTriggeredBy("");
      setInput("");
      setShowModal(false);
    },
  });

  const filteredExecutions = useMemo(() => {
    return executions.items.filter((exec) => {
      const matchesStatus = statusFilter === "all" || exec.status === statusFilter;
      const matchesSearch = searchQuery === "" ||
        exec.skill_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        exec.triggered_by.toLowerCase().includes(searchQuery.toLowerCase());
      return matchesStatus && matchesSearch;
    });
  }, [executions.items, statusFilter, searchQuery]);

  const uniqueStatuses = useMemo(() => {
    const statuses = new Set(executions.items.map((e) => e.status));
    return Array.from(statuses).sort();
  }, [executions.items]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    let parsedInput: Record<string, unknown> = {};
    if (input.trim()) {
      try {
        parsedInput = JSON.parse(input);
      } catch {
        toast.error(tc("errors.inputInvalid"));
        return;
      }
    }

    createExecutionMutation.mutate({
      skill_id: skillId,
      triggered_by: triggeredBy,
      input: parsedInput,
    });
  }

  return (
    <>
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("lede")}
        actions={
          <button
            className="btn btn-primary"
            onClick={() => setShowModal(true)}
          >
            + {t("triggerExecution")}
          </button>
        }
      />

      <div style={{ display: "flex", gap: "var(--space-4)", marginBottom: "var(--space-6)", flexWrap: "wrap", alignItems: "center" }}>
        <div style={{ flex: "1 1 200px", maxWidth: 280 }}>
          <input
            type="text"
            placeholder="Search by skill or operator..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="form-input"
            style={{ width: "100%" }}
          />
        </div>

        <div style={{ display: "flex", gap: "var(--space-2)", flexWrap: "wrap" }}>
          <button
            onClick={() => setStatusFilter("all")}
            className={`btn btn-sm ${statusFilter === "all" ? "btn-primary" : "btn-secondary"}`}
          >
            All
          </button>
          {uniqueStatuses.map((status) => (
            <button
              key={status}
              onClick={() => setStatusFilter(status)}
              className={`btn btn-sm ${statusFilter === status ? "btn-primary" : "btn-secondary"}`}
            >
              {status}
            </button>
          ))}
        </div>

        {(searchQuery || statusFilter !== "all") && (
          <button
            onClick={() => { setSearchQuery(""); setStatusFilter("all"); }}
            className="btn btn-ghost btn-sm"
          >
            Clear
          </button>
        )}
      </div>

      {isLoading ? (
        <ListSkeleton rows={5} />
      ) : (
        <div style={{ opacity: isFetching && !isLoading ? 0.5 : 1, transition: "opacity 0.2s" }}>
          <section className="panel">
            <div className="panel-header">
              <p className="section-kicker">{t("runs")}</p>
              <h2 className="panel-title">
                {filteredExecutions.length === executions.items.length
                  ? t("executionQueue")
                  : `${filteredExecutions.length} of ${executions.items.length} runs`}
              </h2>
            </div>
            <div className="list">
              {filteredExecutions.length === 0 ? (
                <EmptyState title={t("noExecutions")} />
              ) : (
                filteredExecutions.map((execution) => (
                  <article className="list-card" key={execution.id}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">{execution.skill_name}</h3>
                      <p className="list-card-meta">
                        {execution.current_step_id} · {t("triggeredBy")} {execution.triggered_by}
                      </p>
                    </div>
                    <div className="list-card-actions">
                      <Link href={`/executions/${execution.id}`} className="trace-link">
                        {t("viewTrace")}
                      </Link>
                      <span className={`badge badge-${execution.status}`}>{execution.status}</span>
                    </div>
                  </article>
                ))
              )}
            </div>
          </section>
        </div>
      )}

      <FormModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        title={t("trigger")}
        eyebrow={t("eyebrow")}
        error={createExecutionMutation.isError ? createExecutionMutation.error?.message : undefined}
        isPending={createExecutionMutation.isPending}
      >
        <form onSubmit={handleSubmit}>
          <div className="form-fields">
            <label className="form-label">
              {t("skill")}
              <select
                value={skillId}
                onChange={(e) => setSkillId(e.target.value)}
                required
                className="form-select"
              >
                <option value="">{t("selectSkill")}</option>
                {skills.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name} ({s.owner_team})
                  </option>
                ))}
              </select>
            </label>

            <label className="form-label">
              {t("triggeredBy")}
              <input
                type="text"
                value={triggeredBy}
                onChange={(e) => setTriggeredBy(e.target.value)}
                placeholder={t("triggeredByPlaceholder")}
                required
                className="form-input"
              />
            </label>

            <label className="form-label">
              {t("input")}
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder={t("inputPlaceholder")}
                rows={3}
                className="form-textarea"
              />
            </label>
          </div>

          <button type="submit" disabled={createExecutionMutation.isPending} className="form-submit" style={{ marginTop: "var(--space-4)" }}>
            {createExecutionMutation.isPending ? t("triggering") : t("trigger")}
          </button>
        </form>
      </FormModal>
    </>
  );
}
