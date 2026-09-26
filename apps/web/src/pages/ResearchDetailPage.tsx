import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import {
  attachTaskContexts,
  approveProposal,
  detachTaskContext,
  reviewProposal,
  transitionResearchTask,
  updateResearchTask,
  workflowKeys,
} from "../features/workflow/api";
import {
  useProposals,
  useResearchTask,
  useResearchContexts,
  useResearchTasks,
  useTaskFindings,
  useTaskContexts,
  useTaskLinks,
  useTaskOutputs,
  useTaskRuns,
  useTaskSources,
} from "../features/workflow/queries";
import { useProfiles } from "../features/profile/queries";
import { useState } from "react";
import { useAuth } from "../features/auth/AuthProvider";
import { CollapsibleText } from "../components/CollapsibleText";

export function ResearchDetailPage() {
  const { id = "" } = useParams();
  const { user } = useAuth();
  const userId = user?.id ?? "";
  const task = useResearchTask(id),
    links = useTaskLinks(id),
    taskContexts = useTaskContexts(id),
    contexts = useResearchContexts(userId),
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
      {item.status === "draft" && !contexts.isPending && !taskContexts.isPending && (
        <DraftTaskEditor
          contexts={contexts.data?.data ?? []}
          item={item}
          profiles={profiles.data?.data ?? []}
          selectedContexts={taskContexts.data?.data ?? []}
          onSaved={refresh}
        />
      )}
      <div className="mt-6 grid gap-5 md:grid-cols-2">
        <Panel title="Selected Profile">
          <p className="text-sm">
            {profiles.data?.data.find((profile) => profile.id === item.profileId)?.name ??
              "No profile selected"}
          </p>
        </Panel>
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
        <Panel title="Research Context" wide>
          {taskContexts.data?.data.length ? (
            taskContexts.data.data.map((context) => (
              <div className="border-b py-3 text-sm" key={context.id}>
                <p className="font-medium">{context.question}</p>
                <CollapsibleText className="mt-1 text-slate-600" lines={5}>
                  {context.answer}
                </CollapsibleText>
              </div>
            ))
          ) : (
            <p className="text-sm text-slate-500">
              No reusable context attached.
            </p>
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

function DraftTaskEditor({
  item,
  profiles,
  contexts,
  selectedContexts,
  onSaved,
}: {
  item: import("../features/workflow/types").ResearchTask;
  profiles: import("../features/profile/types").Profile[];
  contexts: import("../features/workflow/types").ResearchContext[];
  selectedContexts: import("../features/workflow/types").ResearchContext[];
  onSaved: () => void;
}) {
  const [title, setTitle] = useState(item.title);
  const [description, setDescription] = useState(item.description ?? "");
  const [instructions, setInstructions] = useState(item.instructions ?? "");
  const [taskType, setTaskType] = useState(item.taskType);
  const [priority, setPriority] = useState(item.priority ?? "normal");
  const [profileId, setProfileId] = useState(item.profileId ?? "");
  const [contextIds, setContextIds] = useState(
    selectedContexts.map((context) => context.id),
  );
  const save = useMutation({
    mutationFn: async () => {
      await updateResearchTask(item.id, {
        title,
        description,
        instructions,
        taskType,
        priority,
        profileId: profileId || undefined,
        clearProfile: !profileId,
      });
      const attached = new Set(selectedContexts.map((context) => context.id));
      const selected = new Set(contextIds);
      const additions = contextIds.filter((id) => !attached.has(id));
      const removals = selectedContexts.filter((context) => !selected.has(context.id));
      if (additions.length) await attachTaskContexts(item.id, additions);
      await Promise.all(
        removals.map((context) => detachTaskContext(item.id, context.id)),
      );
    },
    onSuccess: onSaved,
  });
  return (
    <form
      className="mt-6 grid gap-3 rounded-lg border border-indigo-200 bg-white p-5 md:grid-cols-2"
      onSubmit={(event) => {
        event.preventDefault();
        save.mutate();
      }}
    >
      <h2 className="font-semibold md:col-span-2">Edit Draft Research Task</h2>
      <input className="rounded border px-3 py-2 md:col-span-2" required value={title} onChange={(event) => setTitle(event.target.value)} />
      <textarea className="rounded border px-3 py-2" placeholder="Description" value={description} onChange={(event) => setDescription(event.target.value)} />
      <textarea className="rounded border px-3 py-2" placeholder="Instructions" value={instructions} onChange={(event) => setInstructions(event.target.value)} />
      <select className="rounded border px-3 py-2" value={taskType} onChange={(event) => setTaskType(event.target.value)}>
        <option value="scholarship_research">Scholarship research</option>
        <option value="scholarship_discovery">Scholarship discovery</option>
        <option value="programme_research">Programme research</option>
        <option value="funding_research">Funding research</option>
        <option value="general_research">General research</option>
      </select>
      <select className="rounded border px-3 py-2" value={priority} onChange={(event) => setPriority(event.target.value)}>
        <option value="low">Low</option><option value="normal">Normal</option><option value="high">High</option><option value="urgent">Urgent</option>
      </select>
      <label className="grid gap-1 text-sm md:col-span-2">Applicant profile
        <select className="rounded border px-3 py-2" required value={profileId} onChange={(event) => setProfileId(event.target.value)}>
          <option value="">Select a profile</option>
          {profiles.map((profile) => <option key={profile.id} value={profile.id}>{profile.name} ({profile.profileType})</option>)}
        </select>
      </label>
      <fieldset className="rounded border p-3 md:col-span-2">
        <legend className="px-2 text-sm font-medium">Attached research context</legend>
        <div className="grid gap-2">
          {contexts.map((context) => (
            <label className="flex gap-2 text-sm" key={context.id}>
              <input type="checkbox" checked={contextIds.includes(context.id)} onChange={(event) => setContextIds((current) => event.target.checked ? [...current, context.id] : current.filter((id) => id !== context.id))} />
              <span>
                <strong>{context.question}</strong>
                <CollapsibleText className="mt-1 text-slate-500" lines={5}>
                  {context.answer}
                </CollapsibleText>
              </span>
            </label>
          ))}
        </div>
      </fieldset>
      <button className="rounded bg-indigo-700 px-4 py-2 text-white md:col-span-2" disabled={save.isPending} type="submit">{save.isPending ? "Saving…" : "Save draft changes"}</button>
      {save.isError && <p className="text-sm text-red-700 md:col-span-2">{save.error.message}</p>}
    </form>
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
