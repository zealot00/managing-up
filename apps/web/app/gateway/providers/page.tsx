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
import { Drawer } from "../../components/ui/Drawer";
import { PasswordInput } from "../../components/ui/PasswordInput";
import {
  Server,
  Plus,
  Power,
  PowerOff,
  Trash2,
  Key,
  Calculator,
  Edit3,
  Loader2,
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

function BudgetBar({
  budget,
  onEdit,
}: {
  budget: UserBudget | null;
  onEdit: () => void;
}) {
  const t = useTranslations("providers");

  return (
    <div className="budget-bar">
      <div className="budget-bar-stats">
        <Calculator size={14} aria-hidden="true" className="budget-bar-icon" />
        <span className="budget-bar-label">{t("budget")}</span>
        <span className="budget-bar-stat">
          {t("usedThisMonth")} <strong>{budget?.used_this_month.toLocaleString() || 0}</strong>
        </span>
        <span className="budget-bar-separator">·</span>
        <span className="budget-bar-stat">
          {t("monthlyLimitBudget")} <strong>{budget?.monthly_limit.toLocaleString() || 0}</strong>
        </span>
        <span className="budget-bar-separator">·</span>
        <span className="budget-bar-stat">
          {t("dailyLimit")} <strong>{budget?.daily_limit.toLocaleString() || 0}</strong>
        </span>
      </div>
      <button className="btn btn-sm btn-secondary" onClick={onEdit}>
        <Edit3 size={14} aria-hidden="true" />
        {t("budget")}
      </button>
    </div>
  );
}

function ProviderTable({
  items,
  onToggle,
  onDelete,
  togglingId,
}: {
  items: GatewayProviderKey[];
  onToggle: (id: string, current: boolean) => void;
  onDelete: (id: string) => void;
  togglingId: string | null;
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
              <th scope="col">{t("modelPattern")}</th>
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
                      disabled={togglingId === p.id}
                      title={p.is_enabled ? t("disable") : t("enable")}
                      aria-label={p.is_enabled ? t("disable") : t("enable")}
                      style={{ opacity: togglingId === p.id ? 0.5 : 1, cursor: togglingId === p.id ? "not-allowed" : "pointer" }}
                    >
                      {togglingId === p.id ? (
                        <Loader2 size={14} className="animate-spin" aria-hidden="true" />
                      ) : p.is_enabled ? (
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

export default function ProvidersPage() {
  const t = useTranslations("providers");
  const tc = useTranslations("common");
  const { isLoading: isAuthLoading } = useAuth();
  const toast = useToast();

  const [isLoading, setIsLoading] = useState(true);
  const [providers, setProviders] = useState<GatewayProviderKey[]>([]);
  const [budget, setBudget] = useState<UserBudget | null>(null);

  const [showCreateDrawer, setShowCreateDrawer] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [togglingId, setTogglingId] = useState<string | null>(null);

  const [showBudgetDrawer, setShowBudgetDrawer] = useState(false);
  const [isBudgetSubmitting, setIsBudgetSubmitting] = useState(false);

  // Create form state
  const [provider, setProvider] = useState("openai");
  const [apiKey, setApiKey] = useState("");
  const [modelPattern, setModelPattern] = useState("");
  const [monthlyLimit, setMonthlyLimit] = useState(0);

  // Budget edit state
  const [budgetMonthly, setBudgetMonthly] = useState(0);
  const [budgetDaily, setBudgetDaily] = useState(0);

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

  async function handleCreateSubmit(e: FormEvent) {
    e.preventDefault();
    if (!apiKey.trim()) return;
    setIsSubmitting(true);
    try {
      await createProviderKey({
        provider,
        api_key: apiKey.trim(),
        model: modelPattern.trim() || undefined,
        monthly_limit: monthlyLimit,
      });
      toast.success(tc("success") + ": " + t("createProvider"));
      setShowCreateDrawer(false);
      resetCreateForm();
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create provider");
    } finally {
      setIsSubmitting(false);
    }
  }

  function resetCreateForm() {
    setProvider("openai");
    setApiKey("");
    setModelPattern("");
    setMonthlyLimit(0);
  }

  function handleOpenBudgetDrawer() {
    setBudgetMonthly(budget?.monthly_limit ?? 0);
    setBudgetDaily(budget?.daily_limit ?? 0);
    setShowBudgetDrawer(true);
  }

  async function handleToggle(id: string, currentEnabled: boolean) {
    setTogglingId(id);
    try {
      await toggleProviderKey(id, !currentEnabled);
      toast.success(tc("success") + ": " + (!currentEnabled ? t("enable") : t("disable")));
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to toggle provider");
    } finally {
      setTogglingId(null);
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

  async function handleBudgetSave() {
    setIsBudgetSubmitting(true);
    try {
      await updateBudget({ monthly_limit: budgetMonthly, daily_limit: budgetDaily });
      toast.success(tc("success") + ": Budget updated");
      setShowBudgetDrawer(false);
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
          <button className="btn btn-primary" onClick={() => setShowCreateDrawer(true)}>
            <Plus size={16} aria-hidden="true" />
            {t("newProvider")}
          </button>
        </div>
      </div>

      <BudgetBar budget={budget} onEdit={handleOpenBudgetDrawer} />

      <div className="panel" style={{ padding: 0 }}>
        <ProviderTable
          items={providers}
          onToggle={handleToggle}
          onDelete={handleDelete}
          togglingId={togglingId}
        />
      </div>

      {/* Create Provider Drawer */}
      <Drawer
        isOpen={showCreateDrawer}
        onClose={() => { setShowCreateDrawer(false); resetCreateForm(); }}
        title={t("newProvider")}
        description={t("lede")}
      >
        <form onSubmit={handleCreateSubmit}>
          <div style={{ display: "grid", gap: "var(--space-4)" }}>
            <div>
              <label className="form-label" htmlFor="provider-select">{t("provider")}</label>
              <select
                id="provider-select"
                className="form-select"
                value={provider}
                onChange={(e) => setProvider(e.target.value)}
              >
                {PROVIDERS.map((p) => (
                  <option key={p.value} value={p.value}>{p.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="form-label" htmlFor="api-key">{t("apiKey")}</label>
              <PasswordInput
                id="api-key"
                inputClassName="form-input"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder={t("apiKeyPlaceholder")}
                required
              />
            </div>
            <div>
              <label className="form-label" htmlFor="model-pattern">{t("modelPattern")}</label>
              <input
                id="model-pattern"
                type="text"
                className="form-input"
                value={modelPattern}
                onChange={(e) => setModelPattern(e.target.value)}
                placeholder={t("modelPatternPlaceholder")}
              />
            </div>
            <div>
              <label className="form-label">{t("monthlyLimit")}</label>
              <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                <input
                  type="range"
                  min={0}
                  max={10000000}
                  step={10000}
                  value={monthlyLimit}
                  onChange={(e) => setMonthlyLimit(parseInt(e.target.value) || 0)}
                  style={{ flex: 1 }}
                />
                <span style={{ fontSize: "var(--text-sm)", fontWeight: 600, minWidth: 80, textAlign: "right" }}>
                  {monthlyLimit > 0 ? monthlyLimit.toLocaleString() : t("unlimited")}
                </span>
              </div>
            </div>
            <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, paddingTop: "var(--space-4)" }}>
              <button
                type="button"
                className="btn btn-secondary"
                onClick={() => { setShowCreateDrawer(false); resetCreateForm(); }}
              >
                {tc("cancel")}
              </button>
              <button
                type="submit"
                className="btn btn-primary"
                disabled={isSubmitting || !apiKey.trim()}
              >
                {isSubmitting ? t("creating") : t("createProvider")}
              </button>
            </div>
          </div>
        </form>
      </Drawer>

      {/* Budget Edit Drawer */}
      <Drawer
        isOpen={showBudgetDrawer}
        onClose={() => setShowBudgetDrawer(false)}
        title={t("budget")}
      >
        <div style={{ display: "grid", gap: "var(--space-4)" }}>
          <div>
            <label className="form-label" htmlFor="budget-monthly">{t("monthlyLimit")}</label>
            <input
              id="budget-monthly"
              type="number"
              className="form-input"
              value={budgetMonthly}
              onChange={(e) => setBudgetMonthly(parseInt(e.target.value) || 0)}
              min={0}
            />
          </div>
          <div>
            <label className="form-label" htmlFor="budget-daily">{t("dailyLimit")}</label>
            <input
              id="budget-daily"
              type="number"
              className="form-input"
              value={budgetDaily}
              onChange={(e) => setBudgetDaily(parseInt(e.target.value) || 0)}
              min={0}
            />
          </div>
          <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, paddingTop: "var(--space-4)" }}>
            <button
              type="button"
              className="btn btn-secondary"
              onClick={() => setShowBudgetDrawer(false)}
            >
              {tc("cancel")}
            </button>
            <button
              type="button"
              className="btn btn-primary"
              disabled={isBudgetSubmitting}
              onClick={handleBudgetSave}
            >
              {isBudgetSubmitting ? tc("loading") : tc("save")}
            </button>
          </div>
        </div>
      </Drawer>
    </div>
  );
}
