"use client";

import { ReactNode } from "react";

interface PageHeaderProps {
  eyebrow?: ReactNode;
  title: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
}

export function PageHeader({ eyebrow, title, description, actions }: PageHeaderProps) {
  const hasTitle = title !== undefined && title !== null && title !== "";
  return (
    <header className="page-header">
      <div className="page-header-text">
        {eyebrow && <p className="page-header-eyebrow">{eyebrow}</p>}
        {hasTitle && <h1 className="page-header-title">{title}</h1>}
        {description && <p className="page-header-desc">{description}</p>}
      </div>
      {actions && (
        <div className="page-header-actions">
          {actions}
        </div>
      )}
    </header>
  );
}
