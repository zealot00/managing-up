"use client";

import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { getSkillMarket } from "../../lib/api";
import { Badge } from "../../components/ui/Badge";
import { Card } from "../../components/ui/Card";
import { PageHeader } from "../../components/layout/PageHeader";
import { EmptyState } from "../../components/layout/EmptyState";
import { CardGridSkeleton } from "../../components/layout/Skeleton";
import { Package } from "lucide-react";

export function SkillMarketClient() {
  const t = useTranslations("skillMarket");
  const { data: skills, isLoading, isError } = useQuery({
    queryKey: ["skill-market"],
    queryFn: () => getSkillMarket({}),
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
          <p className="form-error">{t("noSkills")}</p>
        </div>
      </>
    );
  }

  return (
    <>
      <PageHeader title={t("title")} description={t("description")} />

      {skills && skills.length > 0 ? (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))",
            gap: "var(--space-5)",
          }}
        >
          {skills.map((skill) => (
            <Card key={skill.id}>
              <div className="card-header">
                <h3 className="card-title">{skill.name}</h3>
                {skill.verified && <Badge variant="completed">{t("verified")}</Badge>}
              </div>
              <p className="card-description">{skill.description}</p>

              {skill.tags && skill.tags.length > 0 && (
                <div className="tags">
                  {skill.tags.map((tag) => (
                    <span key={tag} className="tag">{tag}</span>
                  ))}
                </div>
              )}

              <div className="card-footer">
                <span style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
                  {t("trust")}: {skill.trust_score.toFixed(2)}
                </span>
                {skill.avg_rating > 0 && (
                  <span style={{ fontSize: "var(--text-sm)" }}>
                    ⭐ {skill.avg_rating.toFixed(1)} ({skill.rating_count})
                  </span>
                )}
              </div>

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
          icon={<Package size={32} />}
          title={t("noSkills")}
          description={t("noSkillsDesc")}
        />
      )}
    </>
  );
}
