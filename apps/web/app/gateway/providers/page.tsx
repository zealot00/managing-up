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

export default function ProvidersPage() {
  const t = useTranslations("providers");
  const tc = useTranslations("common");
  const { isLoading: isAuthLoading } = useAuth();
  const toast = useToast();

  const [isLoading, setIsLoading] = useState(true);
  const [providers, setProviders] = useState<GatewayProviderKey[]>([]);
  const [budget, setBudget] = useState<UserBudget | null>(null);

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [provider, setProvider] = useState("openai");
  const [apiKey, setApiKey] = useState("");
  const [modelPattern, setModelPattern] = useState("");
  const [monthlyLimit, setMonthlyLimit] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

  const [showBudgetForm, setShowBudgetForm] = useState(false);
  const [budgetMonthlyLimit, setBudgetMonthlyLimit] = useState(0);
  const [budgetDailyLimit, setBudgetDailyLimit] = useState(0);
  const [isBudgetSubmitting, setIsBudgetSubmitting] = useState(false);

  async function loadData() {
    setIsLoading(true);
    try {
      const [providersResp, budgetResp] = await Promise.all([
        listProviderKeys(),
        getBudget(),
      ]);
      setProviders(providersResp.items);
      setBudget(budgetResp.item);
      setBudgetMonthlyLimit(budgetResp.item.monthly_limit);
      setBudgetDailyLimit(budgetResp.item.daily_limit);
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

  async function handleCreate(e: FormEvent<HTMLFormElement>) {
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
      toast.success(t("createProvider") + " " + tc("success"));
      setShowCreateForm(false);
      setApiKey("");
      setModelPattern("");
      setMonthlyLimit(0);
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create provider");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleToggle(id: string, currentEnabled: boolean) {
    try {
      await toggleProviderKey(id, !currentEnabled);
      toast.success(!currentEnabled ? t("enable") : t("disable") + " " + tc("success"));
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to toggle provider");
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteProviderKey(id);
      toast.success(t("deleteProvider") + " " + tc("success"));
      setDeleteConfirmId(null);
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to delete provider");
    }
  }

  async function handleBudgetUpdate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setIsBudgetSubmitting(true);
    try {
      await updateBudget({
        monthly_limit: budgetMonthlyLimit,
        daily_limit: budgetDailyLimit,
      });
      toast.success(tc("save") + " " + tc("success"));
      setShowBudgetForm(false);
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to update budget");
    } finally {
      setIsBudgetSubmitting(false);
    }
  }

  if (isAuthLoading || isLoading) {
    return (
      <div className="gateway-page">
        <div className="loading-pulse" style={{ width: 200, height: 32, marginBottom: 16 }} />
        <div className="loading-pulse" style={{ width: "100%", height: 400 }} />
      </div>
    );
  }

  return (
    <div className="gateway-page">
      <div className="dashboard-header">
        <div>
          <p className="eyebrow">{t("eyebrow")}</p>
          <h1 className="page-title">{t("title")}</h1>
          <p className="page-lede">{t("lede")}</p>
        </div>
      </div>

      <div className="dashboard-section">
        <div className="dashboard-section-header">
          <h2 className="dashboard-section-title">{t("budget")}</h2>
          <button
            className="form-submit"
            onClick={() => setShowBudgetForm(!showBudgetForm)}
          >
            {showBudgetForm ? tc("cancel") : tc("edit")}
          </button>
        </div>

        {showBudgetForm ? (
          <form className="form-fields" onSubmit={handleBudgetUpdate}>
            <div>
              <label className="form-label" htmlFor="budget-monthly">
                {t("monthlyLimit")}
              </label>
              <input
                id="budget-monthly"
                type="number"
                className="form-input"
                value={budgetMonthlyLimit}
                onChange={(e) => setBudgetMonthlyLimit(parseInt(e.target.value) || 0)}
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
                value={budgetDailyLimit}
                onChange={(e) => setBudgetDailyLimit(parseInt(e.target.value) || 0)}
                min={0}
              />
            </div>
            <button className="form-submit" type="submit" disabled={isBudgetSubmitting}>
              {isBudgetSubmitting ? tc("loading") : tc("save")}
            </button>
          </form>
        ) : (
          <div className="budget-display">
            <div className="budget-stat">
              <span className="budget-stat-label">{t("usedThisMonth")}</span>
              <span className="budget-stat-value">
                {budget?.used_this_month.toLocaleString() || 0}
              </span>
            </div>
            <div className="budget-stat">
              <span className="budget-stat-label">{t("monthlyLimit")}</span>
              <span className="budget-stat-value">
                {budget?.monthly_limit.toLocaleString() || 0}
              </span>
            </div>
            <div className="budget-stat">
              <span className="budget-stat-label">{t("dailyLimit")}</span>
              <span className="budget-stat-value">
                {budget?.daily_limit.toLocaleString() || 0}
              </span>
            </div>
          </div>
        )}
      </div>

      <div className="dashboard-section">
        <div className="dashboard-section-header">
          <h2 className="dashboard-section-title">{t("title")}</h2>
          <button
            className="form-submit"
            onClick={() => setShowCreateForm(!showCreateForm)}
          >
            {showCreateForm ? tc("cancel") : t("newProvider")}
          </button>
        </div>

        {showCreateForm && (
          <form className="form-fields" onSubmit={handleCreate}>
            <div>
              <label className="form-label" htmlFor="provider-select">
                {t("provider")}
              </label>
              <select
                id="provider-select"
                className="form-input"
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
            <div>
              <label className="form-label" htmlFor="monthly-limit">
                {t("monthlyLimit")}
              </label>
              <input
                id="monthly-limit"
                type="number"
                className="form-input"
                value={monthlyLimit}
                onChange={(e) => setMonthlyLimit(parseInt(e.target.value) || 0)}
                min={0}
                placeholder={t("monthlyLimitPlaceholder")}
              />
            </div>
            <button className="form-submit" type="submit" disabled={isSubmitting}>
              {isSubmitting ? t("creating") : t("createProvider")}
            </button>
          </form>
        )}

        {providers.length === 0 ? (
          <div className="empty-state">
            <p className="empty-state-title">{t("noProviders")}</p>
            <p className="empty-state-desc">{t("noProvidersDesc")}</p>
          </div>
        ) : (
          <div className="gateway-table-wrapper">
            <table className="gateway-table">
              <thead>
                <tr>
                  <th>{t("provider")}</th>
                  <th>{t("modelPattern")}</th>
                  <th>{t("monthlyLimit")}</th>
                  <th>{tc("status")}</th>
                  <th>{tc("createdAt")}</th>
                  <th>{tc("actions")}</th>
                </tr>
              </thead>
              <tbody>
                {providers.map((p) => (
                  <tr key={p.id}>
                    <td>{p.provider}</td>
                    <td>{p.model || "*"}</td>
                    <td>{p.monthly_limit.toLocaleString()}</td>
                    <td>
                      <span className={`badge ${p.is_enabled ? "badge-completed" : "badge-failed"}`}>
                        {p.is_enabled ? t("enabled") : t("disabled")}
                      </span>
                    </td>
                    <td>{new Date(p.created_at).toLocaleDateString()}</td>
                    <td>
                      <div className="table-actions">
                        <button
                          className="gateway-button-secondary"
                          onClick={() => void handleToggle(p.id, p.is_enabled)}
                        >
                          {p.is_enabled ? t("disable") : t("enable")}
                        </button>
                        {deleteConfirmId === p.id ? (
                          <div className="confirm-delete">
                            <span>{t("confirmDelete")}</span>
                            <button
                              className="gateway-button-danger"
                              onClick={() => void handleDelete(p.id)}
                            >
                              {tc("confirm")}
                            </button>
                            <button
                              className="gateway-button-secondary"
                              onClick={() => setDeleteConfirmId(null)}
                            >
                              {tc("cancel")}
                            </button>
                          </div>
                        ) : (
                          <button
                            className="gateway-button-danger"
                            onClick={() => setDeleteConfirmId(p.id)}
                          >
                            {tc("delete")}
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}