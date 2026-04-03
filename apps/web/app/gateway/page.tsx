"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { useAuth } from "../../context/AuthContext";
import {
  createGatewayKey,
  GatewayKeyMeta,
  GatewayUsageRow,
  GatewayUserUsageRow,
  getGatewayUsage,
  getGatewayUsageByUsers,
  listGatewayKeys,
  revokeGatewayKey,
} from "../lib/gateway-api";
import BarChart from "../components/BarChart";

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

  const [keyName, setKeyName] = useState("default");
  const [newKeyValue, setNewKeyValue] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [adminUserID, setAdminUserID] = useState("");

  const [activeDetail, setActiveDetail] = useState<DetailView>(null);

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

  async function handleCreateKey(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!keyName.trim()) return;

    setIsSubmitting(true);
    setError(null);
    try {
      const resp = await createGatewayKey(keyName.trim());
      setNewKeyValue(resp.key);
      setKeyName("default");
      await loadData({ preserveError: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create key");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleRevokeKey(id: string) {
    setError(null);
    try {
      await revokeGatewayKey(id);
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
        <div className="skeleton-grid">
          {[1, 2].map((i) => (
            <div key={i} className="skeleton-card" />
          ))}
        </div>
      </>
    );
  }

  return (
    <>
      {error && (
        <div className="dashboard-section" style={{ borderColor: "var(--ink-strong)", background: "rgba(0,0,0,0.03)" }}>
          <p className="form-error">{error}</p>
        </div>
      )}

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
            <div className="dashboard-stat-expand">CLICK FOR DETAILS</div>
          </article>
        ))}
      </div>

      {activeCard && (
        <div className="dashboard-section dashboard-detail-panel">
          <div className="dashboard-section-header">
            <h2 className="dashboard-section-title">{activeCard.label} — Detail Breakdown</h2>
            <button
              className="dashboard-detail-close"
              onClick={() => setActiveDetail(null)}
            >
              ✕ CLOSE
            </button>
          </div>
          <BarChart
            data={activeCard.detail}
            title={`By Provider / Model (Top 10)`}
            valuePrefix={activeCard.id === "cost" ? "$" : ""}
            valueSuffix={activeCard.id !== "cost" ? " tokens" : ""}
          />
        </div>
      )}

      <div className="dashboard-section">
        <form className="form-fields form-row gateway-filter" onSubmit={handleApplyFilter}>
          <div>
            <label className="form-label" htmlFor="gateway-from">{t("startDate")}</label>
            <input
              id="gateway-from"
              type="date"
              className="form-input"
              value={from}
              onChange={(e) => setFrom(e.target.value)}
            />
          </div>
          <div>
            <label className="form-label" htmlFor="gateway-to">{t("endDate")}</label>
            <input
              id="gateway-to"
              type="date"
              className="form-input"
              value={to}
              onChange={(e) => setTo(e.target.value)}
            />
          </div>
          {isAdmin && (
            <div>
              <label className="form-label" htmlFor="gateway-user-id">{t("user")} ID (admin)</label>
              <input
                id="gateway-user-id"
                className="form-input"
                placeholder="user_admin"
                value={adminUserID}
                onChange={(e) => setAdminUserID(e.target.value)}
              />
            </div>
          )}
          <div className="gateway-filter-submit">
            <button className="form-submit" type="submit">{t("apply")}</button>
          </div>
        </form>
      </div>

      {isAdmin && tokenRankingData.length > 0 && (
        <div className="chart-grid">
          <div className="dashboard-section">
            <BarChart
              data={tokenRankingData}
              title={t("userUsage")}
              valueSuffix=" tokens"
            />
          </div>
          <div className="dashboard-section">
            <BarChart
              data={costByProviderData}
              title={t("cost") + " " + t("providerUsage")}
              valuePrefix="$"
            />
          </div>
        </div>
      )}

      <div className="dashboard-section">
        <div className="dashboard-section-header">
          <h2 className="dashboard-section-title">{t("providerUsage")}</h2>
        </div>
        {usage.length === 0 ? (
          <p className="empty-note">{t("noUsage")}</p>
        ) : (
          <div className="gateway-table-wrapper">
            <table className="gateway-table">
              <thead>
                <tr>
                  <th>{t("provider")}</th>
                  <th>{t("model")}</th>
                  <th>{t("requests")}</th>
                  <th>{t("promptTokens").split(" ")[0]}</th>
                  <th>{t("completionTokens").split(" ")[0]}</th>
                  <th>{t("totalTokens").split(" ")[0]}</th>
                  <th>{t("cost")}</th>
                </tr>
              </thead>
              <tbody>
                {usage.map((row) => (
                  <tr key={`${row.provider}:${row.model}`}>
                    <td>{row.provider}</td>
                    <td>{row.model}</td>
                    <td>{row.request_count.toLocaleString()}</td>
                    <td>{row.prompt_tokens.toLocaleString()}</td>
                    <td>{row.completion_tokens.toLocaleString()}</td>
                    <td>{row.total_tokens.toLocaleString()}</td>
                    <td>{formatCurrency(row.total_cost)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="dashboard-section">
        <div className="dashboard-section-header">
          <div>
            <h2 className="dashboard-section-title">{t("keyManagement")}</h2>
          </div>
        </div>
        <div className="gateway-keys-grid">
          <div className="gateway-keys-form">
            <form className="form-fields" onSubmit={handleCreateKey}>
              <label className="form-label" htmlFor="gateway-key-name">
                {t("keyName")}
              </label>
              <input
                id="gateway-key-name"
                className="form-input"
                value={keyName}
                onChange={(e) => setKeyName(e.target.value)}
                placeholder={t("keyNamePlaceholder")}
                disabled={isSubmitting}
              />
              <button className="form-submit" type="submit" disabled={isSubmitting}>
                {isSubmitting ? t("creating") : t("newKey")}
              </button>
            </form>
            {newKeyValue && (
              <div className="gateway-secret">
                <p className="gateway-secret-title">{t("secretWarning")}</p>
                <code className="gateway-secret-code">{newKeyValue}</code>
              </div>
            )}
          </div>
          <div className="gateway-keys-list">
            {keys.length === 0 ? (
              <p className="empty-note">{t("noKeys")}</p>
            ) : (
              <div className="list">
                {keys.map((key) => (
                  <article className="list-card" key={key.id}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">{key.name}</h3>
                      <p className="list-card-meta">
                        {key.key_prefix}... · {tc("createdAt")} {new Date(key.created_at).toLocaleString()}
                        {key.last_used_at ? ` · last used ${new Date(key.last_used_at).toLocaleString()}` : ""}
                      </p>
                    </div>
                    <div className="list-card-actions">
                      <span className={`badge ${key.revoked_at ? "badge-failed" : "badge-completed"}`}>
                        {key.revoked_at ? t("revoke") : tc("status")}
                      </span>
                      {!key.revoked_at && (
                        <button className="gateway-button-secondary" onClick={() => void handleRevokeKey(key.id)}>
                          {t("revoke")}
                        </button>
                      )}
                    </div>
                  </article>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>

      {isAdmin && (
        <div className="dashboard-section">
          <div className="dashboard-section-header">
            <h2 className="dashboard-section-title">{t("userUsage")}</h2>
          </div>
          {usageByUsers.length === 0 ? (
            <p className="empty-note">{t("noUsage")}</p>
          ) : (
            <div className="gateway-table-wrapper">
              <table className="gateway-table">
                <thead>
                  <tr>
                    <th>{t("user")}</th>
                    <th>{t("user")} ID</th>
                    <th>{t("requests")}</th>
                    <th>{t("promptTokens").split(" ")[0]}</th>
                    <th>{t("completionTokens").split(" ")[0]}</th>
                    <th>{t("totalTokens").split(" ")[0]}</th>
                    <th>{t("cost")}</th>
                  </tr>
                </thead>
                <tbody>
                  {usageByUsers.map((row) => (
                    <tr key={row.user_id}>
                      <td>{row.username}</td>
                      <td>{row.user_id}</td>
                      <td>{row.request_count.toLocaleString()}</td>
                      <td>{row.prompt_tokens.toLocaleString()}</td>
                      <td>{row.completion_tokens.toLocaleString()}</td>
                      <td>{row.total_tokens.toLocaleString()}</td>
                      <td>{formatCurrency(row.total_cost)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </>
  );
}