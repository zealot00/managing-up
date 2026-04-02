type LineChartDataPoint = {
  label: string;
  value: number;
};

type Props = {
  data: LineChartDataPoint[];
  title?: string;
  valuePrefix?: string;
  valueSuffix?: string;
  height?: number;
};

export default function LineChart({ data, title, valuePrefix = "", valueSuffix = "", height = 200 }: Props) {
  if (data.length === 0) {
    return (
      <div className="chart-container">
        {title && <h3 className="chart-title">{title}</h3>}
        <p className="chart-empty">No data available</p>
      </div>
    );
  }

  const maxValue = Math.max(...data.map((d) => d.value), 1);
  const minValue = Math.min(...data.map((d) => d.value), 0);
  const range = maxValue - minValue || 1;

  const width = 100;
  const paddingX = 0;
  const paddingY = 20;
  const chartWidth = width - paddingX * 2;
  const chartHeight = height - paddingY * 2;

  const points = data.map((d, i) => {
    const x = paddingX + (i / (data.length - 1 || 1)) * chartWidth;
    const y = paddingY + chartHeight - ((d.value - minValue) / range) * chartHeight;
    return { x, y, ...d };
  });

  const pathD = points.map((p, i) => `${i === 0 ? "M" : "L"} ${p.x} ${p.y}`).join(" ");
  const areaD = `${pathD} L ${points[points.length - 1].x} ${paddingY + chartHeight} L ${points[0].x} ${paddingY + chartHeight} Z`;

  const gridLines = 4;
  const gridValues = Array.from({ length: gridLines + 1 }, (_, i) => {
    return minValue + (range * i) / gridLines;
  });

  return (
    <div className="chart-container">
      {title && <h3 className="chart-title">{title}</h3>}
      <div className="chart-line-wrapper">
        <svg
          viewBox={`0 0 ${width} ${height}`}
          className="chart-line-svg"
          preserveAspectRatio="none"
        >
          {gridValues.map((val, i) => {
            const y = paddingY + chartHeight - (i / gridLines) * chartHeight;
            return (
              <g key={i}>
                <line
                  x1={paddingX}
                  y1={y}
                  x2={paddingX + chartWidth}
                  y2={y}
                  stroke="var(--line)"
                  strokeWidth={0.3}
                />
                <text
                  x={paddingX - 1}
                  y={y + 1}
                  fontSize={2.5}
                  fill="var(--muted)"
                  textAnchor="end"
                  dominantBaseline="middle"
                >
                  {valuePrefix}{Math.round(val).toLocaleString()}{valueSuffix}
                </text>
              </g>
            );
          })}

          <path d={areaD} fill="var(--ink-strong)" opacity={0.08} />
          <path d={pathD} fill="none" stroke="var(--ink-strong)" strokeWidth={0.8} />

          {points.map((p, i) => (
            <circle
              key={i}
              cx={p.x}
              cy={p.y}
              r={1.2}
              fill="var(--surface-raised)"
              stroke="var(--ink-strong)"
              strokeWidth={0.5}
            />
          ))}
        </svg>

        <div className="chart-line-labels">
          {data.map((d, i) => (
            <div key={i} className="chart-line-label">
              {d.label}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
