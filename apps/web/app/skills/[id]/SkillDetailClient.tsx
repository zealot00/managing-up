"use client";

import { useState } from "react";
import Link from "next/link";
import { Users, Shield, GitBranch } from "lucide-react";

interface BreadcrumbItem {
  label: string;
  href?: string;
}

interface SkillData {
  id: string;
  name: string;
  owner_team: string;
  risk_level: string;
  status: string;
  current_version?: string;
  created_at?: string;
}

interface VersionData {
  id: string;
  version: string;
  status: string;
  change_summary: string;
  approval_required: boolean;
  created_at: string;
}

interface Translations {
  overview: string;
  versions: string;
  yamlSpec: string;
  ownerTeam: string;
  riskLevel: string;
  status: string;
  version: string;
  noVersions: string;
  approvalRequired: string;
  noApproval: string;
  id: string;
}

interface SkillDetailClientProps {
  skill: SkillData;
  versions: VersionData[];
  specYaml: string;
  breadcrumb: { items: BreadcrumbItem[] };
  translations: Translations;
}

type TabId = "overview" | "versions" | "spec";

const riskColors: Record<string, string> = {
  low: "var(--success)",
  medium: "var(--warning)",
  high: "var(--danger)",
};

export default function SkillDetailClient({
  skill,
  versions,
  specYaml,
  breadcrumb,
  translations: tx,
}: SkillDetailClientProps) {
  const [activeTab, setActiveTab] = useState<TabId>("overview");

  const tabs: { id: TabId; label: string }[] = [
    { id: "overview", label: tx.overview },
    { id: "versions", label: `${tx.versions} (${versions.length})` },
    { id: "spec", label: tx.yamlSpec },
  ];

  return (
    <>
      {/* Breadcrumb */}
      <nav className="breadcrumb" aria-label="Breadcrumb">
        {breadcrumb.items.map((item, index) => {
          const isLast = index === breadcrumb.items.length - 1;
          return (
            <span key={item.label} className="breadcrumb-item">
              {index > 0 && <span className="breadcrumb-sep" aria-hidden="true">/</span>}
              {isLast ? (
                <span className="breadcrumb-current" aria-current="page">{item.label}</span>
              ) : (
                <Link href={item.href || "/"} className="breadcrumb-link">
                  {item.label}
                </Link>
              )}
            </span>
          );
        })}
      </nav>

      {/* Detail Header */}
      <header className="detail-header">
        <div className="detail-header-main">
          <h1 className="detail-header-title">{skill.name}</h1>
          <span className={`badge badge-${skill.status}`}>{skill.status}</span>
        </div>
        <div className="detail-header-chips">
          <span className="detail-chip">
            <Users size={13} className="detail-chip-icon" aria-hidden="true" />
            <span>{skill.owner_team}</span>
          </span>
          <span className="detail-chip">
            <span
              className="detail-chip-dot"
              style={{ background: riskColors[skill.risk_level] || "var(--muted)" }}
              aria-hidden="true"
            />
            <span>{skill.risk_level} risk</span>
          </span>
          {skill.current_version && (
            <span className="detail-chip">
              <GitBranch size={13} className="detail-chip-icon" aria-hidden="true" />
              <span>v{skill.current_version}</span>
            </span>
          )}
        </div>
      </header>

      {/* Tab Bar */}
      <div className="detail-tabs" role="tablist">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            role="tab"
            aria-selected={activeTab === tab.id}
            aria-controls={`skill-tab-${tab.id}`}
            className={`detail-tab ${activeTab === tab.id ? "detail-tab-active" : ""}`}
            onClick={() => setActiveTab(tab.id)}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab Panels */}
      <div className="detail-panel">
        {activeTab === "overview" && (
          <div role="tabpanel" id="skill-tab-overview" className="detail-grid">
            <div className="detail-row">
              <span className="detail-label">{tx.id}</span>
              <span className="detail-value" style={{ fontFamily: "monospace", fontSize: "var(--text-xs)" }}>
                {skill.id}
              </span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{tx.version}</span>
              <span className="detail-value">{skill.current_version || "—"}</span>
            </div>
          </div>
        )}

        {activeTab === "versions" && (
          <div role="tabpanel" id="skill-tab-versions">
            {versions.length === 0 ? (
              <p className="empty-note">{tx.noVersions}</p>
            ) : (
              <div style={{ display: "flex", flexDirection: "column" }}>
                {versions.map((v) => (
                  <div key={v.id} className="version-item">
                    <div className="version-item-main">
                      <span className="version-item-version">{v.version}</span>
                      <span className={`badge badge-${v.status}`}>{v.status}</span>
                    </div>
                    <p className="version-item-summary">{v.change_summary}</p>
                    <p className="version-item-meta">
                      {v.approval_required ? tx.approvalRequired : tx.noApproval} · {new Date(v.created_at).toLocaleDateString()}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {activeTab === "spec" && (
          <div role="tabpanel" id="skill-tab-spec">
            <pre className="json-block">{specYaml}</pre>
          </div>
        )}
      </div>
    </>
  );
}
