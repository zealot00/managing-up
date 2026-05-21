"use client";

import { useState, FormEvent } from "react";
import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { Spinner } from "../app/components/ui/Spinner";
import { PasswordInput } from "../app/components/ui/PasswordInput";

export default function LoginForm() {
  const t = useTranslations("login");
  const { login } = useAuth();
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await login(username, password);
      router.push("/dashboard");
    } catch (err) {
      setError(t("invalidCredentials"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="login-form">
      <div className="login-form-header">
        <span className="login-form-label">ACCESS</span>
        <h2 className="login-form-title">{t("title")}</h2>
      </div>

      {error && <p className="login-form-error" role="alert">{error}</p>}

      <div className="login-form-fields">
        <div className="login-form-field">
          <label className="login-form-field-label" htmlFor="login-username">{t("username")}</label>
          <input
            id="login-username"
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
            className="login-form-input"
            autoComplete="username"
            placeholder={t("usernamePlaceholder")}
          />
        </div>

        <div className="login-form-field">
          <label className="login-form-field-label" htmlFor="login-password">{t("password")}</label>
          <PasswordInput
            id="login-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            inputClassName="login-form-input"
            autoComplete="current-password"
            placeholder={t("passwordPlaceholder")}
          />
        </div>
      </div>

      <button type="submit" className="login-form-submit" style={{ display: "flex", alignItems: "center", justifyContent: "center", gap: 8 }}>
        {loading ? (
          <>
            <Spinner size="sm" />
            {t("signingIn")}
          </>
        ) : (
          t("signIn")
        )}
      </button>

      <p className="login-form-hint">
        Default: admin / admin
      </p>
    </form>
  );
}
