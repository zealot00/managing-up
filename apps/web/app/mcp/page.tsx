"use client";

import { FormEvent, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useAuth } from "../../context/AuthContext";
import {
  MCPServer,
  listMCPServers,
  createMCPServer,
  updateMCPServer,
  deleteMCPServer,
  approveMCPServer,
} from "../lib/mcp-api";

export default function MCPPage() {
  const t = useTranslations("mcp");
  const tc = useTranslations("common");
  const { user, isLoading: isAuthLoading } = useAuth();

  const [servers, setServers] = useState<MCPServer[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

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
    if (!isAuthLoading) {
      void loadData();
    }
  }, [isAuthLoading]);

  function resetForm() {
    setFormName("");
    setFormDescription("");
    setFormTransportType("stdio");
    setFormCommand("");
    setFormURL("");
    setShowCreateForm(false);
    setEditingId(null);
  }

  async function handleCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!formName.trim() || !formTransportType) return;

    setIsSubmitting(true);
    setError(null);
    try {
      await createMCPServer({
        name: formName.trim(),
        description: formDescription.trim(),
        transport_type: formTransportType,
        command: formTransportType === "stdio" ? formCommand.trim() : undefined,
        url: formTransportType === "sse" ? formURL.trim() : undefined,
      });
      resetForm();
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create MCP server");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleDelete(id: string) {
    setError(null);
    try {
      await deleteMCPServer(id);
      setDeletingId(null);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete MCP server");
    }
  }

  async function handleApprove(id: string, decision: "approved" | "rejected") {
    setError(null);
    try {
      await approveMCPServer(id, {
        decision,
        approver: user?.username || "admin",
      });
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to approve MCP server");
    }
  }

  async function handleToggleEnabled(server: MCPServer) {
    setError(null);
    try {
      await updateMCPServer(server.id, {
        is_enabled: !server.is_enabled,
      });
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update MCP server");
    }
  }

  if (isAuthLoading || isLoading) {
    return (
      <div className="dashboard-content">
        <div className="dashboard-header">
          <div>
            <h1 className="dashboard-title">{t("title")}</h1>
          </div>
        </div>
        <div className="loading-pulse" style={{ width: 200, height: 24, marginTop: 16 }} />
      </div>
    );
  }

  return (
    <div className="dashboard-content">
      <div className="dashboard-header">
        <div>
          <h1 className="dashboard-title">{t("title")}</h1>
          <p className="dashboard-subtitle">{t("subtitle")}</p>
        </div>
        <button className="form-submit" onClick={() => setShowCreateForm(true)}>
          {t("newServer")}
        </button>
      </div>

      {error && (
        <div className="form-error" style={{ marginBottom: 16 }}>
          {error}
        </div>
      )}

      {showCreateForm && (
        <div className="form-panel">
          <div className="form-header">
            <h2 className="form-title">{t("createServer")}</h2>
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
                className="form-input"
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
                  value={formURL}
                  onChange={(e) => setFormURL(e.target.value)}
                  placeholder={t("urlPlaceholder")}
                />
              </div>
            )}
            <div className="form-actions">
              <button type="button" className="gateway-button-secondary" onClick={resetForm}>
                {tc("cancel")}
              </button>
              <button type="submit" className="form-submit" disabled={isSubmitting}>
                {isSubmitting ? t("creating") : t("create")}
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="dashboard-section">
        {servers.length === 0 ? (
          <p className="empty-note">{t("noServers")}</p>
        ) : (
          <div className="list">
            {servers.map((server) => (
              <article className="list-card" key={server.id}>
                <div className="list-card-main">
                  <div className="list-card-header">
                    <h3 className="list-card-title">{server.name}</h3>
                    <span className={`badge ${
                      server.status === "approved" ? "badge-completed" :
                      server.status === "pending" ? "badge-pending" :
                      server.status === "rejected" ? "badge-failed" :
                      "badge-muted"
                    }`}>
                      {server.status}
                    </span>
                  </div>
                  {server.description && (
                    <p className="list-card-meta">{server.description}</p>
                  )}
                  <p className="list-card-meta">
                    {server.transport_type === "stdio" ? `stdio: ${server.command}` : `sse: ${server.url}`}
                    {server.status === "pending" && server.approved_by && (
                      <span> · approved by {server.approved_by}</span>
                    )}
                  </p>
                </div>
                <div className="list-card-actions">
                  <label className="toggle" title={server.is_enabled ? tc("disable") : tc("enable")}>
                    <input
                      type="checkbox"
                      checked={server.is_enabled}
                      onChange={() => void handleToggleEnabled(server)}
                    />
                    <span className="toggle-slider"></span>
                  </label>
                  {server.status === "pending" && (
                    <>
                      <button
                        className="btn-approve"
                        onClick={() => void handleApprove(server.id, "approved")}
                      >
                        {t("approve")}
                      </button>
                      <button
                        className="btn-reject"
                        onClick={() => void handleApprove(server.id, "rejected")}
                      >
                        {t("reject")}
                      </button>
                    </>
                  )}
                  {deletingId === server.id ? (
                    <div className="list-card-confirm-delete">
                      <span>{tc("confirm")}?</span>
                      <button
                        className="gateway-button-secondary"
                        onClick={() => void handleDelete(server.id)}
                      >
                        {tc("delete")}
                      </button>
                      <button
                        className="gateway-button-secondary"
                        onClick={() => setDeletingId(null)}
                      >
                        {tc("cancel")}
                      </button>
                    </div>
                  ) : (
                    <button
                      className="gateway-button-secondary"
                      onClick={() => setDeletingId(server.id)}
                    >
                      {tc("delete")}
                    </button>
                  )}
                </div>
              </article>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
