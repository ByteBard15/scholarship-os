import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import {
  applyImport,
  profileKeys,
  reviewCandidate,
} from "../features/profile/api";
import { useImportCandidates } from "../features/profile/queries";

export function ImportReviewPage() {
  const { id = "", importId = "" } = useParams();
  const query = useImportCandidates(id, importId);
  const client = useQueryClient();
  const refresh = () =>
    client.invalidateQueries({
      queryKey: [...profileKeys.detail(id), "imports", importId, "candidates"],
    });
  const review = useMutation({
    mutationFn: ({
      candidateId,
      action,
    }: {
      candidateId: string;
      action: "accept" | "reject" | "merge";
    }) => reviewCandidate(id, importId, candidateId, action),
    onSuccess: () => void refresh(),
  });
  const apply = useMutation({
    mutationFn: () => applyImport(id, importId),
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: profileKeys.detail(id) }),
  });
  if (query.isPending) return <p>Loading candidates…</p>;
  if (query.isError)
    return <p className="text-red-700">{query.error.message}</p>;
  return (
    <section>
      <Link to={`/profiles/${id}`}>← Back to profile</Link>
      <div className="mt-4 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">
            Review extracted candidates
          </h1>
          <p className="mt-1 text-sm text-slate-600">
            Nothing is written to the profile until every candidate is reviewed
            and accepted candidates are applied.
          </p>
        </div>
        <button
          className="rounded bg-indigo-700 px-4 py-2 text-white"
          onClick={() => apply.mutate()}
        >
          Apply accepted candidates
        </button>
      </div>
      {apply.isError && (
        <p className="mt-3 text-red-700">{apply.error.message}</p>
      )}
      <div className="mt-6 grid gap-4">
        {query.data.data.map((candidate) => (
          <article
            className="rounded-lg border bg-white p-5"
            key={candidate.id}
          >
            <div className="flex justify-between">
              <h2 className="font-medium capitalize">
                {candidate.sectionType.replace("_", " ")} candidate
              </h2>
              <span className="text-sm">
                {candidate.confidence === undefined
                  ? "Unknown confidence"
                  : `${Math.round(candidate.confidence * 100)}% confidence`}
              </span>
            </div>
            <pre className="mt-3 overflow-auto rounded bg-slate-50 p-3 text-xs">
              {JSON.stringify(candidate.candidateData, null, 2)}
            </pre>
            {candidate.matchedEntityId && (
              <p className="mt-2 text-sm text-amber-700">
                Possible duplicate found. Merge only records a proposed field
                diff and does not overwrite the canonical entity.
              </p>
            )}
            {candidate.changes && candidate.changes.length > 0 && (
              <ul className="mt-2 text-xs text-slate-600">
                {candidate.changes.map((change) => (
                  <li key={change.field}>
                    {change.field}: {String(change.existing)} →{" "}
                    {String(change.candidate)}
                  </li>
                ))}
              </ul>
            )}
            <div className="mt-4 flex gap-2">
              <button
                className="rounded border px-3 py-1"
                onClick={() =>
                  review.mutate({ candidateId: candidate.id, action: "accept" })
                }
              >
                Accept
              </button>
              <button
                className="rounded border px-3 py-1"
                onClick={() =>
                  review.mutate({ candidateId: candidate.id, action: "reject" })
                }
              >
                Reject
              </button>
              <button
                disabled={!candidate.matchedEntityId}
                className="rounded border px-3 py-1 disabled:opacity-40"
                onClick={() =>
                  review.mutate({ candidateId: candidate.id, action: "merge" })
                }
              >
                Review merge
              </button>
              <span className="ml-auto text-sm uppercase text-slate-500">
                {candidate.status}
              </span>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
