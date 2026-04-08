import { Suspense } from "react";
import { SkillMarketClient } from "./SkillMarketClient";
import { Skeleton } from "../../components/ui/Skeleton";

export default function SkillMarketPage() {
  return (
    <Suspense fallback={<Skeleton className="h-[400px]" />}>
      <SkillMarketClient />
    </Suspense>
  );
}
