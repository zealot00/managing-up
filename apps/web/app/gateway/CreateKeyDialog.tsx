"use client";

import { FormEvent, useState, useEffect, useCallback, useRef } from "react";
import { useTranslations } from "next-intl";
import { KeyRound, X, Copy, Check } from "lucide-react";
import { createGatewayKey, GatewayKeyMeta } from "../lib/gateway-api";
import { useFocusTrap } from "../hooks/use-focus-trap";

type DialogState = "open" | "submitting" | "success";

interface CreateKeyDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated: (key: GatewayKeyMeta) => void;
  existingKeys: GatewayKeyMeta[];
}

export function CreateKeyDialog({ isOpen, onClose, onCreated, existingKeys }: CreateKeyDialogProps) {
  const t = useTranslations("gateway");
  const tc = useTranslations("common");
  const [state, setState] = useState<DialogState>("open");
  const [keyName, setKeyName] = useState("");
  const [newKeyValue, setNewKeyValue] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const titleId = useRef(`create-key-title-${Math.random().toString(36).slice(2, 9)}`).current;
  const setContainerRef = useFocusTrap(isOpen);

  const activeKeyNames = existingKeys.filter((k) => !k.revoked_at).map((k) => k.name);

  const reset = useCallback(() => {
    setKeyName("");
    setNewKeyValue(null);
    setError(null);
    setState("open");
    setCopied(false);
  }, []);

  useEffect(() => {
    if (isOpen) reset();
  }, [isOpen, reset]);

  useEffect(() => {
    function handleEsc(e: KeyboardEvent) {
      if (e.key === "Escape" && state !== "submitting") handleClose();
    }
    if (isOpen) {
      document.addEventListener("keydown", handleEsc);
      document.body.style.overflow = "hidden";
    }
    return () => {
      document.removeEventListener("keydown", handleEsc);
      document.body.style.overflow = "";
    };
  }, [isOpen, state]);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const trimmed = keyName.trim();
    if (!trimmed) return;

    if (activeKeyNames.includes(trimmed)) {
      setError(t("duplicateName"));
      return;
    }

    setError(null);
    setState("submitting");
    try {
      const resp = await createGatewayKey(trimmed);
      setNewKeyValue(resp.key);
      setState("success");
      onCreated(resp.key_meta);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create key");
      setState("open");
    }
  }

  async function handleCopy() {
    if (!newKeyValue) return;
    try {
      await navigator.clipboard.writeText(newKeyValue);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // fallback: select text
    }
  }

  function handleClose() {
    if (state === "submitting") return;
    onClose();
  }

  if (!isOpen) return null;

  const isSuccess = state === "success";

  return (
    <div
      role="presentation"
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0, 0, 0, 0.5)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1000,
        padding: "var(--space-6)",
        animation: "fadeIn 0.15s ease-out",
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget && state !== "submitting") handleClose();
      }}
    >
      <div
        ref={setContainerRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        style={{
          background: "var(--surface-raised)",
          borderRadius: "var(--radius-lg)",
          padding: "var(--space-6)",
          width: "100%",
          maxWidth: 480,
          boxShadow: "var(--shadow-lg)",
          animation: "scaleIn 0.15s ease-out",
        }}
      >
        <div style={{ display: "flex", gap: "var(--space-4)", alignItems: "flex-start" }}>
          <div
            style={{
              width: 40,
              height: 40,
              borderRadius: "var(--radius-md)",
              background: isSuccess ? "var(--success-bg, rgba(34,197,94,0.1))" : "var(--primary-bg, rgba(59,130,246,0.1))",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              flexShrink: 0,
            }}
          >
            <KeyRound size={20} style={{ color: isSuccess ? "var(--success)" : "var(--primary)" }} />
          </div>

          <div style={{ flex: 1, minWidth: 0 }}>
            <h3
              id={titleId}
              style={{
                fontSize: "var(--text-lg)",
                fontWeight: 700,
                color: "var(--ink-strong)",
                marginBottom: "var(--space-2)",
              }}
            >
              {isSuccess ? t("keyCreated") : t("createKeyTitle")}
            </h3>

            {isSuccess ? (
              <>
                <p style={{ fontSize: "var(--text-sm)", color: "var(--warning)", lineHeight: 1.6, marginBottom: "var(--space-3)" }}>
                  {t("secretWarning")}
                </p>
                <div className="gateway-secret-code" style={{ position: "relative" }}>
                  <code>{newKeyValue}</code>
                  <button
                    onClick={handleCopy}
                    aria-label={copied ? t("copied") : t("copyKey")}
                    style={{
                      position: "absolute",
                      top: "var(--space-2)",
                      right: "var(--space-2)",
                      background: "var(--surface-raised)",
                      border: "1px solid var(--line)",
                      borderRadius: "var(--radius-sm)",
                      padding: "var(--space-1) var(--space-2)",
                      cursor: "pointer",
                      display: "flex",
                      alignItems: "center",
                      gap: "var(--space-1)",
                      fontSize: "var(--text-xs)",
                      color: "var(--ink)",
                    }}
                  >
                    {copied ? <Check size={14} /> : <Copy size={14} />}
                    {copied ? t("copied") : t("copyKey")}
                  </button>
                </div>
              </>
            ) : (
              <form onSubmit={handleSubmit}>
                <label className="form-label" htmlFor="create-key-name" style={{ display: "block", marginBottom: "var(--space-2)" }}>
                  {t("keyName")}
                </label>
                <input
                  id="create-key-name"
                  className="form-input"
                  value={keyName}
                  onChange={(e) => {
                    setKeyName(e.target.value);
                    setError(null);
                  }}
                  placeholder={t("keyNamePlaceholder")}
                  disabled={state === "submitting"}
                  autoFocus
                  style={{ width: "100%" }}
                />
                {error && <p className="form-error" role="alert" style={{ marginTop: "var(--space-2)" }}>{error}</p>}
              </form>
            )}
          </div>

          <button
            onClick={handleClose}
            disabled={state === "submitting"}
            aria-label={tc("cancel")}
            style={{
              background: "none",
              border: "none",
              padding: "var(--space-1)",
              cursor: state === "submitting" ? "not-allowed" : "pointer",
              color: "var(--muted)",
              borderRadius: "var(--radius-sm)",
              transition: "all var(--transition-fast)",
            }}
          >
            <X size={18} />
          </button>
        </div>

        <div
          style={{
            display: "flex",
            gap: "var(--space-3)",
            marginTop: "var(--space-6)",
            justifyContent: "flex-end",
          }}
        >
          {isSuccess ? (
            <button
              onClick={handleClose}
              className="gateway-button-create"
            >
              {t("keySaved")}
            </button>
          ) : (
            <>
              <button
                onClick={handleClose}
                disabled={state === "submitting"}
                className="gateway-button-cancel"
              >
                {tc("cancel")}
              </button>
              <button
                onClick={() => {
                  const form = document.getElementById("create-key-name")?.closest("form");
                  if (form) form.requestSubmit();
                }}
                disabled={state === "submitting" || !keyName.trim()}
                className="gateway-button-create"
                style={{ opacity: !keyName.trim() || state === "submitting" ? 0.5 : 1 }}
              >
                {state === "submitting" ? t("creating") : t("newKey")}
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
