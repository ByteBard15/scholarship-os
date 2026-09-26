import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  applicationKeys,
  applyResearch,
  completeTask,
  reviewFinding,
} from "../features/application/api";
import {
  useApplication,
  useResearch,
  useResearchFindings,
  useResearchSources,
} from "../features/application/queries";
import { runPrefill, workflowKeys } from "../features/workflow/api";
import { useApplicationWorkspace } from "../features/workflow/queries";

const tabs = [
  "Overview",
  "Requirements",
  "Questionnaires",
  "Information Requests",
  "Fields",
  "Deadlines",
  "Funding",
  "Tasks",
  "Research",
  "Contacts",
  "Supervisors",
  "Profile",
  "Links",
] as const;
export function ApplicationDetailPage() {
  const { id = "" } = useParams();
  const [tab, setTab] = useState<(typeof tabs)[number]>("Overview");
  const app = useApplication(id);
  const runs = useResearch(id);
  const latest = runs.data?.data[0]?.id ?? "";
  const sources = useResearchSources(id, latest);
  const findings = useResearchFindings(id, latest);
  const workspace = useApplicationWorkspace(id);
  const client = useQueryClient();
  const refresh = () =>
    client.invalidateQueries({ queryKey: applicationKeys.all });
  const review = useMutation({
    mutationFn: ({
      finding,
      status,
    }: {
      finding: string;
      status: "accepted" | "rejected";
    }) => reviewFinding(id, latest, finding, status),
    onSuccess: () => void refresh(),
  });
  const apply = useMutation({
    mutationFn: () => applyResearch(id, latest),
    onSuccess: () => void refresh(),
  });
  const task = useMutation({
    mutationFn: ({ taskId, done }: { taskId: string; done: boolean }) =>
      completeTask(id, taskId, !done),
    onSuccess: () => void refresh(),
  });
  const prefill = useMutation({
    mutationFn: () => runPrefill(id),
    onSuccess: () => {
      void client.invalidateQueries({ queryKey: workflowKeys.workspace(id) });
    },
  });
  if (app.isPending) return <p>Loading application…</p>;
  if (app.isError) return <p className="text-red-700">{app.error.message}</p>;
  const d = app.data.data,
    a = d.application;
  return (
    <section>
      <p className="text-sm uppercase text-indigo-700">
        {a.status.replaceAll("_", " ")}
      </p>
      <h1 className="mt-1 text-3xl font-semibold">{a.name}</h1>
      <p className="mt-2 text-slate-600">
        {a.institution?.name ?? "Institution unknown"} · Research{" "}
        {a.researchStatus}
      </p>
      <nav className="mt-7 flex flex-wrap gap-2">
        {tabs.map((v) => (
          <button
            key={v}
            onClick={() => setTab(v)}
            className={`rounded px-3 py-1.5 text-sm ${tab === v ? "bg-indigo-700 text-white" : "border bg-white"}`}
          >
            {v}
          </button>
        ))}
      </nav>
      <div className="mt-6 rounded-lg border bg-white p-5">
        {tab === "Overview" && (
          <div className="grid gap-3 md:grid-cols-2">
            <Info label="Institution" value={a.institution?.name} />
            <Info label="Programme" value={a.programme?.name} />
            <Info label="Scholarship" value={a.scholarship?.name} />
            <Info
              label="Intake"
              value={String(a.intakeYear ?? a.intake ?? "Unknown")}
            />
            <Info
              label="Requirements"
              value={`${d.requirements.filter((v) => v.status === "satisfied").length}/${d.requirements.length} satisfied`}
            />
            <Info
              label="Tasks"
              value={`${d.tasks.filter((v) => v.status === "done").length}/${d.tasks.length} complete`}
            />
            <Info
              label="Questionnaires"
              value={`${workspace.preparation.data?.data.questionnairesCompleted ?? 0}/${workspace.preparation.data?.data.questionnairesTotal ?? 0} complete`}
            />
            <Info
              label="Pending information"
              value={String(
                workspace.preparation.data?.data.pendingInformation ?? 0,
              )}
            />
          </div>
        )}
        {tab === "Requirements" && (
          <List empty="No requirements yet.">
            {d.requirements.map((v) => (
              <Card
                key={v.id}
                title={v.title}
                meta={`${v.category} · ${v.isMandatory ? "mandatory" : "optional"} · ${v.status}`}
              >
                <Source sourceId={v.sourceId} sourceUrl={v.sourceUrl} />
              </Card>
            ))}
          </List>
        )}
        {tab === "Questionnaires" && (
          <div>
            <div className="mb-4 flex items-center justify-between">
              <p className="text-sm text-slate-600">
                Factual answers may be filled from evidence. Subjective drafts
                remain suggestions until approved.
              </p>
              <button
                className="rounded bg-indigo-700 px-3 py-2 text-sm text-white"
                disabled={prefill.isPending}
                onClick={() => prefill.mutate()}
              >
                Run Prefill
              </button>
            </div>
            <List empty="No questionnaires yet.">
              {(workspace.questionnaires.data?.data ?? []).map((v) => (
                <Link
                  key={v.id}
                  className="block rounded border p-3"
                  to={`/applications/${id}/questionnaires/${v.id}`}
                >
                  <p className="font-medium">{v.title}</p>
                  <p className="text-xs text-slate-500">
                    {v.questionnaireType ?? "questionnaire"} · {v.status}
                  </p>
                </Link>
              ))}
            </List>
          </div>
        )}
        {tab === "Information Requests" && (
          <List empty="No missing information requests.">
            {(workspace.information.data?.data ?? []).map((v) => (
              <Link
                key={v.id}
                to="/information-requests"
                className="block rounded border p-3"
              >
                <p className="font-medium">{v.title}</p>
                <p className="mt-1 text-sm">{v.prompt}</p>
                <p className="mt-1 text-xs uppercase text-slate-500">
                  {v.status}
                </p>
              </Link>
            ))}
          </List>
        )}
        {tab === "Fields" && (
          <div>
            <p className="mb-4 text-sm text-slate-600">
              Miscellaneous application-form fields with explicit provenance.
            </p>
            <List empty="No application fields yet.">
              {(workspace.fields.data?.data ?? []).map((v) => (
                <Card
                  key={v.id}
                  title={v.label}
                  meta={`${v.status} · ${v.sourceType ?? "manual"}`}
                >
                  <p className="mt-2 text-sm">
                    {v.value == null ? "Not filled" : JSON.stringify(v.value)}
                  </p>
                </Card>
              ))}
            </List>
          </div>
        )}
        {tab === "Deadlines" && (
          <List empty="No deadlines yet.">
            {d.deadlines.map((v) => (
              <Card
                key={v.id}
                title={v.title}
                meta={`${v.deadlineType} · ${v.datePrecision} · ${v.urgency}`}
              >
                <p className="text-sm">
                  {v.deadlineAt
                    ? new Date(v.deadlineAt).toLocaleString()
                    : (v.rawDeadlineText ?? "Exact date unknown")}
                </p>
                <Source sourceId={v.sourceId} sourceUrl={v.sourceUrl} />
              </Card>
            ))}
          </List>
        )}
        {tab === "Funding" && (
          <List empty="Funding information is unknown.">
            {d.funding.map((v) => (
              <Card
                key={v.id}
                title={v.fundingType}
                meta={v.tuitionCoverage ?? "Tuition coverage unknown"}
              >
                <p className="text-sm text-slate-600">
                  Stipend:{" "}
                  {v.stipendAmount == null
                    ? "Unknown"
                    : `${v.currency ?? ""} ${v.stipendAmount} ${v.stipendPeriod ?? ""}`}{" "}
                  · Travel: {v.travelCoverage ?? "Unknown"} · Insurance:{" "}
                  {v.insuranceCoverage ?? "Unknown"} · Accommodation:{" "}
                  {v.accommodationCoverage ?? "Unknown"}
                </p>
              </Card>
            ))}
          </List>
        )}
        {tab === "Tasks" && (
          <List empty="No checklist tasks.">
            {d.tasks.map((v) => (
              <div
                key={v.id}
                className="flex items-center gap-3 rounded border p-3"
              >
                <input
                  type="checkbox"
                  checked={v.status === "done"}
                  onChange={() =>
                    task.mutate({ taskId: v.id, done: v.status === "done" })
                  }
                />
                <div>
                  <p
                    className={
                      v.status === "done" ? "line-through text-slate-500" : ""
                    }
                  >
                    {v.title}
                  </p>
                  <p className="text-xs text-slate-500">
                    {v.taskType ?? "task"}
                    {v.dueAt
                      ? ` · due ${new Date(v.dueAt).toLocaleDateString()}`
                      : ""}
                    {v.parentTaskId ? " · subtask" : ""}
                  </p>
                </div>
              </div>
            ))}
          </List>
        )}
        {tab === "Research" && (
          <ResearchPanel
            runs={runs.data?.data ?? []}
            sources={sources.data?.data ?? []}
            findings={findings.data?.data ?? []}
            onReview={(finding, status) => review.mutate({ finding, status })}
            onApply={() => apply.mutate()}
            busy={review.isPending || apply.isPending}
          />
        )}
        {tab === "Contacts" && (
          <List empty="No contacts.">
            {d.contacts.map((v) => (
              <Card
                key={v.id}
                title={v.name ?? v.organization ?? "Contact"}
                meta={`${v.contactType ?? "contact"} · ${v.email ?? "email unknown"}`}
              />
            ))}
          </List>
        )}
        {tab === "Supervisors" && (
          <List empty="No supervisors.">
            {d.supervisors.map((v) => (
              <Card
                key={v.id}
                title={v.name}
                meta={`${v.department ?? "Department unknown"} · ${v.contactStatus ?? "contact status unknown"}`}
              />
            ))}
          </List>
        )}
        {tab === "Profile" && (
          <div>
            <p className="text-sm text-slate-600">
              Application-specific facts and emphasis are isolated in the linked
              derived profile.
            </p>
            <Link
              className="mt-3 inline-block text-indigo-700"
              to={`/profiles/${a.applicantProfileId}`}
            >
              Open application profile and lineage →
            </Link>
          </div>
        )}
        {tab === "Links" && (
          <List empty="No application links.">
            {d.urls.map((v) => (
              <a
                key={v.id}
                className="block rounded border p-3 text-indigo-700"
                href={v.url}
                target="_blank"
                rel="noreferrer"
              >
                {v.label ?? v.urlType}
                {v.isOfficial ? " · official" : ""}
              </a>
            ))}
          </List>
        )}
      </div>
    </section>
  );
}
function Info({ label, value }: { label: string; value?: string }) {
  return (
    <div>
      <p className="text-xs uppercase text-slate-500">{label}</p>
      <p>{value ?? "Unknown"}</p>
    </div>
  );
}
function List({
  children,
  empty,
}: {
  children: React.ReactNode;
  empty: string;
}) {
  return (
    <div className="grid gap-3">
      {Array.isArray(children) && children.length === 0 ? (
        <p className="text-slate-500">{empty}</p>
      ) : (
        children
      )}
    </div>
  );
}
function Card({
  title,
  meta,
  children,
}: {
  title: string;
  meta: string;
  children?: React.ReactNode;
}) {
  return (
    <div className="rounded border p-3">
      <p className="font-medium">{title}</p>
      <p className="mt-1 text-xs text-slate-500">{meta}</p>
      {children}
    </div>
  );
}
function Source({
  sourceId,
  sourceUrl,
}: {
  sourceId?: string;
  sourceUrl?: string;
}) {
  return (
    <p className="mt-2 text-xs text-slate-500">
      {sourceId
        ? "Research provenance recorded"
        : sourceUrl
          ? "Manual source URL recorded"
          : "Manual / no source"}
      {sourceUrl && (
        <>
          {" "}
          ·{" "}
          <a className="text-indigo-700" href={sourceUrl}>
            source
          </a>
        </>
      )}
    </p>
  );
}
function ResearchPanel({
  runs,
  sources,
  findings,
  onReview,
  onApply,
  busy,
}: {
  runs: import("../features/application/types").ResearchRun[];
  sources: import("../features/application/types").ResearchSource[];
  findings: import("../features/application/types").ResearchFinding[];
  onReview: (id: string, status: "accepted" | "rejected") => void;
  onApply: () => void;
  busy: boolean;
}) {
  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-semibold">Research review</h2>
          <p className="text-xs text-slate-500">
            Last researched:{" "}
            {runs[0]?.completedAt
              ? new Date(runs[0].completedAt).toLocaleDateString()
              : "Never"}
          </p>
        </div>
        <Link className="rounded border px-3 py-2 text-sm" to="/research">
          Open Research Inbox
        </Link>
      </div>
      <div className="mt-5">
        <h3 className="font-medium">Sources</h3>
        {sources.map((v) => (
          <a
            className="mt-2 block text-sm text-indigo-700"
            key={v.id}
            href={v.url}
          >
            {v.title ?? v.url} · {v.isOfficial ? "official" : "secondary"}
          </a>
        ))}
      </div>
      <div className="mt-5 grid gap-3">
        <h3 className="font-medium">Findings</h3>
        {findings.map((v) => (
          <div
            key={v.id}
            className={`rounded border p-3 ${v.verificationStatus === "conflicting" ? "border-red-400" : ""}`}
          >
            <div className="flex justify-between gap-3">
              <span>
                {v.category}: {v.field}
              </span>
              <span className="text-xs">{v.reviewStatus}</span>
            </div>
            <pre className="mt-2 overflow-auto whitespace-pre-wrap text-xs text-slate-600">
              {JSON.stringify(v.value, null, 2)}
            </pre>
            <p className="text-xs text-slate-500">
              Confidence:{" "}
              {v.confidence == null
                ? "unknown"
                : Math.round(v.confidence * 100) + "%"}{" "}
              · {v.verificationStatus}
            </p>
            {v.reviewStatus === "pending" && (
              <div className="mt-2 flex gap-2">
                <button
                  className="text-sm text-green-700"
                  onClick={() => onReview(v.id, "accepted")}
                >
                  Accept
                </button>
                <button
                  className="text-sm text-red-700"
                  onClick={() => onReview(v.id, "rejected")}
                >
                  Reject
                </button>
              </div>
            )}
          </div>
        ))}
      </div>
      {findings.length > 0 && (
        <button
          className="mt-4 rounded border px-3 py-2 text-sm disabled:opacity-50"
          disabled={busy || findings.some((v) => v.reviewStatus === "pending")}
          onClick={onApply}
        >
          Apply accepted findings
        </button>
      )}
    </div>
  );
}
