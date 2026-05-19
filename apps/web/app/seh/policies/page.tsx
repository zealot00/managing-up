import { getTranslations } from "next-intl/server";
import { getSEHPolicies } from "../../lib/seh-api";
import Breadcrumb from "../../../components/Breadcrumb";
import { PageHeader } from "../../components/layout/PageHeader";

export default async function SEHPoliciesPage() {
  const t = await getTranslations("seh");
  const tc = await getTranslations("common");

  const policies = await getSEHPolicies().catch(() => []);

  return (
    <>
      <Breadcrumb />
      <PageHeader
        eyebrow="SEH"
        title={t("policies")}
        description={t("policiesPageLede")}
      />

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
<th scope="col">Name</th>
                    <th scope="col">Provenance</th>
                    <th scope="col">Min Diversity</th>
                    <th scope="col">Min Golden</th>
                    <th scope="col">Policy ID</th>
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
    </>
  );
}
