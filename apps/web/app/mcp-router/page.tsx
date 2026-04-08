import { Suspense } from "react";
import { MCPRouterDashboardClient } from "./MCPRouterDashboardClient";
import { Skeleton } from "../components/layout/Skeleton";

export default function MCPRouterPage() {
  return (
    <Suspense fallback={<Skeleton className="h-[400px]" />}>
      <MCPRouterDashboardClient />
    </Suspense>
  );
}
