"use client";

import { FormEvent, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useAuth } from "../../../context/AuthContext";
import {
  createProviderKey,
  deleteProviderKey,
  getBudget,
  listProviderKeys,
  toggleProviderKey,
  updateBudget,
  GatewayProviderKey,
  UserBudget,
} from "../../lib/gateway-api";
import { useToast } from "../../../components/ToastProvider";
import Breadcrumb from "../../../components/Breadcrumb";
import { EmptyState } from "../../components/layout/EmptyState";
import { ConfirmDialog } from "../../components/ui/ConfirmDialog";
import {
  Server,
  Plus,
  X,
  Power,
  PowerOff,
  Trash2,
  Key,
  Calculator,
  ChevronDown,
  ChevronUp,
} from "lucide-react";

const PROVIDERS = [
  { value: "openai", label: "OpenAI" },
  { value: "anthropic", label: "Anthropic" },
  { value: "google", label: "Google" },
  { value: "azure", label: "Azure" },
  { value: "ollama", label: "Ollama" },
  { value: "minimax", label: "Minimax" },
  { value: "zhipuai", label: "Zhipu AI" },
  { value: "deepseek", label: "DeepSeek" },
  { value: "baidu", label: "Baidu" },
  { value: "alibaba", label: "Alibaba" },
];

function BudgetInline({
  budget,
  expanded,
  onToggle,
  onEdit,
}: {
  budget: UserBudget | null;
  expanded: boolean;
  onToggle: () => void;
  onEdit: () => void;
}) {
  const t = useTranslations("providers");
  const tc = useTranslations("common");

  return (
    <div className="panel">
      <button
        className="panel-header-button"
        onClick={onToggle}
        style={{
          width: "100%",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          padding: "12px 16px",
          cursor: "pointer",
          background: "transparent",
          border: "none",
          borderBottom: expanded ? "1px solid var(--line)" : "1px solid transparent",
          marginBottom: expanded ? 0 : -1,
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <Calculator size={16} aria-hidden="true" style={{ color: "var(--muted)" }} />
          <span style={{ fontWeight: 600, fontSize: "var(--text-sm)", color: "var(--ink)" }}>
            {t("budget")}
          </span>
        </div>
        {expanded ? (
          <ChevronUp size={16} aria-hidden="true" />
        ) : (
          <ChevronDown size={16} aria-hidden="true" />
        )}
      </button>

      {expanded && (
        <div style={{ padding: "12px 16px" }}>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(3, 1fr)",
              gap: 16,
              marginBottom: 16,
            }}
          >
            <div>
              <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)", marginBottom: 4, display: "flex", alignItems: "center", gap: 4 }}>
                {t("usedThisMonth")}
                <span title={t("usedThisMonthTooltip")} style={{ cursor: "help" }}>?</span>
              </div>
              <div style={{ fontWeight: 700, fontSize: "var(--text-base)", color: "var(--ink)" }}>
                {budget?.used_this_month.toLocaleString() || 0}
              </div>
            </div>
            <div>
              <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)", marginBottom: 4 }}>
                {t("monthlyLimit")}
              </div>
              <div style={{ fontWeight: 700, fontSize: "var(--text-base)", color: "var(--ink)" }}>
                {budget?.monthly_limit.toLocaleString() || 0}
              </div>
            </div>
            <div>
              <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)", marginBottom: 4 }}>
                {t("dailyLimit")}
              </div>
              <div style={{ fontWeight: 700, fontSize: "var(--text-base)", color: "var(--ink)" }}>
                {budget?.daily_limit.toLocaleString() || 0}
              </div>
            </div>
          </div>
          <button className="btn btn-sm btn-secondary" onClick={onEdit}>
            {tc("edit")}
          </button>
        </div>
      )}
    </div>
  );
}

function BudgetEditForm({
  budget,
  onSave,
  onCancel,
  isSubmitting,
}: {
  budget: UserBudget | null;
  onSave: (monthly: number, daily: number) => void;
  onCancel: () => void;
  isSubmitting: boolean;
}) {
  const t = useTranslations("providers");
  const tc = useTranslations("common");
  const [monthly, setMonthly] = useState(budget?.monthly_limit ?? 0);
  const [daily, setDaily] = useState(budget?.daily_limit ?? 0);

  return (
    <div className="panel" style={{ padding: 16 }}>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr auto",
          gap: 12,
          alignItems: "end",
        }}
      >
        <div>
          <label className="form-label" htmlFor="budget-monthly">
            {t("monthlyLimit")}
          </label>
          <input
            id="budget-monthly"
            type="number"
            className="form-input"
            value={monthly}
            onChange={(e) => setMonthly(parseInt(e.target.value) || 0)}
            min={0}
          />
        </div>
        <div>
          <label className="form-label" htmlFor="budget-daily">
            {t("dailyLimit")}
          </label>
          <input
            id="budget-daily"
            type="number"
            className="form-input"
            value={daily}
            onChange={(e) => setDaily(parseInt(e.target.value) || 0)}
            min={0}
          />
        </div>
        <div style={{ display: "flex", gap: 8 }}>
          <button className="btn btn-sm btn-secondary" onClick={onCancel}>
            {tc("cancel")}
          </button>
          <button
            className="btn btn-sm btn-primary"
            disabled={isSubmitting}
            onClick={() => onSave(monthly, daily)}
          >
            {tc("save")}
          </button>
        </div>
      </div>
    </div>
  );
}

function ProviderTable({
  items,
  onToggle,
  onDelete,
}: {
  items: GatewayProviderKey[];
  onToggle: (id: string, current: boolean) => void;
  onDelete: (id: string) => void;
}) {
  const t = useTranslations("providers");
  const tc = useTranslations("common");
  const [confirmingId, setConfirmingId] = useState<string | null>(null);
  const confirmingProvider = items.find(p => p.id === confirmingId) || null;

  if (items.length === 0) {
    return (
      <EmptyState
        icon={<Server size={48} aria-hidden="true" />}
        title={t("noProviders")}
        description={t("noProvidersDesc")}
      />
    );
  }

  return (
    <>
      <ConfirmDialog
        isOpen={confirmingId !== null}
        onClose={() => setConfirmingId(null)}
        onConfirm={() => { onDelete(confirmingId!); setConfirmingId(null); }}
        title={tc("deleteConfirmTitle", { name: confirmingProvider?.provider || "" })}
        description={tc("deleteConfirmDescription")}
        confirmText={tc("confirm")}
        cancelText={tc("cancel")}
        variant="danger"
      />
      <div className="gateway-table-wrapper">
        <table className="gateway-table">
          <thead>
            <tr>
              <th style={{ width: 140 }}>{t("provider")}</th>
              <th>{t("modelPattern")}</th>
              <th style={{ width: 120 }}>{t("monthlyLimit")}</th>
              <th style={{ width: 90 }}>{tc("status")}</th>
              <th style={{ width: 100 }}>{tc("createdAt")}</th>
              <th style={{ width: 160 }}>{tc("actions")}</th>
            </tr>
          </thead>
          <tbody>
            {items.map((p) => (
              <tr key={p.id}>
                <td>
                  <span style={{ fontWeight: 600, fontSize: "var(--text-sm)" }}>{p.provider}</span>
                </td>
                <td>
                  <code
                    style={{
                      fontSize: "var(--text-xs)",
                      fontFamily: "'IBM Plex Mono', monospace",
                      color: "var(--ink)",
                    }}
                  >
                    {p.model || "*"}
                  </code>
                </td>
                <td>
                  <span style={{ fontSize: "var(--text-sm)", color: "var(--muted)" }}>
                    {p.monthly_limit > 0 ? p.monthly_limit.toLocaleString() : "—"}
                  </span>
                </td>
                <td>
                  <span
                    className={`badge ${p.is_enabled ? "badge-completed" : "badge-muted"}`}
                    style={{ display: "inline-flex", alignItems: "center", gap: 4 }}
                  >
                    {p.is_enabled ? t("enabled") : t("disabled")}
                  </span>
                </td>
                <td>
                  <span style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>
                    {new Date(p.created_at).toLocaleDateString()}
                  </span>
                </td>
                <td>
                  <div style={{ display: "flex", gap: 4, alignItems: "center" }}>
                    <button
                      className={`mcp-toggle-btn ${p.is_enabled ? "mcp-toggle-btn-on" : "mcp-toggle-btn-off"}`}
                      onClick={() => onToggle(p.id, p.is_enabled)}
                      title={p.is_enabled ? t("disable") : t("enable")}
                      aria-label={p.is_enabled ? t("disable") : t("enable")}
                    >
                      {p.is_enabled ? (
                        <Power size={14} aria-hidden="true" />
                      ) : (
                        <PowerOff size={14} aria-hidden="true" />
                      )}
                    </button>
                    <button
                      className="btn btn-sm btn-ghost"
                      onClick={() => setConfirmingId(p.id)}
                      aria-label={`Delete ${p.provider} key`}
                    >
                      <Trash2 size={14} aria-hidden="true" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}

function CreateProviderForm({
  onClose,
  onSuccess,
  isSubmitting,
}: {
  onClose: () => void;
  onSuccess: () => void;
  isSubmitting: boolean;
}) {
  const t = useTranslations("providers");
  const tc = useTranslations("common");
  const toast = useToast();
  const [provider, setProvider] = useState("openai");
  const [apiKey, setApiKey] = useState("");
  const [modelPattern, setModelPattern] = useState("");
  const [monthlyLimit, setMonthlyLimit] = useState(0);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!apiKey.trim()) return;
    try {
      await createProviderKey({
        provider,
        api_key: apiKey.trim(),
        model: modelPattern.trim() || undefined,
        monthly_limit: monthlyLimit,
      });
      toast.success(tc("success") + ": " + t("createProvider"));
      onSuccess();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create provider");
    }
  }

  return (
    <div className="panel">
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          padding: "12px 16px",
          borderBottom: "1px solid var(--line)",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <Key size={16} aria-hidden="true" style={{ color: "var(--muted)" }} />
          <h2 className="form-title" style={{ margin: 0 }}>
            {t("newProvider")}
          </h2>
        </div>
        <button className="btn btn-ghost btn-sm" onClick={onClose} aria-label={tc("close")}>
          <X size={18} aria-hidden="true" />
        </button>
      </div>
      <form className="form-fields" onSubmit={handleSubmit} style={{ padding: 16 }}>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr 1fr",
            gap: 12,
            alignItems: "end",
          }}
        >
          <div>
            <label className="form-label" htmlFor="provider-select">
              {t("provider")}
            </label>
            <select
              id="provider-select"
              className="form-select"
              value={provider}
              onChange={(e) => setProvider(e.target.value)}
            >
              {PROVIDERS.map((p) => (
                <option key={p.value} value={p.value}>
                  {p.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="form-label" htmlFor="api-key">
              {t("apiKey")}
            </label>
            <input
              id="api-key"
              type="password"
              className="form-input"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder={t("apiKeyPlaceholder")}
              required
            />
          </div>
          <div>
            <label className="form-label" htmlFor="model-pattern">
              {t("modelPattern")}
            </label>
            <input
              id="model-pattern"
              type="text"
              className="form-input"
              value={modelPattern}
              onChange={(e) => setModelPattern(e.target.value)}
              placeholder={t("modelPatternPlaceholder")}
            />
          </div>
        </div>
        <div
          style={{
            display: "flex",
            justifyContent: "flex-end",
            gap: 8,
            alignItems: "center",
          }}
        >
          <span style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>
            {t("monthlyLimit")}: {monthlyLimit.toLocaleString()}
          </span>
          <input
            type="range"
            min={0}
            max={10000000}
            step={10000}
            value={monthlyLimit}
            onChange={(e) => setMonthlyLimit(parseInt(e.target.value) || 0)}
            style={{ width: 120 }}
          />
          <button type="button" className="btn btn-secondary btn-sm" onClick={() => setMonthlyLimit(0)}>
            {t("unlimited")}
          </button>
          <button type="submit" className="btn btn-primary btn-sm" disabled={isSubmitting || !apiKey.trim()}>
            {isSubmitting ? t("creating") : t("createProvider")}
          </button>
        </div>
      </form>
    </div>
  );
}

export default function ProvidersPage() {
  const t = useTranslations("providers");
  const tc = useTranslations("common");
  const { isLoading: isAuthLoading } = useAuth();
  const toast = useToast();

  const [isLoading, setIsLoading] = useState(true);
  const [providers, setProviders] = useState<GatewayProviderKey[]>([]);
  const [budget, setBudget] = useState<UserBudget | null>(null);

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [budgetExpanded, setBudgetExpanded] = useState(true);
  const [budgetEditing, setBudgetEditing] = useState(false);
  const [isBudgetSubmitting, setIsBudgetSubmitting] = useState(false);

  async function loadData() {
    setIsLoading(true);
    try {
      const [providersResp, budgetResp] = await Promise.all([listProviderKeys(), getBudget()]);
      setProviders(providersResp.items);
      setBudget(budgetResp.item);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load data");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    if (!isAuthLoading) {
      void loadData();
    }
  }, [isAuthLoading]);

  async function handleCreateSuccess() {
    setShowCreateForm(false);
    await loadData();
  }

  async function handleToggle(id: string, currentEnabled: boolean) {
    try {
      await toggleProviderKey(id, !currentEnabled);
      toast.success(tc("success") + ": " + (!currentEnabled ? t("enable") : t("disable")));
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to toggle provider");
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteProviderKey(id);
      toast.success(tc("success") + ": Provider key deleted");
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to delete provider");
    }
  }

  async function handleBudgetSave(monthly: number, daily: number) {
    setIsBudgetSubmitting(true);
    try {
      await updateBudget({ monthly_limit: monthly, daily_limit: daily });
      toast.success(tc("success") + ": Budget updated");
      setBudgetEditing(false);
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to update budget");
    } finally {
      setIsBudgetSubmitting(false);
    }
  }

  if (isAuthLoading || isLoading) {
    return (
      <div className="admin-content">
        <Breadcrumb />
        <div className="loading-pulse" style={{ width: 200, height: 32, marginBottom: 16 }} />
        <div className="loading-pulse" style={{ width: "100%", height: 300 }} />
      </div>
    );
  }

  return (
    <div className="admin-content">
      <Breadcrumb />

      <div className="page-header">
        <div className="page-header-content">
          <p className="section-kicker">{t("eyebrow")}</p>
          <h1 className="panel-title">{t("title")}</h1>
          <p style={{ fontSize: "var(--text-sm)", color: "var(--muted)", marginTop: 4 }}>
            {t("lede")}
          </p>
        </div>
        <div className="page-header-actions">
          <button className="btn btn-primary" onClick={() => setShowCreateForm(true)}>
            <Plus size={16} aria-hidden="true" />
            {t("newProvider")}
          </button>
        </div>
      </div>

      {budgetEditing ? (
        <BudgetEditForm
          budget={budget}
          onSave={handleBudgetSave}
          onCancel={() => setBudgetEditing(false)}
          isSubmitting={isBudgetSubmitting}
        />
      ) : (
        <BudgetInline
          budget={budget}
          expanded={budgetExpanded}
          onToggle={() => setBudgetExpanded((v) => !v)}
          onEdit={() => setBudgetEditing(true)}
        />
      )}

      {showCreateForm && (
        <CreateProviderForm
          onClose={() => setShowCreateForm(false)}
          onSuccess={handleCreateSuccess}
          isSubmitting={isSubmitting}
        />
      )}

      <div className="panel" style={{ padding: 0 }}>
        <ProviderTable
          items={providers}
          onToggle={handleToggle}
          onDelete={handleDelete}
        />
      </div>
    </div>
  );
}
