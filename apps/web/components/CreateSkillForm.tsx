"use client";

import { useState, useTransition } from "react";
import { createSkill } from "../app/lib/api";
import { revalidatePath } from "next/cache";

export default function CreateSkillForm() {
  const [isOpen, setIsOpen] = useState(false);
  const [isPending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);

    setError(null);

    startTransition(async () => {
      try {
        await createSkill({
          name: formData.get("name") as string,
          owner_team: formData.get("owner_team") as string,
          risk_level: formData.get("risk_level") as string,
        });
        revalidatePath("/skills");
        setIsOpen(false);
        e.currentTarget.reset();
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to create skill");
      }
    });
  }

  return (
    <section className="form-panel">
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <p className="section-kicker">Actions</p>
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          style={{
            minHeight: "44px",
            padding: "0 20px",
            border: "1px solid var(--primary-deep)",
            borderRadius: "999px",
            background: isOpen ? "var(--primary-deep)" : "var(--primary)",
            color: "#fff",
            fontSize: "0.85rem",
            fontWeight: "700",
            letterSpacing: "0.06em",
            textTransform: "uppercase",
            cursor: "pointer",
            transition: "background 0.15s",
          }}
        >
          {isOpen ? "Cancel" : "New Skill"}
        </button>
      </div>

      {isOpen && (
        <form onSubmit={handleSubmit}>
          {error && <p className="form-error">{error}</p>}

          <div className="form-fields">
            <div>
              <label htmlFor="name" className="form-label">
                Name
              </label>
              <input
                id="name"
                name="name"
                type="text"
                required
                className="form-input"
                disabled={isPending}
              />
            </div>

            <div>
              <label htmlFor="owner_team" className="form-label">
                Owner Team
              </label>
              <input
                id="owner_team"
                name="owner_team"
                type="text"
                required
                className="form-input"
                disabled={isPending}
              />
            </div>

            <div>
              <label htmlFor="risk_level" className="form-label">
                Risk Level
              </label>
              <select
                id="risk_level"
                name="risk_level"
                required
                className="form-select"
                disabled={isPending}
              >
                <option value="">Select risk level</option>
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </div>
          </div>

          <div className="form-actions">
            <button type="submit" className="form-submit" disabled={isPending}>
              {isPending ? "Creating..." : "Create Skill"}
            </button>
          </div>
        </form>
      )}
    </section>
  );
}
