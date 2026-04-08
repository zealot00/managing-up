interface SkeletonProps {
  width?: string | number;
  height?: string | number;
  borderRadius?: string;
  className?: string;
  style?: React.CSSProperties;
}

export function Skeleton({ width, height, borderRadius, className = "", style }: SkeletonProps) {
  return (
    <div
      className={`loading-pulse ${className}`}
      style={{
        width: width ?? "100%",
        height: height ?? 16,
        borderRadius: borderRadius ?? "var(--radius-sm)",
        flexShrink: 0,
        ...style,
      }}
    />
  );
}