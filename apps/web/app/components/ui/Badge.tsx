"use client";

type BadgeVariant =
  | "succeeded"
  | "completed"
  | "approved"
  | "published"
  | "running"
  | "pending"
  | "waiting"
  | "waiting_approval"
  | "draft"
  | "muted"
  | "rejected"
  | "failed"
  | "active"
  | "easy"
  | "medium"
  | "hard"
  | "low"
  | "high"
  | "warning"
  | "success"
  | "outline";

interface BadgeProps {
  variant: BadgeVariant;
  children: React.ReactNode;
  className?: string;
}

export function Badge({ variant, children, className = "" }: BadgeProps) {
  return (
    <span className={`badge badge-${variant} ${className}`}>
      {children}
    </span>
  );
}
