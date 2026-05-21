"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";

interface BreadcrumbItem {
  label: string;
  href?: string;
}

export default function Breadcrumb() {
  const t = useTranslations("nav");
  const pathname = usePathname();

  const pathSegments = pathname.split("/").filter(Boolean);

  const breadcrumbs: BreadcrumbItem[] = [
    { label: t("dashboard"), href: "/dashboard" },
  ];

  let currentPath = "";
  for (const segment of pathSegments) {
    currentPath += `/${segment}`;

    if (segment === "dashboard") continue;

    const routeLabels: Record<string, string> = {
      skills: t("skills"),
      executions: t("executions"),
      tasks: t("tasks"),
      evaluations: t("evaluations"),
      experiments: t("experiments"),
      approvals: t("approvals"),
      replays: t("replays"),
      gateway: t("gateway"),
      providers: t("providers"),
      mcp: t("mcp"),
      "mcp-router": t("mcpRouter"),
      metrics: t("mcpRouterMetrics"),
      sessions: t("mcpRouterSessions"),
      seh: t("sehModule"),
      datasets: t("sehDatasets"),
      runs: t("sehRuns"),
      policies: t("sehPolicies"),
      capabilities: t("capabilities"),
      sweeps: t("sweeps"),
      traces: t("traces"),
      market: t("market"),
      "my-skills": t("mySkills"),
      snapshots: t("snapshots"),
      "from-trace": t("taskBuilder"),
      "fallback-chains": t("fallbackChains"),
    };

    const label = routeLabels[segment] || segment;
    breadcrumbs.push({
      label,
      href: currentPath,
    });
  }

  if (breadcrumbs.length <= 1) {
    return null;
  }

  return (
    <nav className="breadcrumb" aria-label="Breadcrumb">
      {breadcrumbs.map((crumb, index) => {
        const isLast = index === breadcrumbs.length - 1;

        return (
          <span key={crumb.label} className="breadcrumb-item">
            {index > 0 && <span className="breadcrumb-sep" aria-hidden="true">/</span>}
            {isLast ? (
              <span className="breadcrumb-current" aria-current="page">{crumb.label}</span>
            ) : (
              <Link href={crumb.href || "/"} className="breadcrumb-link">
                {crumb.label}
              </Link>
            )}
          </span>
        );
      })}
    </nav>
  );
}