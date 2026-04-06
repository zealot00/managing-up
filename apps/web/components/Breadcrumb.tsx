"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { ChevronRight, Home } from "lucide-react";

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
      seh: t("sehModule"),
      datasets: t("sehDatasets"),
      runs: t("sehRuns"),
      policies: t("sehPolicies"),
      "from-trace": t("taskBuilder"),
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
    <nav className="toprail" aria-label="Breadcrumb">
      {breadcrumbs.map((crumb, index) => {
        const isLast = index === breadcrumbs.length - 1;

        if (isLast) {
          return (
            <span key={crumb.label} className="toprail-chip">
              {crumb.label}
            </span>
          );
        }

        return (
          <Link key={crumb.label} href={crumb.href || "/"} className="toprail-link">
            {index === 0 && <Home size={12} aria-hidden="true" />}
            <span>{crumb.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}