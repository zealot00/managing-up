import { getSEHRuns } from "../../lib/seh-api";
import SEHRunsList from "../../components/SEHRunsList";

export default async function SEHRunsPage() {
  const runsResp = await getSEHRuns(100, 0).catch(() => ({
    runs: [],
    pagination: { limit: 100, offset: 0, total: 0, has_more: false },
  }));

  return (
    <SEHRunsList
      runs={runsResp.runs}
      total={runsResp.pagination.total}
      hasMore={runsResp.pagination.has_more}
    />
  );
}
