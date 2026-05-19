import { getTranslations } from "next-intl/server";
import TaskFromTraceForm from "../../../components/TaskFromTraceForm";
import Breadcrumb from "../../../components/Breadcrumb";
import { PageHeader } from "../../components/layout/PageHeader";

export default async function TaskBuilderPage() {
  const t = await getTranslations("tasks");

  return (
    <>
      <Breadcrumb />
      <PageHeader
        eyebrow={t("taskBuilder.eyebrow")}
        title={t("taskBuilder.title")}
        description={t("taskBuilder.lede")}
      />

      <div style={{ maxWidth: 640 }}>
        <TaskFromTraceForm />
      </div>
    </>
  );
}
