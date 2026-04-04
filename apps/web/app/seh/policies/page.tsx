import Link from "next/link";
import { getTranslations } from "next-intl/server";
import { getSEHPolicies } from "../../lib/seh-api";

export default async function SEHPoliciesPage() {
  const t = await getTranslations("seh");
  const tc = await getTranslations("common");

  const policies = await getSEHPolicies().catch(() => []);

  return (
    <main className="shell">
      <section className="toprail">
        <Link href="/seh">{tc("back")} to SEH</Link>
      </section>

      <header className="hero-page hero-compact">
        <p className="eyebrow">SEH</p>
        <h1>{t("policies")}</h1>
        <p className="lede">{t("policiesPageLede")}</p>
      </header>

      <section className="panel">
        <div className="panel-header">
          <p className="section-kicker">SEH</p>
          <h2 className="panel-title">{t("activePolicies", { count: policies.length })}</h2>
        </div>
        <div className="list">
          {policies.length === 0 ? (
            <p className="empty-note">{t("noPolicies")}</p>
          ) : (
            policies.map((policy) => (
              <article className="list-card" key={policy.policy_id}>
                <div className="list-card-main">
                  <h3 className="list-card-title">{policy.name}</h3>
                  <p className="list-card-meta">
                    {policy.require_provenance ? `${t("provenanceRequired")} · ` : ""}
                    {t("minDiversity")}: {policy.min_source_diversity} · {t("minGolden")}: {policy.min_golden_weight}
                  </p>
                </div>
                <span className="badge badge-muted">{policy.policy_id}</span>
              </article>
            ))
          )}
        </div>
      </section>
    </main>
  );
}
