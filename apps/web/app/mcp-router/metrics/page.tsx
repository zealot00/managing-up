export default function MCPRouterMetricsPage() {
  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">MCP Router Metrics</h1>
      <p className="text-muted mb-6">
        Prometheus metrics available at <code>/metrics</code> endpoint
      </p>
      
      <div className="card p-6">
        <h3 className="text-lg font-semibold mb-4">Request Metrics</h3>
        <div className="space-y-4">
          <MetricsCard 
            title="Total Requests" 
            description="Total number of MCP router requests"
            metric="mcp_router_requests_total"
          />
          <MetricsCard 
            title="Request Duration" 
            description="Histogram of request latency"
            metric="mcp_router_request_duration_seconds"
          />
          <MetricsCard 
            title="Match Failures" 
            description="Counter of route match failures"
            metric="mcp_router_match_failures_total"
          />
        </div>
      </div>
      
      <div className="card p-6 mt-6">
        <h3 className="text-lg font-semibold mb-4">Integration</h3>
        <p className="text-muted">
          Metrics are exposed in Prometheus format. Configure your Prometheus server to scrape:
        </p>
        <code className="block bg-muted p-2 rounded mt-2">
          {`http://localhost:8080/metrics`}
        </code>
      </div>
    </div>
  );
}

function MetricsCard({ title, description, metric }: { title: string; description: string; metric: string }) {
  return (
    <div className="border border-border rounded p-4">
      <h4 className="font-medium">{title}</h4>
      <p className="text-sm text-muted">{description}</p>
      <code className="block bg-muted p-2 rounded mt-2 text-xs">
        {metric}
      </code>
    </div>
  );
}
