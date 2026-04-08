import { Suspense } from "react";
import { MySkillsClient } from "./MySkillsClient";
import { Skeleton } from "../../components/ui/Skeleton";

export default function MySkillsPage() {
  return (
    <Suspense fallback={<Skeleton className="h-[400px]" />}>
      <MySkillsClient />
    </Suspense>
  );
}