import { getSEHDatasets } from "../../lib/seh-api";
import SEHDatasetsList from "../../components/SEHDatasetsList";

export default async function SEHDatasetsPage() {
  let error = false;
  const datasetsResp = await getSEHDatasets(100, 0).catch(() => {
    error = true;
    return {
      datasets: [],
      pagination: { limit: 100, offset: 0, total: 0, has_more: false },
    };
  });

  return (
    <SEHDatasetsList
      datasets={datasetsResp.datasets}
      total={datasetsResp.pagination.total}
      hasMore={datasetsResp.pagination.has_more}
      error={error}
    />
  );
}
