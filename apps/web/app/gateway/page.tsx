"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Key } from "lucide-react";
import { useAuth } from "../../context/AuthContext";
import {
  GatewayKeyMeta,
  GatewayUsageRow,
  GatewayUserUsageRow,
  getGatewayUsage,
  getGatewayUsageByUsers,
  listGatewayKeys,
  revokeGatewayKey,
} from "../lib/gateway-api";
import BarChart from "../components/BarChart";
import Breadcrumb from "../../components/Breadcrumb";
import { PageHeader } from "../components/layout/PageHeader";
import { EmptyState } from "../components/layout/EmptyState";
import { ConfirmDialog } from "../components/ui/ConfirmDialog";
import { CreateKeyDialog } from "./CreateKeyDialog";

function sumBy<T>(items: T[], selector: (item: T) => number): number {
  return items.reduce((acc, item) => acc + selector(item), 0);
}

function formatCurrency(value: number): string {
  if (value < 0.01) return `$${value.toFixed(4)}`;
  return `$${value.toFixed(2)}`;
}

type DetailView = "requests" | "tokens" | "prompt" | "completion" | "cost" | null;

export default function GatewayPage() {
  const t = useTranslations("gateway");
  const tc = useTranslations("common");
  const { user, isLoading: isAuthLoading } = useAuth();
  const isAdmin = user?.role === "admin";

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [keys, setKeys] = useState<GatewayKeyMeta[]>([]);
  const [usage, setUsage] = useState<GatewayUsageRow[]>([]);
  const [usageByUsers, setUsageByUsers] = useState<GatewayUserUsageRow[]>([]);

  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [adminUserID, setAdminUserID] = useState("");

  const [activeDetail, setActiveDetail] = useState<DetailView>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [revokeTarget, setRevokeTarget] = useState<GatewayKeyMeta | null>(null);
  const [revokeSuccess, setRevokeSuccess] = useState(false);

  const totalRequests = useMemo(() => sumBy(usage, (item) => item.request_count), [usage]);
  const totalTokens = useMemo(() => sumBy(usage, (item) => item.total_tokens), [usage]);
  const promptTokens = useMemo(() => sumBy(usage, (item) => item.prompt_tokens), [usage]);
  const completionTokens = useMemo(() => sumBy(usage, (item) => item.completion_tokens), [usage]);
  const totalCost = useMemo(() => sumBy(usage, (item) => item.total_cost), [usage]);

  const requestsByProviderData = useMemo(() => {
    return usage
      .map((u) => ({ label: `${u.provider}/${u.model}`, value: u.request_count }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 10);
  }, [usage]);

  const tokensByModelData = useMemo(() => {
    return usage
      .map((u) => ({ label: `${u.provider}/${u.model}`, value: u.total_tokens }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 10);
  }, [usage]);

  const promptByModelData = useMemo(() => {
    return usage
      .map((u) => ({ label: `${u.provider}/${u.model}`, value: u.prompt_tokens }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 10);
  }, [usage]);

  const completionByModelData = useMemo(() => {
    return usage
      .map((u) => ({ label: `${u.provider}/${u.model}`, value: u.completion_tokens }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 10);
  }, [usage]);

  const costByProviderData = useMemo(() => {
    return usage
      .map((u) => ({ label: `${u.provider}/${u.model}`, value: u.total_cost }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 10);
  }, [usage]);

  const tokenRankingData = useMemo(() => {
    return usageByUsers
      .map((u) => ({ label: u.username || u.user_id, value: u.total_tokens }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 10);
  }, [usageByUsers]);

  async function loadData(opts?: { preserveError?: boolean }) {
    if (!opts?.preserveError) setError(null);
    setIsLoading(true);
    try {
      const usageParams = {
        from: from || undefined,
        to: to || undefined,
        user_id: isAdmin && adminUserID ? adminUserID : undefined,
      };

      const [keysResp, usageResp, usersResp] = await Promise.all([
        listGatewayKeys(),
        getGatewayUsage(usageParams),
        isAdmin ? getGatewayUsageByUsers({ from: from || undefined, to: to || undefined }) : Promise.resolve({ items: [] }),
      ]);

      setKeys(keysResp.items);
      setUsage(usageResp.items);
      setUsageByUsers(usersResp.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load gateway data");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    if (!isAuthLoading) {
      void loadData();
    }
  }, [isAuthLoading, isAdmin]);

  async function handleRevokeKey() {
    if (!revokeTarget) return;
    setError(null);
    try {
      await revokeGatewayKey(revokeTarget.id);
      setRevokeSuccess(true);
      setTimeout(() => setRevokeSuccess(false), 3000);
      await loadData({ preserveError: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to revoke key");
    }
  }

  async function handleApplyFilter(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    await loadData();
  }

  const statCards = [
    { id: "requests" as DetailView, icon: "#", value: totalRequests.toLocaleString(), label: t("totalRequests"), detail: requestsByProviderData },
    { id: "tokens" as DetailView, icon: "Σ", value: totalTokens.toLocaleString(), label: t("totalTokens"), detail: tokensByModelData },
    { id: "prompt" as DetailView, icon: "↑", value: promptTokens.toLocaleString(), label: t("promptTokens"), detail: promptByModelData },
    { id: "completion" as DetailView, icon: "↓", value: completionTokens.toLocaleString(), label: t("completionTokens"), detail: completionByModelData },
    { id: "cost" as DetailView, icon: "$", value: formatCurrency(totalCost), label: t("totalCost"), detail: costByProviderData },
  ];

  const activeCard = statCards.find((c) => c.id === activeDetail);

  if (isAuthLoading || isLoading) {
    return (
      <>
        <div className="dashboard-stats">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="dashboard-stat-card">
              <div className="loading-pulse loading-pulse-short" style={{ width: 80, marginBottom: 8 }} />
              <div className="loading-pulse" style={{ width: 60, height: 32 }} />
            </div>
          ))}
        </div>
        <div className="gateway-layout">
          <div className="skeleton-card" />
          <div className="skeleton-card" />
        </div>
      </>
    );
  }

  return (
    <>
      <Breadcrumb />
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("lede")}
      />

      {error && (
        <div className="dashboard-section" style={{ borderColor: "var(--ink-strong)", background: "rgba(0,0,0,0.03)" }} role="alert">
          <p className="form-error">{error}</p>
        </div>
      )}

      {revokeSuccess && (
        <div className="gateway-success-notice" role="status">
          {t("revokeSuccess")}
        </div>
      )}

      {/* ── Stat Cards + Detail Panel ── */}
      <div className="dashboard-stats">
        {statCards.map((card) => (
          <article
            key={card.id}
            className={`dashboard-stat-card ${activeDetail === card.id ? "dashboard-stat-card-active" : ""}`}
            onClick={() => setActiveDetail(activeDetail === card.id ? null : card.id)}
            style={{ cursor: "pointer" }}
          >
            <div className="dashboard-stat-icon">{card.icon}</div>
            <div className="dashboard-stat-value">{card.value}</div>
            <div className="dashboard-stat-label">{card.label}</div>
            <div className="dashboard-stat-expand">{t("clickForDetails")}</div>
          </article>
        ))}
      </div>

      {activeCard && (
        <div className="dashboard-section dashboard-detail-panel">
          <div className="dashboard-section-header">
            <h2 className="dashboard-section-title">{activeCard.label} {t("detailBreakdown")}</h2>
            <button
              className="dashboard-detail-close"
              onClick={() => setActiveDetail(null)}
            >
              {t("closeDetails")}
            </button>
          </div>
          <BarChart
            data={activeCard.detail}
            title={t("byProviderModel")}
            valuePrefix={activeCard.id === "cost" ? "$" : ""}
            valueSuffix={activeCard.id !== "cost" ? t("tokensSuffix") : ""}
          />
        </div>
      )}

      {/* ── Two-Column Layout: Keys + Usage ── */}
      <div className="gateway-layout">
        {/* Left: Key Management */}
        <div className="dashboard-section">
          <div className="gateway-keys-header">
            <h2 className="dashboard-section-title">{t("keyManagement")}</h2>
            {keys.length > 0 && (
              <button className="gateway-button-create" onClick={() => setCreateDialogOpen(true)}>
                {t("newKey")}
              </button>
            )}
          </div>
          {keys.length === 0 ? (
            <EmptyState
              icon={<Key size={32} />}
              title={t("emptyTitle")}
              description={t("emptyDesc")}
              action={
                <button className="gateway-button-create" onClick={() => setCreateDialogOpen(true)}>
                  {t("createFirstKey")}
                </button>
              }
            />
          ) : (
            <div className="gateway-table-wrapper">
              <table className="gateway-table">
                <thead>
                  <tr>
                    <th scope="col">{tc("name")}</th>
                    <th scope="col">{t("keyPrefix")}</th>
                    <th scope="col">{tc("status")}</th>
                    <th scope="col"></th>
                  </tr>
                </thead>
                <tbody>
                  {keys.map((key) => (
                    <tr key={key.id} className={key.revoked_at ? "revoked-row" : undefined}>
                      <td>
                        {key.name}
                        <span className="text-muted" style={{ display: "block", fontSize: "var(--text-xs)" }}>
                          {new Date(key.created_at).toLocaleDateString()}
                        </span>
                      </td>
                      <td className="cell-mono">{key.key_prefix}...</td>
                      <td>
                        <span className={`badge ${key.revoked_at ? "badge-failed" : "badge-completed"}`}>
                          {key.revoked_at ? t("revoked") : tc("status")}
                        </span>
                      </td>
                      <td>
                        {!key.revoked_at && (
                          <button className="gateway-button-secondary" onClick={() => setRevokeTarget(key)}>
                            {t("revoke")}
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Right: Usage Analytics */}
        <div className="dashboard-section">
          <div className="dashboard-section-header">
            <h2 className="dashboard-section-title">{t("usageAnalytics")}</h2>
          </div>

          <form className="gateway-filter-form" onSubmit={handleApplyFilter}>
            <div className="gateway-filter-row">
              <div className="gateway-filter-field">
                <label className="form-label" htmlFor="gateway-from">{t("startDate")}</label>
                <input
                  id="gateway-from"
                  type="date"
                  className="form-input"
                  value={from}
                  onChange={(e) => setFrom(e.target.value)}
                />
              </div>
              <div className="gateway-filter-field">
                <label className="form-label" htmlFor="gateway-to">{t("endDate")}</label>
                <input
                  id="gateway-to"
                  type="date"
                  className="form-input"
                  value={to}
                  onChange={(e) => setTo(e.target.value)}
                />
              </div>
              <div className="gateway-filter-submit">
                <button className="gateway-button-secondary" type="submit">{t("apply")}</button>
              </div>
            </div>
          </form>

          {isAdmin && (
            <div className="gateway-filter-field" style={{ marginBottom: "var(--space-4)" }}>
              <label className="form-label" htmlFor="gateway-user-id">{t("user")} ID</label>
              <input
                id="gateway-user-id"
                className="form-input"
                placeholder="user_admin"
                value={adminUserID}
                onChange={(e) => setAdminUserID(e.target.value)}
              />
            </div>
          )}

          {usage.length === 0 ? (
            <p className="empty-note">{t("noUsage")}</p>
          ) : (
            <div className="gateway-table-wrapper">
              <table className="gateway-table">
                <thead>
                  <tr>
                    <th scope="col">{t("provider")}</th>
                    <th scope="col">{t("model")}</th>
                    <th scope="col">{t("requests")}</th>
                    <th scope="col">{t("tokens")}</th>
                    <th scope="col">{t("cost")}</th>
                  </tr>
                </thead>
                <tbody>
                  {usage.map((row) => (
                    <tr key={`${row.client_name}:${row.provider}:${row.model}`}>
                      <td>{row.provider}</td>
                      <td>{row.model}</td>
                      <td>{row.request_count.toLocaleString()}</td>
                      <td>{row.total_tokens.toLocaleString()}</td>
                      <td>{formatCurrency(row.total_cost)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {isAdmin && usageByUsers.length > 0 && (
            <>
              <h3 className="dashboard-section-title" style={{ marginTop: "var(--space-5)" }}>{t("userUsage")}</h3>
              <div className="gateway-table-wrapper">
                <table className="gateway-table">
                  <thead>
                    <tr>
                      <th scope="col">{t("user")}</th>
                      <th scope="col">{t("requests")}</th>
                      <th scope="col">{t("tokens")}</th>
                      <th scope="col">{t("cost")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {usageByUsers.map((row) => (
                      <tr key={row.user_id}>
                        <td>{row.username || row.user_id}</td>
                        <td>{row.request_count.toLocaleString()}</td>
                        <td>{row.total_tokens.toLocaleString()}</td>
                        <td>{formatCurrency(row.total_cost)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          )}

          {isAdmin && tokenRankingData.length > 0 && (
            <div style={{ marginTop: "var(--space-5)" }}>
              <BarChart
                data={costByProviderData}
                title={t("cost") + " " + t("providerUsage")}
                valuePrefix="$"
              />
            </div>
          )}
        </div>
      </div>

      <CreateKeyDialog
        isOpen={createDialogOpen}
        onClose={() => setCreateDialogOpen(false)}
        onCreated={() => void loadData({ preserveError: true })}
        existingKeys={keys}
      />

      <ConfirmDialog
        isOpen={revokeTarget !== null}
        onClose={() => setRevokeTarget(null)}
        onConfirm={handleRevokeKey}
        title={t("revokeConfirmTitle")}
        description={revokeTarget ? t("revokeConfirmDesc", { name: revokeTarget.name }) : undefined}
        confirmText={t("revoke")}
        variant="danger"
      />
    </>
  );
}
