"use client";

import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { getMySkills } from "../../lib/api";
import { Badge } from "../../components/ui/Badge";
import { Card } from "../../components/ui/Card";
import { PageHeader } from "../../components/layout/PageHeader";
import { EmptyState } from "../../components/layout/EmptyState";
import { CardGridSkeleton } from "../../components/layout/Skeleton";
import { FileEdit, Globe } from "lucide-react";

export function MySkillsClient() {
  const t = useTranslations("mySkills");
  const { data: skills, isLoading, isError } = useQuery({
    queryKey: ["my-skills"],
    queryFn: getMySkills,
  });

  if (isLoading) {
    return (
      <>
        <PageHeader title={t("title")} description={t("description")} />
        <CardGridSkeleton count={6} columns={3} />
      </>
    );
  }

  if (isError) {
    return (
      <>
        <PageHeader title={t("title")} description={t("description")} />
        <div className="panel" role="alert">
          <p className="form-error">{t("noDrafts")}</p>
        </div>
      </>
    );
  }

  const drafts = skills?.filter(s => s.status === "draft") ?? [];
  const published = skills?.filter(s => s.status === "published") ?? [];

  return (
    <>
      <PageHeader title={t("title")} description={t("description")} />

      <section style={{ marginBottom: "var(--space-8)" }}>
        <h2 style={{ fontSize: "var(--text-lg)", fontWeight: 600, color: "var(--ink-strong)", marginBottom: "var(--space-4)" }}>
          {t("drafts")} ({drafts.length})
        </h2>
        {drafts.length > 0 ? (
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))",
              gap: "var(--space-5)",
            }}
          >
            {drafts.map((skill) => (
              <Card key={skill.id}>
                <h3 className="card-title" style={{ fontSize: "var(--text-base)" }}>{skill.name}</h3>
                <p className="card-description" style={{ marginTop: "var(--space-1)" }}>{skill.description}</p>
                <div style={{ marginTop: "var(--space-3)" }}>
                  <Badge variant="pending">{skill.draft_source}</Badge>
                </div>
              </Card>
            ))}
          </div>
        ) : (
          <EmptyState
            icon={<FileEdit size={32} />}
            title={t("noDrafts")}
            description={t("noDraftsDesc")}
          />
        )}
      </section>

      <section>
        <h2 style={{ fontSize: "var(--text-lg)", fontWeight: 600, color: "var(--ink-strong)", marginBottom: "var(--space-4)" }}>
          {t("published")} ({published.length})
        </h2>
        {published.length > 0 ? (
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))",
              gap: "var(--space-5)",
            }}
          >
            {published.map((skill) => (
              <Card key={skill.id}>
                <div className="card-header">
                  <h3 className="card-title" style={{ fontSize: "var(--text-base)" }}>{skill.name}</h3>
                  {skill.verified && <Badge variant="completed">{t("verified")}</Badge>}
                </div>
                <p className="card-description">{skill.description}</p>
                {skill.sop_name && (
                  <div style={{ marginTop: "var(--space-2)", fontSize: "var(--text-xs)", color: "var(--muted)" }}>
                    {t("sop")}: {skill.sop_name}
                  </div>
                )}
              </Card>
            ))}
          </div>
        ) : (
          <EmptyState
            icon={<Globe size={32} />}
            title={t("noPublished")}
            description={t("noPublishedDesc")}
          />
        )}
      </section>
    </>
  );
}
