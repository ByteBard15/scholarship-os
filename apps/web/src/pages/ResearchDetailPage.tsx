import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import {
  approveProposal,
  reviewProposal,
  transitionResearchTask,
  workflowKeys,
} from "../features/workflow/api";
import {
  useProposals,
  useResearchTask,
  useResearchTasks,
  useTaskFindings,
  useTaskLinks,
  useTaskOutputs,
  useTaskRuns,
  useTaskSources,
} from "../features/workflow/queries";
import { useProfiles } from "../features/profile/queries";
import { useState } from "react";
import { useAuth } from "../features/auth/AuthProvider";

export function ResearchDetailPage() {
  const { id = "" } = useParams();
  const { user } = useAuth();
  const userId = user?.id ?? "";
  const task = useResearchTask(id),
    links = useTaskLinks(id),
    outputs = useTaskOutputs(id),
    runs = useTaskRuns(id),
    proposals = useProposals(userId),
    profiles = useProfiles(userId),
    relatedTasks = useResearchTasks(`?userId=${userId}`);
  const latestRun = runs.data?.data[0]?.id ?? "";
  const sources = useTaskSources(id, latestRun),
    findings = useTaskFindings(id, latestRun);
  const [parent, setParent] = useState("");
  const client = useQueryClient();
  const refresh = () =>
    client.invalidateQueries({ queryKey: workflowKeys.all });
  const transition = useMutation({
    mutationFn: (action: "queue" | "cancel" | "retry") =>
      transitionResearchTask(id, action),
    onSuccess: () => void refresh(),
  });
  const approve = useMutation({
    mutationFn: (proposalId: string) => approveProposal(proposalId, parent),
    onSuccess: () => void refresh(),
  });
  const review = useMutation({
    mutationFn: ({
      proposalId,
      action,
    }: {
      proposalId: string;
      action: "reject" | "reopen";
    }) => reviewProposal(proposalId, action),
    onSuccess: () => void refresh(),
  });
  if (task.isPending) return <p>Loading research task…</p>;
  if (task.isError) return <p className="text-red-700">{task.error.message}</p>;
  const item = task.data.data,
    taskProposals =
      proposals.data?.data.filter(
        (proposal) => proposal.researchTaskId === id,
      ) ?? [];
  const parents =
    profiles.data?.data.filter(
      (profile) =>
        profile.profileType === "master" || profile.profileType === "domain",
    ) ?? [];
  return (
    <section>
      <Link className="text-sm text-indigo-700" to="/research">
        ← Research Inbox
      </Link>
      <h1 className="mt-3 text-3xl font-semibold">{item.title}</h1>
      <p className="mt-2 text-slate-600">
        {item.description ?? "No description"}
      </p>
      <div className="mt-4 flex gap-2">
        <span className="rounded bg-slate-100 px-2 py-1 text-xs uppercase">
          {item.status.replaceAll("_", " ")}
        </span>
        {["draft", "ready"].includes(item.status) && (
          <button
            className="rounded border px-3 py-1 text-sm"
            onClick={() => transition.mutate("queue")}
          >
            Queue
          </button>
        )}
        {item.status === "queued" && (
          <span className="rounded bg-indigo-50 px-3 py-1 text-sm text-indigo-700">
            Waiting for a research agent
          </span>
        )}
        {item.status === "failed" && (
          <button
            className="rounded border px-3 py-1 text-sm"
            onClick={() => transition.mutate("retry")}
          >
            Retry
          </button>
        )}
        {["queued", "running", "review_required"].includes(item.status) && (
          <button
            className="rounded border px-3 py-1 text-sm"
            onClick={() => transition.mutate("cancel")}
          >
            Cancel
          </button>
        )}
      </div>
      <div className="mt-6 grid gap-5 md:grid-cols-2">
        <Panel title="Instructions">
          <p className="text-sm whitespace-pre-wrap">
            {item.instructions ?? "No custom instructions."}
          </p>
        </Panel>
        <Panel title="Links">
          {links.data?.data.length ? (
            links.data.data.map((link) => (
              <a
                key={link.id}
                className="block text-sm text-indigo-700"
                href={link.url}
                target="_blank"
                rel="noreferrer"
              >
                {link.label ?? link.url}
              </a>
            ))
          ) : (
            <p className="text-sm text-slate-500">No links.</p>
          )}
        </Panel>
      </div>
      <Panel title="Research Results" wide>
        <p className="text-sm text-slate-600">
          {sources.data?.data.length ?? 0} sources ·{" "}
          {findings.data?.data.length ?? 0} findings ·{" "}
          {runs.data?.data.length ?? 0} runs
        </p>
        <div className="mt-3 grid gap-2">
          {sources.data?.data.map((source) => (
            <a
              key={source.id}
              href={source.url}
              className="rounded border p-3 text-sm text-indigo-700"
              target="_blank"
              rel="noreferrer"
            >
              {source.title ?? source.url}
              {source.isOfficial ? " · official" : ""}
            </a>
          ))}
          {findings.data?.data.map((finding) => (
            <div key={finding.id} className="rounded border p-3 text-sm">
              <strong>
                {finding.category}: {finding.field}
              </strong>
              <p className="text-xs text-slate-500">
                {finding.verificationStatus} · review {finding.reviewStatus}
              </p>
            </div>
          ))}
        </div>
      </Panel>
      <Panel title="Application Proposals" wide>
        <div className="mb-3">
          <select
            className="rounded border px-3 py-2"
            value={parent}
            onChange={(event) => setParent(event.target.value)}
          >
            <option value="">Choose parent profile before approval</option>
            {parents.map((profile) => (
              <option key={profile.id} value={profile.id}>
                {profile.name} ({profile.profileType})
              </option>
            ))}
          </select>
        </div>
        {taskProposals.length ? (
          taskProposals.map((proposal) => (
            <article key={proposal.id} className="mb-3 rounded border p-4">
              <div className="flex justify-between">
                <h3 className="font-medium">{proposal.name}</h3>
                <span className="text-xs uppercase">{proposal.status}</span>
              </div>
              <p className="mt-2 text-sm">
                {proposal.proposedInstitution?.name ?? "Institution unknown"} ·{" "}
                {proposal.proposedProgramme?.name ?? "Programme unknown"} ·{" "}
                {proposal.proposedScholarship?.name ?? "Scholarship unknown"}
              </p>
              <p className="mt-2 text-sm text-slate-600">{proposal.summary}</p>
              <p className="mt-2 text-xs text-slate-500">
                Confidence:{" "}
                {proposal.confidence == null
                  ? "not supplied"
                  : `${Math.round(proposal.confidence * 100)}%`}{" "}
                · {proposal.sources.length} provenance source(s)
              </p>
              <div className="mt-3 flex gap-2">
                {proposal.status === "pending" && (
                  <>
                    <button
                      disabled={!parent}
                      onClick={() => approve.mutate(proposal.id)}
                      className="rounded bg-emerald-700 px-3 py-1.5 text-sm text-white disabled:opacity-40"
                    >
                      Approve Application
                    </button>
                    <button
                      onClick={() =>
                        review.mutate({
                          proposalId: proposal.id,
                          action: "reject",
                        })
                      }
                      className="rounded border px-3 py-1.5 text-sm"
                    >
                      Reject
                    </button>
                  </>
                )}
                {proposal.status === "rejected" && (
                  <button
                    onClick={() =>
                      review.mutate({
                        proposalId: proposal.id,
                        action: "reopen",
                      })
                    }
                    className="rounded border px-3 py-1.5 text-sm"
                  >
                    Reopen
                  </button>
                )}
              </div>
            </article>
          ))
        ) : (
          <p className="text-sm text-slate-500">No proposals yet.</p>
        )}
      </Panel>
      <div className="grid gap-5 md:grid-cols-2">
        <Panel title="Outputs">
          {outputs.data?.data.length ? (
            outputs.data.data.map((output) => (
              <p key={output.id} className="border-b py-2 text-sm">
                {output.outputType.replaceAll("_", " ")} · {output.entityId}
              </p>
            ))
          ) : (
            <p className="text-sm text-slate-500">No outputs recorded.</p>
          )}
        </Panel>
        <Panel title="Follow-up Tasks">
          {relatedTasks.data?.data.filter(
            (candidate) => candidate.parentTaskId === id,
          ).length ? (
            relatedTasks.data?.data
              .filter((candidate) => candidate.parentTaskId === id)
              .map((candidate) => (
                <Link
                  key={candidate.id}
                  to={`/research/${candidate.id}`}
                  className="block border-b py-2 text-sm text-indigo-700"
                >
                  {candidate.title} · {candidate.status}
                </Link>
              ))
          ) : (
            <p className="text-sm text-slate-500">No follow-up tasks.</p>
          )}
        </Panel>
      </div>
      <Panel title="Activity" wide>
        <p className="text-sm">
          Task created {new Date(item.createdAt).toLocaleString()}
        </p>
        {item.startedAt && (
          <p className="mt-2 text-sm">
            Research started {new Date(item.startedAt).toLocaleString()}
          </p>
        )}
        {runs.data?.data.map((run) => (
          <p key={run.id} className="mt-2 text-sm">
            Run {run.id} · {run.status} ·{" "}
            {new Date(run.createdAt).toLocaleString()}
          </p>
        ))}
        {item.completedAt && (
          <p className="mt-2 text-sm">
            Task completed {new Date(item.completedAt).toLocaleString()}
          </p>
        )}
      </Panel>
    </section>
  );
}
function Panel({
  title,
  children,
  wide = false,
}: {
  title: string;
  children: React.ReactNode;
  wide?: boolean;
}) {
  return (
    <section
      className={`mt-5 rounded-lg border bg-white p-5 ${wide ? "md:col-span-2" : ""}`}
    >
      <h2 className="font-semibold">{title}</h2>
      <div className="mt-3">{children}</div>
    </section>
  );
}
