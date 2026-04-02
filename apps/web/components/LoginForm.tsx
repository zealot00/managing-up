"use client";

import { useState, FormEvent } from "react";
import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";

export default function LoginForm() {
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
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="login-form">
      <div className="login-form-header">
        <span className="login-form-label">ACCESS</span>
        <h2 className="login-form-title">Sign in</h2>
      </div>

      {error && <p className="login-form-error">{error}</p>}

      <div className="login-form-fields">
        <div className="login-form-field">
          <label className="login-form-field-label">USERNAME</label>
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
            className="login-form-input"
            autoComplete="username"
            placeholder="admin"
          />
        </div>

        <div className="login-form-field">
          <label className="login-form-field-label">PASSWORD</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            className="login-form-input"
            autoComplete="current-password"
            placeholder="••••••••"
          />
        </div>
      </div>

      <button type="submit" disabled={loading} className="login-form-submit">
        {loading ? "AUTHENTICATING..." : "SIGN IN"}
      </button>

      <p className="login-form-hint">
        Default: admin / admin
      </p>
    </form>
  );
}
