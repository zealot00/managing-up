"use client";

import { useState } from "react";

type Props = {
  title: string;
  data: unknown;
  defaultCollapsed?: boolean;
};

export default function JsonFold({ title, data, defaultCollapsed = true }: Props) {
  const [collapsed, setCollapsed] = useState(defaultCollapsed);

  return (
    <div className="json-fold">
      <button
        className="json-fold-header"
        onClick={() => setCollapsed(!collapsed)}
        type="button"
      >
        <span>{collapsed ? "▶" : "▼"} {title}</span>
        <span style={{ opacity: 0.6 }}>{collapsed ? "expand" : "collapse"}</span>
      </button>
      {!collapsed && (
        <div className="json-fold-content">
          <pre className="json-block" style={{ border: "none", padding: 0 }}>
            {JSON.stringify(data, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
}
