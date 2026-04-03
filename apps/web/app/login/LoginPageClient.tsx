"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import LoginForm from "../../components/LoginForm";

const defaultTip = {
  content: "Talk is cheap. Show me the code.",
  author: "Linus Torvalds",
};

export default function LoginPageClient() {
  const t = useTranslations("login");
  const [tip, setTip] = useState(defaultTip);

  useEffect(() => {
    const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";
    fetch(`${baseUrl}/api/v1/tip`)
      .then((res) => res.json())
      .then((data) => {
        if (data.data && data.data.content) {
          setTip({
            content: data.data.content,
            author: data.data.author || "",
          });
        }
      })
      .catch(() => {});
  }, []);

  return (
    <div className="login-page">
      <div className="login-bg">
        <svg className="login-bg-svg" viewBox="0 0 100 100" preserveAspectRatio="none">
          <line x1="0" y1="0" x2="100" y2="100" stroke="var(--line)" strokeWidth="0.3" />
          <line x1="100" y1="0" x2="0" y2="100" stroke="var(--line)" strokeWidth="0.3" />
          <line x1="50" y1="0" x2="50" y2="100" stroke="var(--line)" strokeWidth="0.2" />
          <line x1="0" y1="50" x2="100" y2="50" stroke="var(--line)" strokeWidth="0.2" />
          <circle cx="50" cy="50" r="30" fill="none" stroke="var(--line-strong)" strokeWidth="0.3" />
          <circle cx="50" cy="50" r="20" fill="none" stroke="var(--line)" strokeWidth="0.2" />
          <circle cx="50" cy="50" r="10" fill="none" stroke="var(--line-strong)" strokeWidth="0.3" />
          <line x1="20" y1="20" x2="80" y2="80" stroke="var(--line)" strokeWidth="0.15" />
          <line x1="80" y1="20" x2="20" y2="80" stroke="var(--line)" strokeWidth="0.15" />
          <line x1="35" y1="0" x2="35" y2="100" stroke="var(--line)" strokeWidth="0.1" />
          <line x1="65" y1="0" x2="65" y2="100" stroke="var(--line)" strokeWidth="0.1" />
          <line x1="0" y1="35" x2="100" y2="35" stroke="var(--line)" strokeWidth="0.1" />
          <line x1="0" y1="65" x2="100" y2="65" stroke="var(--line)" strokeWidth="0.1" />
        </svg>
      </div>

      <div className="login-container">
        <div className="login-left">
          <div className="login-brand">
            <img src="/logo.svg" alt="managing-up" className="login-logo" />
            <div className="login-brand-text">
              <span className="login-brand-name">MANAGING UP</span>
              <span className="login-brand-sub">{t("subtitle").split(" ")[0]}</span>
            </div>
          </div>

          <div className="login-quote">
            <span className="login-quote-mark">"</span>
            <p className="login-quote-text">{tip.content}</p>
            {tip.author && (
              <p className="login-quote-author">— {tip.author}</p>
            )}
          </div>

          <div className="login-footer">
            <span className="login-footer-text">ENTERPRISE AI QUALITY INFRASTRUCTURE</span>
          </div>
        </div>

        <div className="login-right">
          <LoginForm />
        </div>
      </div>
    </div>
  );
}