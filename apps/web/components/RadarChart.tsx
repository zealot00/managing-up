"use client";

import { useState } from "react";

export type RadarDataPoint = {
  name: string;
  score: number;
  experimentId: string;
  experimentName: string;
  timestamp: string;
};

export type RadarChartData = {
  capabilities: string[];
  experiments: Array<{
    id: string;
    name: string;
    color: string;
    scores: number[];
  }>;
};

type Props = {
  data: RadarChartData;
  width?: number;
  height?: number;
};

const GRID_LEVELS = [0.25, 0.5, 0.75, 1.0];
const DEFAULT_SIZE = 480;
const PADDING = 60;
const LABEL_OFFSET = 30;

const EXPERIMENT_COLORS = [
  "var(--primary)",
  "var(--success)",
  "var(--warning)",
  "#8b5cf6",
  "#ec4899",
  "#06b6d4",
];

function polarToCartesian(
  centerX: number,
  centerY: number,
  radius: number,
  angleInDegrees: number
): { x: number; y: number } {
  const angleInRadians = ((angleInDegrees - 90) * Math.PI) / 180.0;
  return {
    x: centerX + radius * Math.cos(angleInRadians),
    y: centerY + radius * Math.sin(angleInRadians),
  };
}

function buildPolygonPath(
  centerX: number,
  centerY: number,
  maxRadius: number,
  scores: number[]
): string {
  if (scores.length === 0) return "";

  const angleStep = 360 / scores.length;
  const points = scores.map((score, i) => {
    const radius = (score / 100) * maxRadius;
    const angle = i * angleStep;
    return polarToCartesian(centerX, centerY, radius, angle);
  });

  return points.map((p, i) => `${i === 0 ? "M" : "L"} ${p.x} ${p.y}`).join(" ") + " Z";
}

function buildGridPath(
  centerX: number,
  centerY: number,
  maxRadius: number,
  levels: number[]
): string {
  if (levels.length === 0) return "";
  const angleStep = 360 / Math.max(4, levels.length * 2);
  const numAxes = Math.max(6, levels.length * 2);

  return levels
    .map((level) => {
      const radius = level * maxRadius;
      const points: string[] = [];
      for (let i = 0; i < numAxes; i++) {
        const angle = i * angleStep;
        const p = polarToCartesian(centerX, centerY, radius, angle);
        points.push(`${i === 0 ? "M" : "L"} ${p.x} ${p.y}`);
      }
      return points.join(" ") + " Z";
    })
    .join(" ");
}

export default function RadarChart({ data, width = DEFAULT_SIZE, height = DEFAULT_SIZE }: Props) {
  const [hoveredPoint, setHoveredPoint] = useState<{
    x: number;
    y: number;
    capability: string;
    experiment: string;
    score: number;
  } | null>(null);

  const centerX = width / 2;
  const centerY = height / 2;
  const maxRadius = Math.min(centerX, centerY) - PADDING - LABEL_OFFSET;
  const numAxes = data.capabilities.length;
  const angleStep = numAxes > 0 ? 360 / numAxes : 0;

  const axisLines = Array.from({ length: numAxes }, (_, i) => {
    const angle = i * angleStep;
    const end = polarToCartesian(centerX, centerY, maxRadius, angle);
    return { x1: centerX, y1: centerY, x2: end.x, y2: end.y };
  });

  const labels = data.capabilities.map((cap, i) => {
    const angle = i * angleStep;
    const labelPos = polarToCartesian(centerX, centerY, maxRadius + LABEL_OFFSET, angle);
    let textAnchor = "middle";
    if (angle > 45 && angle < 135) textAnchor = "start";
    else if (angle > 225 && angle < 315) textAnchor = "end";
    return {
      x: labelPos.x,
      y: labelPos.y,
      text: cap,
      textAnchor,
    };
  });

  const gridPath = buildGridPath(centerX, centerY, maxRadius, GRID_LEVELS);

  const experimentPolygons = data.experiments.map((exp, expIndex) => ({
    ...exp,
    path: buildPolygonPath(centerX, centerY, maxRadius, exp.scores),
    color: EXPERIMENT_COLORS[expIndex % EXPERIMENT_COLORS.length],
  }));

  const handleMouseMove = (
    e: React.MouseEvent<SVGElement>,
    expIndex: number,
    capIndex: number
  ) => {
    const svgRect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - svgRect.left;
    const y = e.clientY - svgRect.top;
    const exp = data.experiments[expIndex];
    const capability = data.capabilities[capIndex];
    if (exp && capability !== undefined) {
      setHoveredPoint({
        x,
        y,
        capability,
        experiment: exp.name,
        score: exp.scores[capIndex],
      });
    }
  };

  return (
    <div style={{ position: "relative", width, height }}>
      <svg width={width} height={height} viewBox={`0 0 ${width} ${height}`}>
        <defs>
          {experimentPolygons.map((exp, i) => (
            <linearGradient
              key={`gradient-${i}`}
              id={`gradient-${i}`}
              x1="0%"
              y1="0%"
              x2="100%"
              y2="100%"
            >
              <stop offset="0%" stopColor={exp.color} stopOpacity={0.4} />
              <stop offset="100%" stopColor={exp.color} stopOpacity={0.15} />
            </linearGradient>
          ))}
        </defs>

        <g opacity={0.4}>
          <path d={gridPath} fill="none" stroke="var(--line-strong)" strokeWidth={1} />
        </g>

        <g>
          {axisLines.map((line, i) => (
            <line
              key={i}
              x1={line.x1}
              y1={line.y1}
              x2={line.x2}
              y2={line.y2}
              stroke="var(--line-strong)"
              strokeWidth={1}
            />
          ))}
        </g>

        {numAxes > 0 &&
          GRID_LEVELS.map((level) => {
            const angle = -90;
            const pos = polarToCartesian(centerX, centerY, level * maxRadius, angle);
            return (
              <text
                key={level}
                x={pos.x - 8}
                y={pos.y}
                fontSize={10}
                fill="var(--muted)"
                textAnchor="end"
                dominantBaseline="middle"
              >
                {Math.round(level * 100)}
              </text>
            );
          })}

        <g>
          {experimentPolygons.map((exp, expIndex) => (
            <path
              key={exp.id}
              d={exp.path}
              fill={`url(#gradient-${expIndex})`}
              stroke={exp.color}
              strokeWidth={2}
              strokeOpacity={0.8}
              style={{ cursor: "crosshair" }}
              onMouseMove={(e) => {
                const capIndex = Math.floor(
                  ((Math.atan2(e.nativeEvent.offsetY - centerY, e.nativeEvent.offsetX - centerX) *
                    180) /
                    Math.PI +
                    90 +
                    360) %
                    360 /
                    angleStep
                );
                handleMouseMove(e, expIndex, capIndex % numAxes);
              }}
              onMouseLeave={() => setHoveredPoint(null)}
            />
          ))}
        </g>

        <g>
          {experimentPolygons.map((exp, expIndex) =>
            exp.scores.map((score, capIndex) => {
              const angle = capIndex * angleStep;
              const radius = (score / 100) * maxRadius;
              const pos = polarToCartesian(centerX, centerY, radius, angle);
              return (
                <circle
                  key={`${exp.id}-${capIndex}`}
                  cx={pos.x}
                  cy={pos.y}
                  r={4}
                  fill={exp.color}
                  stroke="var(--surface-raised)"
                  strokeWidth={2}
                  style={{ cursor: "pointer" }}
                  onMouseMove={(e) => handleMouseMove(e, expIndex, capIndex)}
                  onMouseLeave={() => setHoveredPoint(null)}
                />
              );
            })
          )}
        </g>

        <g>
          {labels.map((label, i) => (
            <text
              key={i}
              x={label.x}
              y={label.y}
              fontSize={12}
              fontWeight={600}
              fill="var(--ink)"
              textAnchor={label.textAnchor}
              dominantBaseline="middle"
            >
              {label.text}
            </text>
          ))}
        </g>
      </svg>

      {hoveredPoint && (
        <div
          style={{
            position: "absolute",
            left: hoveredPoint.x + 12,
            top: hoveredPoint.y - 12,
            background: "var(--surface-raised)",
            border: "1px solid var(--line-strong)",
            borderRadius: "var(--radius-sm)",
            padding: "8px 12px",
            pointerEvents: "none",
            boxShadow: "var(--shadow)",
            zIndex: 10,
          }}
        >
          <p style={{ margin: 0, fontWeight: 600, fontSize: "0.85rem", color: "var(--ink-strong)" }}>
            {hoveredPoint.capability}
          </p>
          <p style={{ margin: "4px 0 0", fontSize: "0.8rem", color: "var(--muted)" }}>
            {hoveredPoint.experiment}: {hoveredPoint.score.toFixed(1)}
          </p>
        </div>
      )}
    </div>
  );
}
