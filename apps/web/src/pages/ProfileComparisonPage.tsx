import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { compareProfiles, profileKeys } from "../features/profile/api";

export function ProfileComparisonPage() {
  const { id = "", otherId = "" } = useParams();
  const query = useQuery({
    queryKey: [...profileKeys.detail(id), "compare", otherId],
    queryFn: () => compareProfiles(id, otherId),
    enabled: Boolean(id && otherId),
  });
  if (query.isPending) return <p>Comparing profiles…</p>;
  if (query.isError)
    return <p className="text-red-700">{query.error.message}</p>;
  const comparison = query.data.data;
  return (
    <section>
      <Link to={`/profiles/${otherId}`}>← Back to profile</Link>
      <h1 className="mt-4 text-2xl font-semibold">
        {comparison.baseProfile.name} vs {comparison.comparedProfile.name}
      </h1>
      <div className="mt-6 grid gap-4 md:grid-cols-2">
        <Summary
          title="Inherited entities"
          count={comparison.inheritedEntities.length}
        />
        <Summary
          title="Appended entities"
          count={comparison.appendedEntities.length}
        />
        <Summary
          title="Hidden entities"
          count={comparison.hiddenEntities.length}
        />
        <Summary
          title="Modified fields"
          count={comparison.modifiedFields.length}
        />
      </div>
      <pre className="mt-6 overflow-auto rounded-lg border bg-white p-4 text-xs">
        {JSON.stringify(comparison, null, 2)}
      </pre>
    </section>
  );
}
function Summary({ title, count }: { title: string; count: number }) {
  return (
    <div className="rounded-lg border bg-white p-5">
      <p className="text-sm text-slate-500">{title}</p>
      <p className="mt-1 text-2xl font-semibold">{count}</p>
    </div>
  );
}
