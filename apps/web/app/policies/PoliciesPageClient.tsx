"use client";

import { useState, FormEvent } from "react";
import { useTranslations } from "next-intl";
import { useToast } from "../../components/ToastProvider";
import Breadcrumb from "../../components/Breadcrumb";
import {
  PolicyVersion,
  PolicyRule,
  getPolicies,
  getPolicy,
  createPolicy,
  updatePolicy,
} from "../lib/api";
import {
  Shield,
  Plus,
  X,
  ChevronDown,
  ChevronRight,
  Save,
  Trash2,
} from "lucide-react";

export default function PoliciesPageClient() {
  const t = useTranslations("policies");
  const tc = useTranslations("common");
  const toast = useToast();

  const [policies, setPolicies] = useState<PolicyVersion[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [selectedPolicy, setSelectedPolicy] = useState<PolicyVersion | null>(null);

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [formName, setFormName] = useState("");
  const [formVersion, setFormVersion] = useState("v1");
  const [formDescription, setFormDescription] = useState("");
  const [formRules, setFormRules] = useState<PolicyRule[]>([]);

  async function loadPolicies() {
    setError(null);
    setIsLoading(true);
    try {
      const resp = await getPolicies();
      setPolicies(resp.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load policies");
    } finally {
      setIsLoading(false);
    }
  }

  useState(() => {
    void loadPolicies();
  });

  async function loadPolicyDetail(id: string) {
    try {
      const policy = await getPolicy(id);
      setSelectedPolicy(policy);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load policy");
    }
  }

  function toggleExpanded(id: string) {
    if (expandedId === id) {
      setExpandedId(null);
      setSelectedPolicy(null);
    } else {
      setExpandedId(id);
      void loadPolicyDetail(id);
    }
  }

  function resetForm() {
    setFormName("");
    setFormVersion("v1");
    setFormDescription("");
    setFormRules([]);
    setShowCreateForm(false);
  }

  function addRule() {
    setFormRules([
      ...formRules,
      {
        id: `rule_${Date.now()}`,
        version: formVersion,
        condition: "",
        action: "allow",
        reason: "",
        priority: formRules.length + 1,
        is_active: true,
      },
    ]);
  }

  function updateRule(index: number, updates: Partial<PolicyRule>) {
    const newRules = [...formRules];
    newRules[index] = { ...newRules[index], ...updates };
    setFormRules(newRules);
  }

  function removeRule(index: number) {
    setFormRules(formRules.filter((_, i) => i !== index));
  }

  async function handleCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!formName.trim()) return;

    setIsSubmitting(true);
    setError(null);
    try {
      await createPolicy({
        name: formName.trim(),
        version: formVersion || "v1",
        description: formDescription.trim(),
        rules: formRules,
      });
      toast.success(tc("success") + ": Policy created");
      resetForm();
      await loadPolicies();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create policy");
    } finally {
      setIsSubmitting(false);
    }
  }

  if (isLoading) {
    return (
      <div className="admin-content">
        <Breadcrumb />
        <div className="loading-pulse loading-pulse-long" style={{ marginBottom: 16 }} />
        <div className="skeleton-grid">
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
            {t("newPolicy")}
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
              <h2 className="form-title">{t("createPolicy")}</h2>
            </div>
            <button
              className="btn btn-ghost btn-sm"
              onClick={resetForm}
              aria-label={tc("close")}
            >
              <X size={18} aria-hidden="true" />
            </button>
          </div>
          <form className="form-fields" onSubmit={(e) => void handleCreate(e)}>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
              <div>
                <label className="form-label" htmlFor="policy-name">{t("policyName")}</label>
                <input
                  id="policy-name"
                  className="form-input"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  placeholder={t("policyNamePlaceholder")}
                  required
                  autoFocus
                />
              </div>
              <div>
                <label className="form-label" htmlFor="policy-version">{t("version")}</label>
                <input
                  id="policy-version"
                  className="form-input"
                  value={formVersion}
                  onChange={(e) => setFormVersion(e.target.value)}
                  placeholder="v1"
                />
              </div>
            </div>
            <div>
              <label className="form-label" htmlFor="policy-desc">{tc("description")}</label>
              <input
                id="policy-desc"
                className="form-input"
                value={formDescription}
                onChange={(e) => setFormDescription(e.target.value)}
                placeholder={t("policyDescPlaceholder")}
              />
            </div>

            <div>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
                <label className="form-label" style={{ marginBottom: 0 }}>{t("rules")}</label>
                <button type="button" className="btn btn-sm btn-secondary" onClick={addRule}>
                  <Plus size={14} aria-hidden="true" /> {t("addRule")}
                </button>
              </div>
              {formRules.length === 0 ? (
                <p style={{ fontSize: "var(--text-xs)", color: "var(--muted)", padding: 12, background: "var(--surface)", borderRadius: 4 }}>
                  {t("noRulesHint")}
                </p>
              ) : (
                <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                  {formRules.map((rule, index) => (
                    <div key={rule.id} style={{ display: "grid", gridTemplateColumns: "1fr auto 1fr auto", gap: 8, alignItems: "center", padding: 8, background: "var(--surface)", borderRadius: 4 }}>
                      <input
                        className="form-input"
                        value={rule.condition}
                        onChange={(e) => updateRule(index, { condition: e.target.value })}
                        placeholder={t("conditionPlaceholder")}
                      />
                      <select
                        className="form-select"
                        value={rule.action}
                        onChange={(e) => updateRule(index, { action: e.target.value })}
                        style={{ width: "auto" }}
                      >
                        <option value="allow">{t("actionAllow")}</option>
                        <option value="deny">{t("actionDeny")}</option>
                        <option value="flag">{t("actionFlag")}</option>
                      </select>
                      <input
                        className="form-input"
                        value={rule.reason}
                        onChange={(e) => updateRule(index, { reason: e.target.value })}
                        placeholder={t("reasonPlaceholder")}
                      />
                      <button type="button" className="btn btn-ghost btn-sm" onClick={() => removeRule(index)}>
                        <Trash2 size={14} aria-hidden="true" />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>

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

      {policies.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state-icon">
            <Shield size={48} aria-hidden="true" />
          </div>
          <h3 className="empty-state-title">{t("noPolicies")}</h3>
          <p className="empty-state-description">
            {t("noPoliciesDescription")}
          </p>
          <div className="empty-state-action" style={{ marginTop: 16 }}>
            <button className="btn btn-primary" onClick={() => setShowCreateForm(true)}>
              <Plus size={16} aria-hidden="true" />
              {t("newPolicy")}
            </button>
          </div>
        </div>
      ) : (
        <div className="panel">
          <div className="gateway-table-wrapper">
            <table className="gateway-table">
              <thead>
                <tr>
                  <th style={{ width: 40 }}></th>
                  <th>{t("name")}</th>
                  <th>{t("version")}</th>
                  <th>{t("rules")}</th>
                  <th style={{ width: 100 }}>{t("status")}</th>
                </tr>
              </thead>
              <tbody>
                {policies.map((policy) => {
                  const isExpanded = expandedId === policy.id;
                  const detail = isExpanded ? selectedPolicy : null;

                  return (
                    <>
                      <tr key={policy.id}>
                        <td>
                          <button
                            className="btn btn-ghost btn-sm"
                            onClick={() => toggleExpanded(policy.id)}
                          >
                            {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                          </button>
                        </td>
                        <td>
                          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                            <Shield size={16} aria-hidden="true" style={{ color: "var(--muted)", flexShrink: 0 }} />
                            <div>
                              <div style={{ fontWeight: 600, color: "var(--ink-strong)", fontSize: "var(--text-sm)" }}>
                                {policy.name}
                              </div>
                              {policy.description && (
                                <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)", marginTop: 2 }}>
                                  {policy.description}
                                </div>
                              )}
                            </div>
                          </div>
                        </td>
                        <td>
                          <code style={{ fontSize: "var(--text-xs)" }}>{policy.version}</code>
                        </td>
                        <td>
                          <span className="badge badge-muted">{policy.rules?.length || 0} rules</span>
                        </td>
                        <td>
                          {policy.is_default ? (
                            <span className="badge badge-completed">{t("default")}</span>
                          ) : (
                            <span className="badge badge-muted">{t("inactive")}</span>
                          )}
                        </td>
                      </tr>
                      {isExpanded && detail && (
                        <tr key={`${policy.id}-detail`}>
                          <td colSpan={5} style={{ padding: 16, background: "var(--background-subtle)" }}>
                            <h4 style={{ fontSize: "var(--text-sm)", fontWeight: 600, marginBottom: 12 }}>
                              {t("rules")}:
                            </h4>
                            {detail.rules && detail.rules.length > 0 ? (
                              <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
                                {detail.rules.map((rule) => (
                                  <div key={rule.id} style={{ display: "flex", gap: 12, alignItems: "center", fontSize: "var(--text-xs)", padding: 6, background: "var(--surface)", borderRadius: 4 }}>
                                    <span style={{ fontWeight: 500, minWidth: 60 }}>{rule.action}</span>
                                    <code style={{ flex: 1, fontFamily: "'IBM Plex Mono', monospace" }}>{rule.condition}</code>
                                    <span style={{ color: "var(--muted)" }}>{rule.reason}</span>
                                    <span style={{ color: rule.is_active ? "var(--success)" : "var(--muted)" }}>
                                      {rule.is_active ? t("active") : t("inactive")}
                                    </span>
                                  </div>
                                ))}
                              </div>
                            ) : (
                              <p style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("noRules")}</p>
                            )}
                          </td>
                        </tr>
                      )}
                    </>
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