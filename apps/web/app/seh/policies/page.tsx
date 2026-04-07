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
        <Link href="/seh" className="toprail-link">← {tc("back")} to SEH</Link>
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
        {policies.length === 0 ? (
          <p className="empty-note">{t("noPolicies")}</p>
        ) : (
          <div className="table-wrapper">
            <table className="table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Provenance</th>
                  <th>Min Diversity</th>
                  <th>Min Golden</th>
                  <th>Policy ID</th>
                </tr>
              </thead>
              <tbody>
                {policies.map((policy) => (
                  <tr key={policy.policy_id}>
                    <td>{policy.name}</td>
                    <td>{policy.require_provenance ? t("provenanceRequired") : "—"}</td>
                    <td>{policy.min_source_diversity}</td>
                    <td>{policy.min_golden_weight}</td>
                    <td>
                      <span className="badge badge-muted">{policy.policy_id}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </main>
  );
}
