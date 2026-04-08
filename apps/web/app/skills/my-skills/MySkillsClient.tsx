"use client";

import { useQuery } from "@tanstack/react-query";
import { getMySkills } from "../../lib/api";
import { Badge } from "../../components/ui/Badge";
import { Card } from "../../components/ui/Card";

export function MySkillsClient() {
  const { data: skills, isLoading } = useQuery({
    queryKey: ["my-skills"],
    queryFn: getMySkills,
  });

  if (isLoading) {
    return <div className="p-6">Loading...</div>;
  }

  const drafts = skills?.filter(s => s.status === "draft") ?? [];
  const published = skills?.filter(s => s.status === "published") ?? [];

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">My Skills</h1>
      
      <section className="mb-8">
        <h2 className="text-lg font-semibold mb-4">Drafts ({drafts.length})</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {drafts.map((skill) => (
            <Card key={skill.id} className="p-4">
              <h3 className="font-semibold">{skill.name}</h3>
              <p className="text-sm text-muted mt-1">{skill.description}</p>
              <Badge variant="warning" className="mt-2">{skill.draft_source}</Badge>
            </Card>
          ))}
          {drafts.length === 0 && (
            <p className="text-muted col-span-full">No draft skills.</p>
          )}
        </div>
      </section>

      <section>
        <h2 className="text-lg font-semibold mb-4">Published ({published.length})</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {published.map((skill) => (
            <Card key={skill.id} className="p-4">
              <div className="flex justify-between items-start">
                <h3 className="font-semibold">{skill.name}</h3>
                {skill.verified && <Badge variant="success">Verified</Badge>}
              </div>
              <p className="text-sm text-muted mt-1">{skill.description}</p>
              {skill.sop_name && (
                <div className="mt-2 text-xs text-muted">
                  SOP: {skill.sop_name}
                </div>
              )}
            </Card>
          ))}
          {published.length === 0 && (
            <p className="text-muted col-span-full">No published skills.</p>
          )}
        </div>
      </section>
    </div>
  );
}