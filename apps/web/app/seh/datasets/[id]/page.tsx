import Link from "next/link";
import { notFound } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { getSEHDataset, getSEHDatasetCases, getSEHDatasetLineage, getSEHCaseLineage } from "../../../lib/seh-api";
import JsonFold from "../../../components/JsonFold";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function SEHDatasetDetailPage({ params }: Props) {
  const t = await getTranslations("seh");
  const tc = await getTranslations("common");
  const { id } = await params;

  const dataset = await getSEHDataset(id).catch(() => null);
  if (!dataset) {
    notFound();
  }

  const [casesResp, lineage] = await Promise.all([
    getSEHDatasetCases(id, 50, 0).catch(() => ({
      manifest: {},
      cases: [],
      pagination: { limit: 50, offset: 0, total: 0, has_more: false },
    })),
    getSEHDatasetLineage(id).catch(() => ({ versions: [] })),
  ]);

  // Fetch case lineage for each case
  const caseLineages = await Promise.all(
    casesResp.cases.slice(0, 10).map(async (c) => {
      const ln = await getSEHCaseLineage(c.case_id).catch(() => null);
      return { case_id: c.case_id, lineage: ln };
    })
  );

  return (
    <main className="shell">
      <section className="toprail">
        <Link href="/seh/datasets">{tc("back")} to {t("datasets")}</Link>
      </section>

      <section className="hero-page hero-compact">
        <p className="eyebrow">SEH</p>
        <h1>{dataset.name}</h1>
        <p className="lede">
          {dataset.version} · {dataset.owner} · {dataset.case_count} {t("cases")}
        </p>
      </section>

      <section className="panel-grid panel-grid-wide">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">SEH</p>
            <h2 className="panel-title">{t("datasetInfo")}</h2>
          </div>
          <div className="detail-grid">
            <div className="detail-row">
              <span className="detail-label">{tc("id")}</span>
              <span className="detail-value">{dataset.dataset_id}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{tc("name")}</span>
              <span className="detail-value">{dataset.name}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{tc("version")}</span>
              <span className="detail-value">{dataset.version}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{tc("owner")}</span>
              <span className="detail-value">{dataset.owner}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{tc("createdAt")}</span>
              <span className="detail-value">{new Date(dataset.created_at).toLocaleString()}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Checksum</span>
              <span className="detail-value" style={{ fontFamily: "monospace", fontSize: "var(--text-xs)" }}>{dataset.checksum || "-"}</span>
            </div>
          </div>
          {dataset.description && (
            <p className="list-card-meta" style={{ marginTop: "var(--space-4)" }}>
              {dataset.description}
            </p>
          )}
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">SEH</p>
            <h2 className="panel-title">{t("datasetLineage")}</h2>
          </div>
          {lineage.versions.length === 0 ? (
            <p className="empty-note">{t("noLineage")}</p>
          ) : (
            <JsonFold title={t("datasetLineage")} data={lineage.versions} defaultCollapsed={false} />
          )}
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">SEH</p>
            <h2 className="panel-title">{t("datasetCases", { count: casesResp.cases.length })}</h2>
          </div>
          {casesResp.cases.length === 0 ? (
            <p className="empty-note">{t("noCases")}</p>
          ) : (
            <div className="list">
              {casesResp.cases.map((testCase) => {
                const caseLn = caseLineages.find(cl => cl.case_id === testCase.case_id);
                return (
                  <article className="list-card" key={testCase.case_id} style={{ display: "block" }}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">
                        {testCase.skill} · {testCase.source}
                      </h3>
                      <p className="list-card-meta">{testCase.tags.join(", ") || "-"}</p>
                    </div>
                    <div className="list-card-actions">
                      <span className={`badge badge-${testCase.status}`}>{testCase.status}</span>
                    </div>

                    {/* Case provenance fold */}
                    <div style={{ marginTop: "var(--space-2)" }}>
                      <JsonFold title={`${t("caseLineage")}: ${testCase.case_id.slice(0, 12)}...`} data={{
                        provenance: testCase.provenance,
                        input: testCase.input,
                        expected: testCase.expected,
                        lineage: caseLn?.lineage,
                      }} />
                    </div>
                  </article>
                );
              })}
            </div>
          )}
        </article>
      </section>
    </main>
  );
}
