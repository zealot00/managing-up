import { notFound } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { getSkill, getSkillVersions, getSkillSpec } from "../../lib/api";
import SkillDetailClient from "./SkillDetailClient";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function SkillDetailPage({ params }: Props) {
  const { id } = await params;
  const t = await getTranslations("skills");
  const tn = await getTranslations("nav");

  let skill;
  try {
    skill = await getSkill(id);
  } catch {
    notFound();
  }

  const versionsData = await getSkillVersions().catch(() => ({ items: [] as Array<{ id: string; skill_id: string; version: string; status: string; change_summary: string; approval_required: boolean; created_at: string }> }));
  const versions = versionsData.items.filter((v) => v.skill_id === id);

  const specData = await getSkillSpec(id).catch(() => ({ spec_yaml: "" }));

  return (
    <SkillDetailClient
      skill={skill}
      versions={versions}
      specYaml={specData.spec_yaml}
      breadcrumb={{
        items: [
          { label: tn("dashboard"), href: "/dashboard" },
          { label: tn("skills"), href: "/skills" },
          { label: skill.name },
        ],
      }}
      translations={{
        overview: t("skillDetail"),
        versions: t("versions"),
        yamlSpec: t("yamlSpec"),
        ownerTeam: t("ownerTeam"),
        riskLevel: t("riskLevel"),
        status: t("status") || "Status",
        version: t("version") || "Version",
        noVersions: t("noVersions"),
        approvalRequired: t("approvalRequired"),
        noApproval: t("noApproval"),
        id: "ID",
      }}
    />
  );
}
