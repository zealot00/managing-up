import { getTranslations } from "next-intl/server";
import TaskFromTraceForm from "../../../components/TaskFromTraceForm";

export default async function TaskBuilderPage() {
  const t = await getTranslations("tasks");
  const tc = await getTranslations("common");

  return (
    <main className="shell">
      <section className="toprail" aria-label="Tasks navigation">
        <a className="toprail-link" href="/">
          {tc("dashboard")}
        </a>
        <a className="toprail-link" href="/tasks">
          {tc("tasks")}
        </a>
        <a className="toprail-link" href="/tasks/from-trace">
          {t("taskBuilder")}
        </a>
      </section>

      <section className="hero-page hero-compact">
        <p className="eyebrow">{t("taskBuilder")}</p>
        <h1>{t("taskBuilder.title")}</h1>
        <p className="lede">
          {t("taskBuilder.lede")}
        </p>
      </section>

      <div className="content-grid">
        <article className="panel">
          <div className="panel-header">
            <h2>{t("taskBuilder.taskConfig")}</h2>
          </div>

          <TaskFromTraceForm />
        </article>
      </div>
    </main>
  );
}