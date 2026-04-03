"use client";

import { useEffect } from "react";
import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";

export default function HomePage() {
  const t = useTranslations("common");
  const tl = useTranslations("landing");
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace("/dashboard");
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div className="landing-page">
        <section className="landing-hero">
          <div className="loading-pulse loading-pulse-short" style={{ width: 200, marginBottom: 16 }} />
          <div className="loading-pulse loading-pulse-medium" style={{ width: 400, marginBottom: 16 }} />
          <div className="loading-pulse loading-pulse-long" style={{ width: 600 }} />
        </section>
      </div>
    );
  }

  if (isAuthenticated) {
    return null;
  }

  return (
    <div className="landing-page">
      <section className="landing-hero">
        <p className="landing-eyebrow">{tl("eyebrow")}</p>
        <h1 className="landing-title">{tl("title")}</h1>
        <p className="landing-subtitle">
          {tl("subtitle")}
        </p>
        <div className="landing-actions">
          <a href="/login" className="landing-btn landing-btn-primary">{t("getStarted")}</a>
          <a href="https://github.com/your-org/managing-up" className="landing-btn landing-btn-secondary" target="_blank" rel="noopener noreferrer">{t("viewOnGitHub")}</a>
        </div>
      </section>

      <section className="landing-features">
        <div className="landing-features-grid">
          <article className="landing-feature">
            <div className="landing-feature-icon">◈</div>
            <h3 className="landing-feature-title">{tl("features.skillRegistry.title")}</h3>
            <p className="landing-feature-desc">{tl("features.skillRegistry.desc")}</p>
          </article>
          <article className="landing-feature">
            <div className="landing-feature-icon">▸</div>
            <h3 className="landing-feature-title">{tl("features.executionEngine.title")}</h3>
            <p className="landing-feature-desc">{tl("features.executionEngine.desc")}</p>
          </article>
          <article className="landing-feature">
            <div className="landing-feature-icon">◎</div>
            <h3 className="landing-feature-title">{tl("features.evaluationPipeline.title")}</h3>
            <p className="landing-feature-desc">{tl("features.evaluationPipeline.desc")}</p>
          </article>
          <article className="landing-feature">
            <div className="landing-feature-icon">↻</div>
            <h3 className="landing-feature-title">{tl("features.traceReplay.title")}</h3>
            <p className="landing-feature-desc">{tl("features.traceReplay.desc")}</p>
          </article>
        </div>
      </section>
    </div>
  );
}
