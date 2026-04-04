"use client";

import { FormEvent, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useAuth } from "../../context/AuthContext";
import { useToast } from "../../components/ToastProvider";
import Breadcrumb from "../../components/Breadcrumb";
import {
  MCPServer,
  listMCPServers,
  createMCPServer,
  updateMCPServer,
  deleteMCPServer,
  approveMCPServer,
} from "../lib/mcp-api";
import {
  Server,
  Terminal,
  Globe,
  CheckCircle,
  Clock,
  XCircle,
  Ban,
  Trash2,
  Plus,
  X,
  Power,
  PowerOff,
} from "lucide-react";

const STATUS_CONFIG: Record<string, { label: string; badgeClass: string; icon: React.ReactNode }> = {
  approved: {
    label: "Approved",
    badgeClass: "badge badge-completed",
    icon: <CheckCircle size={12} aria-hidden="true" />,
  },
  pending: {
    label: "Pending",
    badgeClass: "badge badge-pending",
    icon: <Clock size={12} aria-hidden="true" />,
  },
  rejected: {
    label: "Rejected",
    badgeClass: "badge badge-failed",
    icon: <XCircle size={12} aria-hidden="true" />,
  },
  disabled: {
    label: "Disabled",
    badgeClass: "badge badge-muted",
    icon: <Ban size={12} aria-hidden="true" />,
  },
};

export default function MCPPage() {
  const t = useTranslations("mcp");
  const tc = useTranslations("common");
  const { user } = useAuth();
  const toast = useToast();

  const [servers, setServers] = useState<MCPServer[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [approvingId, setApprovingId] = useState<string | null>(null);

  const [formName, setFormName] = useState("");
  const [formDescription, setFormDescription] = useState("");
  const [formTransportType, setFormTransportType] = useState("stdio");
  const [formCommand, setFormCommand] = useState("");
  const [formURL, setFormURL] = useState("");

  async function loadData() {
    setError(null);
    setIsLoading(true);
    try {
      const resp = await listMCPServers();
      setServers(resp.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load MCP servers");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void loadData();
  }, []);

  function resetForm() {
    setFormName("");
    setFormDescription("");
    setFormTransportType("stdio");
    setFormCommand("");
    setFormURL("");
    setShowCreateForm(false);
  }

  async function handleCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!formName.trim()) return;

    setIsSubmitting(true);
    setError(null);
    try {
      await createMCPServer({
        name: formName.trim(),
        description: formDescription.trim() || undefined,
        transport_type: formTransportType,
        command: formTransportType === "stdio" ? formCommand.trim() || undefined : undefined,
        url: formTransportType === "sse" ? formURL.trim() || undefined : undefined,
      });
      toast.success(tc("success") + ": MCP server created");
      resetForm();
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create MCP server");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteMCPServer(id);
      toast.success(tc("success") + ": MCP server deleted");
      setDeletingId(null);
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to delete MCP server");
    }
  }

  async function handleApprove(id: string, decision: "approved" | "rejected") {
    setApprovingId(id);
    try {
      await approveMCPServer(id, {
        decision,
        approver: user?.username || "admin",
      });
      toast.success(
        decision === "approved"
          ? tc("success") + ": MCP server approved"
          : tc("success") + ": MCP server rejected"
      );
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to update MCP server");
    } finally {
      setApprovingId(null);
    }
  }

  async function handleToggleEnabled(server: MCPServer) {
    try {
      await updateMCPServer(server.id, {
        is_enabled: !server.is_enabled,
      });
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to update MCP server");
    }
  }

  function connectionString(server: MCPServer): string {
    if (server.transport_type === "stdio") {
      return server.command || "—";
    }
    return server.url || "—";
  }

  if (isLoading) {
    return (
      <div className="admin-content">
        <Breadcrumb />
        <div className="loading-pulse loading-pulse-long" style={{ marginBottom: 16 }} />
        <div className="skeleton-grid">
          <div className="skeleton-card" />
          <div className="skeleton-card" />
          <div className="skeleton-card" />
        </div>
      </div>
    );
  }

  return (
    <div className="admin-content">
      <Breadcrumb />

      <div className="page-header">
        <div className="page-header-content">
          <p className="section-kicker">{t("title")}</p>
          <h1 className="panel-title">{t("subtitle")}</h1>
        </div>
        <div className="page-header-actions">
          <button
            className="btn btn-primary"
            onClick={() => setShowCreateForm(true)}
          >
            <Plus size={16} aria-hidden="true" />
            {t("newServer")}
          </button>
        </div>
      </div>

      {error && (
        <div className="form-error" style={{ marginBottom: 16 }}>
          {error}
        </div>
      )}

      {showCreateForm && (
        <div className="form-panel">
          <div className="form-header" style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <div>
              <p className="section-kicker">{t("title")}</p>
              <h2 className="form-title">{t("createServer")}</h2>
            </div>
            <button
              className="btn btn-ghost btn-sm"
              onClick={resetForm}
              aria-label={tc("close")}
            >
              <X size={18} aria-hidden="true" />
            </button>
          </div>
          <form className="form-fields" onSubmit={handleCreate}>
            <div>
              <label className="form-label" htmlFor="mcp-name">{t("serverName")}</label>
              <input
                id="mcp-name"
                className="form-input"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
                placeholder={t("serverNamePlaceholder")}
                required
                autoFocus
              />
            </div>
            <div>
              <label className="form-label" htmlFor="mcp-desc">{tc("description")}</label>
              <input
                id="mcp-desc"
                className="form-input"
                value={formDescription}
                onChange={(e) => setFormDescription(e.target.value)}
                placeholder={t("serverDescPlaceholder")}
              />
            </div>
            <div>
              <label className="form-label" htmlFor="mcp-transport">{t("transport")}</label>
              <select
                id="mcp-transport"
                className="form-select"
                value={formTransportType}
                onChange={(e) => setFormTransportType(e.target.value)}
              >
                <option value="stdio">{t("transportStdio")}</option>
                <option value="sse">{t("transportSSE")}</option>
              </select>
            </div>
            {formTransportType === "stdio" && (
              <div>
                <label className="form-label" htmlFor="mcp-command">{t("command")}</label>
                <input
                  id="mcp-command"
                  className="form-input"
                  value={formCommand}
                  onChange={(e) => setFormCommand(e.target.value)}
                  placeholder={t("commandPlaceholder")}
                />
              </div>
            )}
            {formTransportType === "sse" && (
              <div>
                <label className="form-label" htmlFor="mcp-url">{t("url")}</label>
                <input
                  id="mcp-url"
                  className="form-input"
                  type="url"
                  value={formURL}
                  onChange={(e) => setFormURL(e.target.value)}
                  placeholder={t("urlPlaceholder")}
                />
              </div>
            )}
            <div className="form-actions">
              <button type="button" className="btn btn-secondary" onClick={resetForm}>
                {tc("cancel")}
              </button>
              <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
                {isSubmitting ? t("creating") : t("create")}
              </button>
            </div>
          </form>
        </div>
      )}

      {servers.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state-icon">
            <Server size={48} aria-hidden="true" />
          </div>
          <h3 className="empty-state-title">{t("noServers")}</h3>
          <p className="empty-state-description">
            Register an MCP server to enable AI agent tool integration.
          </p>
          <div className="empty-state-action" style={{ marginTop: 16 }}>
            <button className="btn btn-primary" onClick={() => setShowCreateForm(true)}>
              <Plus size={16} aria-hidden="true" />
              {t("newServer")}
            </button>
          </div>
        </div>
      ) : (
        <div className="panel">
          <div className="gateway-table-wrapper">
            <table className="gateway-table">
              <thead>
                <tr>
                  <th style={{ width: 200 }}>Name</th>
                  <th>Transport</th>
                  <th>Connection</th>
                  <th style={{ width: 100 }}>Status</th>
                  <th style={{ width: 80 }}>Enabled</th>
                  <th style={{ width: 200 }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {servers.map((server) => {
                  const status = STATUS_CONFIG[server.status] || STATUS_CONFIG.disabled;
                  const isPending = server.status === "pending";
                  const isDeleting = deletingId === server.id;
                  const isApproving = approvingId === server.id;

                  return (
                    <tr key={server.id}>
                      <td>
                        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                          <Server size={16} aria-hidden="true" style={{ color: "var(--muted)", flexShrink: 0 }} />
                          <div>
                            <div style={{ fontWeight: 600, color: "var(--ink-strong)", fontSize: "var(--text-sm)" }}>
                              {server.name}
                            </div>
                            {server.description && (
                              <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)", marginTop: 2 }}>
                                {server.description}
                              </div>
                            )}
                          </div>
                        </div>
                      </td>
                      <td>
                        <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
                          {server.transport_type === "stdio" ? (
                            <Terminal size={14} aria-hidden="true" style={{ color: "var(--muted)" }} />
                          ) : (
                            <Globe size={14} aria-hidden="true" style={{ color: "var(--muted)" }} />
                          )}
                          <code style={{ fontSize: "var(--text-xs)", fontFamily: "inherit" }}>
                            {server.transport_type}
                          </code>
                        </div>
                      </td>
                      <td>
                        <code style={{ fontSize: "var(--text-xs)", fontFamily: "'IBM Plex Mono', monospace", color: "var(--ink)" }}>
                          {connectionString(server)}
                        </code>
                      </td>
                      <td>
                        <span className={status.badgeClass} style={{ display: "inline-flex", alignItems: "center", gap: 4 }}>
                          {status.icon}
                          {status.label}
                        </span>
                      </td>
                      <td>
                        <button
                          className={`mcp-toggle-btn ${server.is_enabled ? "mcp-toggle-btn-on" : "mcp-toggle-btn-off"}`}
                          onClick={() => void handleToggleEnabled(server)}
                          title={server.is_enabled ? tc("disable") || "Disable" : tc("enable") || "Enable"}
                          aria-label={server.is_enabled ? "Disable server" : "Enable server"}
                        >
                          {server.is_enabled ? <Power size={14} aria-hidden="true" /> : <PowerOff size={14} aria-hidden="true" />}
                        </button>
                      </td>
                      <td>
                        {isDeleting ? (
                          <div style={{ display: "flex", gap: 4, alignItems: "center" }}>
                            <span style={{ fontSize: "var(--text-xs)", color: "var(--danger)" }}>Confirm?</span>
                            <button
                              className="btn btn-sm"
                              style={{ background: "var(--danger)", color: "#fff", border: "none" }}
                              onClick={() => void handleDelete(server.id)}
                            >
                              {tc("delete")}
                            </button>
                            <button
                              className="btn btn-sm btn-secondary"
                              onClick={() => setDeletingId(null)}
                            >
                              {tc("cancel")}
                            </button>
                          </div>
                        ) : (
                          <div style={{ display: "flex", gap: 4, alignItems: "center" }}>
                            {isPending && (
                              <>
                                <button
                                  className="btn btn-sm"
                                  style={{ background: "var(--success)", color: "#fff", border: "none" }}
                                  disabled={isApproving}
                                  onClick={() => void handleApprove(server.id, "approved")}
                                >
                                  {isApproving ? "..." : t("approve")}
                                </button>
                                <button
                                  className="btn btn-sm"
                                  style={{ background: "var(--danger)", color: "#fff", border: "none" }}
                                  disabled={isApproving}
                                  onClick={() => void handleApprove(server.id, "rejected")}
                                >
                                  {isApproving ? "..." : t("reject")}
                                </button>
                              </>
                            )}
                            <button
                              className="btn btn-sm btn-ghost"
                              onClick={() => setDeletingId(server.id)}
                              aria-label={`Delete ${server.name}`}
                            >
                              <Trash2 size={14} aria-hidden="true" />
                            </button>
                          </div>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
