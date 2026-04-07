import { getTranslations } from "next-intl/server";
import TaskFromTraceForm from "../../../components/TaskFromTraceForm";

export default async function TaskBuilderPage() {
  const t = await getTranslations("tasks");

  return (
    <main className="shell">
      <header className="hero-page hero-compact" style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: "var(--space-6)" }}>
        <div style={{ flex: 1 }}>
          <p className="eyebrow">{t("taskBuilder")}</p>
          <h1>{t("taskBuilder.title")}</h1>
          <p className="lede">{t("taskBuilder.lede")}</p>
        </div>
      </header>

      <div style={{ maxWidth: 640 }}>
        <TaskFromTraceForm />
      </div>
    </main>
  );
}
