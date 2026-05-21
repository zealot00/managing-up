"use client";

import { useEffect, useState } from "react";
import { useTranslations, useLocale } from "next-intl";
import { useToast } from "../../components/ToastProvider";
import Breadcrumb from "../../components/Breadcrumb";
import { getUserPreferences, updateUserPreferences, UserPreferences } from "../lib/user-api";
import { Globe, PanelLeftClose, PanelLeftOpen } from "lucide-react";
import { Skeleton } from "../components/layout/Skeleton";

export default function PreferencesPage() {
  const t = useTranslations("preferences");
  const toast = useToast();
  const currentLocale = useLocale();

  const [prefs, setPrefs] = useState<UserPreferences | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [language, setLanguage] = useState(currentLocale);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    async function load() {
      try {
        const p = await getUserPreferences();
        setPrefs(p);
        setLanguage(p.language);
        setSidebarCollapsed(p.sidebar_collapsed);
      } catch {
        // use defaults
      } finally {
        setIsLoading(false);
      }
    }
    void load();
  }, []);

  async function handleLanguageChange(lang: string) {
    setLanguage(lang);
    setIsSaving(true);
    try {
      const updated = await updateUserPreferences({ language: lang });
      setPrefs(updated);
      // Update NEXT_LOCALE cookie for next-intl
      document.cookie = `NEXT_LOCALE=${lang};path=/;max-age=${60 * 60 * 24 * 365}`;
      // Reload to apply new locale
      window.location.reload();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("saveFailed"));
      setLanguage(prefs?.language || currentLocale);
    } finally {
      setIsSaving(false);
    }
  }

  async function handleSidebarToggle() {
    const newValue = !sidebarCollapsed;
    setSidebarCollapsed(newValue);
    setIsSaving(true);
    try {
      const updated = await updateUserPreferences({ sidebar_collapsed: newValue });
      setPrefs(updated);
      toast.success(t("saved"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("saveFailed"));
      setSidebarCollapsed(!newValue);
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <>
      <Breadcrumb />
      <div>
        <div className="section-kicker">{t("eyebrow")}</div>
        <h1 className="panel-title">{t("title")}</h1>
        <p style={{ color: "var(--muted)", fontSize: "var(--text-sm)", marginBottom: "var(--space-6)" }}>
          {t("lede")}
        </p>
      </div>

      {isLoading ? (
        <>
          <div className="panel" style={{ marginBottom: "var(--space-6)" }}>
            <div style={{ padding: "var(--space-4)", borderBottom: "1px solid var(--line)" }}>
              <Skeleton width={100} height={16} />
            </div>
            <div style={{ padding: "var(--space-4)" }}>
              <Skeleton width={280} height={14} style={{ marginBottom: "var(--space-3)" }} />
              <div style={{ display: "flex", gap: "var(--space-3)" }}>
                <Skeleton width={64} height={36} borderRadius="var(--radius-sm)" />
                <Skeleton width={64} height={36} borderRadius="var(--radius-sm)" />
              </div>
            </div>
          </div>
          <div className="panel">
            <div style={{ padding: "var(--space-4)", borderBottom: "1px solid var(--line)" }}>
              <Skeleton width={100} height={16} />
            </div>
            <div style={{ padding: "var(--space-4)" }}>
              <Skeleton width={200} height={14} style={{ marginBottom: "var(--space-2)" }} />
              <Skeleton width={44} height={24} borderRadius="12px" />
            </div>
          </div>
        </>
      ) : (
        <>
          {/* Language */}
          <div className="panel" style={{ marginBottom: "var(--space-6)" }}>
            <div style={{ padding: "var(--space-4)", borderBottom: "1px solid var(--line)" }}>
              <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
                <Globe size={18} style={{ color: "var(--muted)" }} />
                <h2 style={{ fontSize: "var(--text-sm)", fontWeight: 600, margin: 0 }}>{t("language")}</h2>
              </div>
            </div>
            <div style={{ padding: "var(--space-4)" }}>
              <p style={{ fontSize: "var(--text-sm)", color: "var(--muted)", marginBottom: "var(--space-3)" }}>
                {t("languageDesc")}
              </p>
              <div style={{ display: "flex", gap: "var(--space-3)" }}>
                <button
                  className={`btn ${language === "en" ? "btn-primary" : "btn-secondary"}`}
                  onClick={() => handleLanguageChange("en")}
                  disabled={isSaving}
                >
                  {t("langEN")}
                </button>
                <button
                  className={`btn ${language === "zh" ? "btn-primary" : "btn-secondary"}`}
                  onClick={() => handleLanguageChange("zh")}
                  disabled={isSaving}
                >
                  {t("langZH")}
                </button>
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="panel">
            <div style={{ padding: "var(--space-4)", borderBottom: "1px solid var(--line)" }}>
              <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
                {sidebarCollapsed ? (
                  <PanelLeftOpen size={18} style={{ color: "var(--muted)" }} />
                ) : (
                  <PanelLeftClose size={18} style={{ color: "var(--muted)" }} />
                )}
                <h2 style={{ fontSize: "var(--text-sm)", fontWeight: 600, margin: 0 }}>{t("sidebar")}</h2>
              </div>
            </div>
            <div style={{ padding: "var(--space-4)" }}>
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                }}
              >
                <div>
                  <div style={{ fontWeight: 500, fontSize: "var(--text-sm)" }}>{t("sidebarCollapsed")}</div>
                  <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("sidebarCollapsedDesc")}</div>
                </div>
                <button
                  className={`toggle ${sidebarCollapsed ? "toggle-active" : ""}`}
                  onClick={handleSidebarToggle}
                  disabled={isSaving}
                  role="switch"
                  aria-checked={sidebarCollapsed}
                  style={{
                    width: 44,
                    height: 24,
                    borderRadius: 12,
                    border: "none",
                    cursor: isSaving ? "not-allowed" : "pointer",
                    background: sidebarCollapsed ? "var(--primary)" : "var(--line-strong)",
                    position: "relative",
                    transition: "background var(--transition-fast)",
                    flexShrink: 0,
                  }}
                >
                  <span
                    style={{
                      position: "absolute",
                      top: 2,
                      left: sidebarCollapsed ? 22 : 2,
                      width: 20,
                      height: 20,
                      borderRadius: 10,
                      background: "white",
                      transition: "left var(--transition-fast)",
                      boxShadow: "var(--shadow-sm)",
                    }}
                  />
                </button>
              </div>
            </div>
          </div>
        </>
      )}
    </>
  );
}
