"use client";

import { FormEvent, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useAuth } from "../../context/AuthContext";
import { useToast } from "../../components/ToastProvider";
import Breadcrumb from "../../components/Breadcrumb";
import { getUserProfile, changePassword, UserProfile } from "../lib/user-api";
import { User, Lock, Shield, Calendar } from "lucide-react";
import { Skeleton } from "../components/layout/Skeleton";
import { PasswordInput } from "../components/ui/PasswordInput";

export default function ProfilePage() {
  const t = useTranslations("profile");
  const tc = useTranslations("common");
  const toast = useToast();
  const { user } = useAuth();

  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    async function load() {
      try {
        const p = await getUserProfile();
        setProfile(p);
      } catch {
        // fallback to auth context data
      } finally {
        setIsLoading(false);
      }
    }
    void load();
  }, []);

  async function handlePasswordChange(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (newPassword !== confirmPassword) {
      toast.error(t("passwordMismatch"));
      return;
    }

    if (newPassword.length < 6) {
      toast.error(t("newPasswordPlaceholder"));
      return;
    }

    setIsSubmitting(true);
    try {
      await changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      });
      toast.success(t("passwordUpdated"));
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err) {
      const msg = err instanceof Error ? err.message : t("passwordUpdateFailed");
      if (msg.toLowerCase().includes("incorrect") || msg.toLowerCase().includes("invalid password")) {
        toast.error(t("incorrectPassword"));
      } else {
        toast.error(msg);
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  const displayProfile = profile || (user ? { id: user.id, username: user.username, role: user.role, created_at: "" } : null);

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

      {/* User Information */}
      <div className="panel" style={{ marginBottom: "var(--space-6)" }}>
        <div style={{ padding: "var(--space-4)", borderBottom: "1px solid var(--line)" }}>
          <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
            <User size={18} style={{ color: "var(--muted)" }} />
            <h2 style={{ fontSize: "var(--text-sm)", fontWeight: 600, margin: 0 }}>{t("userInfo")}</h2>
          </div>
        </div>
        <div style={{ padding: "var(--space-4)" }}>
          {isLoading ? (
            <div style={{ display: "grid", gap: "var(--space-4)" }}>
              <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
                <Skeleton width={16} height={16} />
                <div>
                  <Skeleton width={80} height={12} style={{ marginBottom: 4 }} />
                  <Skeleton width={120} height={16} />
                </div>
              </div>
              <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
                <Skeleton width={16} height={16} />
                <div>
                  <Skeleton width={40} height={12} style={{ marginBottom: 4 }} />
                  <Skeleton width={60} height={20} borderRadius="var(--radius-sm)" />
                </div>
              </div>
            </div>
          ) : displayProfile ? (
            <div style={{ display: "grid", gap: "var(--space-4)" }}>
              <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
                <User size={16} style={{ color: "var(--muted)" }} />
                <div>
                  <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("username")}</div>
                  <div style={{ fontWeight: 500 }}>{displayProfile.username}</div>
                </div>
              </div>
              <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
                <Shield size={16} style={{ color: "var(--muted)" }} />
                <div>
                  <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("role")}</div>
                  <span className="badge badge-info">{displayProfile.role}</span>
                </div>
              </div>
              {displayProfile.created_at && (
                <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
                  <Calendar size={16} style={{ color: "var(--muted)" }} />
                  <div>
                    <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("createdAt")}</div>
                    <div>{new Date(displayProfile.created_at).toLocaleDateString()}</div>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <p style={{ color: "var(--danger)" }}>{tc("noData")}</p>
          )}
        </div>
      </div>

      {/* Change Password */}
      <div className="panel">
        <div style={{ padding: "var(--space-4)", borderBottom: "1px solid var(--line)" }}>
          <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
            <Lock size={18} style={{ color: "var(--muted)" }} />
            <h2 style={{ fontSize: "var(--text-sm)", fontWeight: 600, margin: 0 }}>{t("changePassword")}</h2>
          </div>
        </div>
        <form onSubmit={handlePasswordChange} style={{ padding: "var(--space-4)" }}>
          <div style={{ display: "grid", gap: "var(--space-4)", maxWidth: 400 }}>
            <div>
              <label className="form-label" htmlFor="current-password">{t("currentPassword")}</label>
              <PasswordInput
                id="current-password"
                inputClassName="form-input"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder={t("currentPasswordPlaceholder")}
                required
              />
            </div>
            <div>
              <label className="form-label" htmlFor="new-password">{t("newPassword")}</label>
              <PasswordInput
                id="new-password"
                inputClassName="form-input"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder={t("newPasswordPlaceholder")}
                required
                minLength={6}
              />
              {newPassword && newPassword.length < 6 && (
                <span className="form-hint" style={{ color: "var(--warning)" }}>
                  {t("newPasswordPlaceholder")}
                </span>
              )}
            </div>
            <div>
              <label className="form-label" htmlFor="confirm-password">{t("confirmPassword")}</label>
              <PasswordInput
                id="confirm-password"
                inputClassName="form-input"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder={t("confirmPasswordPlaceholder")}
                required
                minLength={6}
                style={confirmPassword && newPassword !== confirmPassword ? { borderColor: "var(--danger)" } : undefined}
              />
              {confirmPassword && newPassword !== confirmPassword && (
                <span className="form-hint" style={{ color: "var(--danger)" }}>
                  {t("passwordMismatch")}
                </span>
              )}
            </div>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isSubmitting || (confirmPassword.length > 0 && newPassword !== confirmPassword)}
            >
              {isSubmitting ? t("updating") : t("updatePassword")}
            </button>
          </div>
        </form>
      </div>
    </>
  );
}
