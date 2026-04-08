"use client";

import { useQuery } from "@tanstack/react-query";
import { getSkillMarket } from "../../lib/api";
import { Badge } from "../../components/ui/Badge";
import { Card } from "../../components/ui/Card";

export function SkillMarketClient() {
  const { data: skills, isLoading } = useQuery({
    queryKey: ["skill-market"],
    queryFn: () => getSkillMarket({}),
  });

  if (isLoading) {
    return <div className="p-6">Loading...</div>;
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">Skill Market</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {skills?.map((skill) => (
          <Card key={skill.id} className="p-4">
            <div className="flex justify-between items-start mb-2">
              <h3 className="font-semibold">{skill.name}</h3>
              {skill.verified && (
                <Badge variant="success">Verified</Badge>
              )}
            </div>
            <p className="text-sm text-muted mb-3">{skill.description}</p>
            
            <div className="flex flex-wrap gap-1 mb-3">
              {skill.tags?.map((tag) => (
                <Badge key={tag} variant="outline">{tag}</Badge>
              ))}
            </div>
            
            <div className="flex justify-between items-center text-sm">
              <span className="text-muted">
                Trust: {skill.trust_score.toFixed(2)}
              </span>
              {skill.avg_rating > 0 && (
                <span>⭐ {skill.avg_rating.toFixed(1)} ({skill.rating_count})</span>
              )}
            </div>
            
            {skill.sop_name && (
              <div className="mt-2 text-xs text-muted">
                SOP: {skill.sop_name}
              </div>
            )}
          </Card>
        ))}
      </div>
      
      {(!skills || skills.length === 0) && (
        <div className="text-center py-12 text-muted">
          No skills found in the market.
        </div>
      )}
    </div>
  );
}
