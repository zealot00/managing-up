type BarChartData = {
  label: string;
  value: number;
  color?: string;
};

type Props = {
  data: BarChartData[];
  title?: string;
  valuePrefix?: string;
  valueSuffix?: string;
  maxBars?: number;
};

export default function BarChart({ data, title, valuePrefix = "", valueSuffix = "", maxBars = 8 }: Props) {
  const displayData = data.slice(0, maxBars);
  const maxValue = Math.max(...displayData.map((d) => d.value), 1);

  return (
    <div className="chart-container">
      {title && <h3 className="chart-title">{title}</h3>}
      <div className="chart-bars">
        {displayData.map((item, index) => {
          const percentage = (item.value / maxValue) * 100;
          return (
            <div key={item.label} className="chart-bar-row">
              <div className="chart-bar-label" title={item.label}>
                {item.label}
              </div>
              <div className="chart-bar-track">
                <div
                  className="chart-bar-fill"
                  style={{
                    width: `${percentage}%`,
                    background: item.color || `var(--ink-strong)`,
                    opacity: 0.7 + (index * 0.03),
                  }}
                />
              </div>
              <div className="chart-bar-value">
                {valuePrefix}{typeof item.value === "number" ? item.value.toLocaleString() : item.value}{valueSuffix}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
