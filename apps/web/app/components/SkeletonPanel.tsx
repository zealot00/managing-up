"use client";
import { SkeletonCard } from "./SkeletonCard";

export function SkeletonPanel({ height = 320 }: { height?: number }) {
  return (
    <article className="panel" style={{ minHeight: height }}>
      <div className="loading-pulse loading-pulse-short" style={{ marginBottom: 12 }} />
      <div className="loading-pulse loading-pulse-long" style={{ marginBottom: 8 }} />
      <SkeletonCard />
      <SkeletonCard />
    </article>
  );
}