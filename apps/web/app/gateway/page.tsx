"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
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

function sumBy<T>(items: T[], selector: (item: T) => number): number {
  return items.reduce((acc, item) => acc + selector(item), 0);
}

export default function GatewayPage() {
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

  const totalRequests = useMemo(() => sumBy(usage, (item) => item.request_count), [usage]);
  const totalTokens = useMemo(() => sumBy(usage, (item) => item.total_tokens), [usage]);
  const promptTokens = useMemo(() => sumBy(usage, (item) => item.prompt_tokens), [usage]);
  const completionTokens = useMemo(() => sumBy(usage, (item) => item.completion_tokens), [usage]);

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
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

  if (isAuthLoading || isLoading) {
    return (
      <main className="shell">
        <header className="hero-page hero-compact">
          <p className="eyebrow">Gateway</p>
          <h1>Model Proxy & Usage</h1>
          <p className="lede">Loading gateway configuration and usage data.</p>
        </header>
        <div className="skeleton-grid">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="skeleton-card" />
          ))}
        </div>
      </main>
    );
  }

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Gateway</p>
        <h1>Model Proxy & Usage</h1>
        <p className="lede">
          Manage platform API keys, proxy model requests, and track token consumption by user/provider/model.
        </p>
      </header>

      {error && (
        <section className="panel">
          <p className="form-error">{error}</p>
        </section>
      )}

      <section className="panel grid panel-grid-wide">
        <article className="form-panel">
          <div className="form-header">
            <h2 className="form-title">Create Gateway Key</h2>
            <p className="form-description">Each key maps to your account and can be revoked independently.</p>
          </div>
          <form className="form-fields" onSubmit={handleCreateKey}>
            <label className="form-label" htmlFor="gateway-key-name">
              Key Name
            </label>
            <input
              id="gateway-key-name"
              className="form-input"
              value={keyName}
              onChange={(e) => setKeyName(e.target.value)}
              placeholder="default"
              disabled={isSubmitting}
            />
            <button className="form-submit" type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Creating..." : "Create Key"}
            </button>
          </form>
          {newKeyValue && (
            <div className="gateway-secret">
              <p className="gateway-secret-title">Copy now: this key is shown only once</p>
              <code className="gateway-secret-code">{newKeyValue}</code>
            </div>
          )}
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Your Keys</p>
            <h2 className="panel-title">Active & Revoked</h2>
          </div>
          {keys.length === 0 ? (
            <p className="empty-note">No gateway keys yet.</p>
          ) : (
            <div className="list">
              {keys.map((key) => (
                <article className="list-card" key={key.id}>
                  <div className="list-card-main">
                    <h3 className="list-card-title">{key.name}</h3>
                    <p className="list-card-meta">
                      {key.key_prefix}... · created {new Date(key.created_at).toLocaleString()}
                      {key.last_used_at ? ` · last used ${new Date(key.last_used_at).toLocaleString()}` : ""}
                    </p>
                  </div>
                  <div className="list-card-actions">
                    <span className={`badge ${key.revoked_at ? "badge-failed" : "badge-completed"}`}>
                      {key.revoked_at ? "revoked" : "active"}
                    </span>
                    {!key.revoked_at && (
                      <button className="gateway-button-secondary" onClick={() => void handleRevokeKey(key.id)}>
                        Revoke
                      </button>
                    )}
                  </div>
                </article>
              ))}
            </div>
          )}
        </article>
      </section>

      <section className="panel">
        <form className="form-fields form-row gateway-filter" onSubmit={handleApplyFilter}>
          <div>
            <label className="form-label" htmlFor="gateway-from">From</label>
            <input
              id="gateway-from"
              type="date"
              className="form-input"
              value={from}
              onChange={(e) => setFrom(e.target.value)}
            />
          </div>
          <div>
            <label className="form-label" htmlFor="gateway-to">To</label>
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
              <label className="form-label" htmlFor="gateway-user-id">User ID (admin)</label>
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
            <button className="form-submit" type="submit">Apply Filter</button>
          </div>
        </form>
      </section>

      <section className="stats">
        <article className="metric-card">
          <div className="metric-card-icon">#</div>
          <div className="metric-card-value">{totalRequests}</div>
          <div className="metric-card-label">Requests</div>
        </article>
        <article className="metric-card">
          <div className="metric-card-icon">Σ</div>
          <div className="metric-card-value">{totalTokens}</div>
          <div className="metric-card-label">Total Tokens</div>
        </article>
        <article className="metric-card">
          <div className="metric-card-icon">↑</div>
          <div className="metric-card-value">{promptTokens}</div>
          <div className="metric-card-label">Prompt Tokens</div>
        </article>
        <article className="metric-card">
          <div className="metric-card-icon">↓</div>
          <div className="metric-card-value">{completionTokens}</div>
          <div className="metric-card-label">Completion Tokens</div>
        </article>
      </section>

      <section className="panel">
        <div className="panel-header">
          <p className="section-kicker">Usage</p>
          <h2 className="panel-title">By Provider / Model</h2>
        </div>
        {usage.length === 0 ? (
          <p className="empty-note">No usage data in the selected time range.</p>
        ) : (
          <div className="gateway-table-wrapper">
            <table className="gateway-table">
              <thead>
                <tr>
                  <th>Provider</th>
                  <th>Model</th>
                  <th>Requests</th>
                  <th>Prompt</th>
                  <th>Completion</th>
                  <th>Total</th>
                </tr>
              </thead>
              <tbody>
                {usage.map((row) => (
                  <tr key={`${row.provider}:${row.model}`}>
                    <td>{row.provider}</td>
                    <td>{row.model}</td>
                    <td>{row.request_count}</td>
                    <td>{row.prompt_tokens}</td>
                    <td>{row.completion_tokens}</td>
                    <td>{row.total_tokens}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {isAdmin && (
        <section className="panel">
          <div className="panel-header">
            <p className="section-kicker">Admin</p>
            <h2 className="panel-title">All Users Usage</h2>
          </div>
          {usageByUsers.length === 0 ? (
            <p className="empty-note">No user-level usage data yet.</p>
          ) : (
            <div className="gateway-table-wrapper">
              <table className="gateway-table">
                <thead>
                  <tr>
                    <th>User</th>
                    <th>User ID</th>
                    <th>Requests</th>
                    <th>Prompt</th>
                    <th>Completion</th>
                    <th>Total</th>
                  </tr>
                </thead>
                <tbody>
                  {usageByUsers.map((row) => (
                    <tr key={row.user_id}>
                      <td>{row.username}</td>
                      <td>{row.user_id}</td>
                      <td>{row.request_count}</td>
                      <td>{row.prompt_tokens}</td>
                      <td>{row.completion_tokens}</td>
                      <td>{row.total_tokens}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      )}
    </main>
  );
}
