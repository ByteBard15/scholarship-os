import { Link } from "react-router-dom";
import { useDashboard } from "../features/application/queries";
import {
  useInformationRequests,
  useProposals,
  useResearchTasks,
} from "../features/workflow/queries";
import { useAuth } from "../features/auth/AuthProvider";
export function DashboardPage() {
  const query = useDashboard();
  const { user } = useAuth();
  const userId = user?.id ?? "";
  const information = useInformationRequests(
    `?userId=${userId}&status=pending`,
  );
  const proposals = useProposals(userId);
  const research = useResearchTasks(`?userId=${userId}&status=review_required`);
  if (query.isPending) return <p>Loading dashboard…</p>;
  if (query.isError)
    return <p className="text-red-700">{query.error.message}</p>;
  const d = query.data.data;
  return (
    <section>
      <p className="text-sm font-medium uppercase tracking-wide text-indigo-700">
        Application workspace
      </p>
      <h1 className="mt-2 text-3xl font-semibold">Scholarship OS dashboard</h1>
      <div className="mt-6 grid gap-3 sm:grid-cols-4 lg:grid-cols-6">
        <Metric
          label="Active applications"
          value={d.activeApplications.length}
        />
        <Metric label="Need research" value={d.applicationsNeedingResearch} />
        <Metric label="Pending findings" value={d.researchPendingReview} />
        <Metric
          label="Research conflicts"
          value={d.unresolvedResearchConflicts}
        />
        <Metric
          label="Needs your input"
          value={information.data?.meta.count ?? 0}
        />
        <Metric
          label="Pending proposals"
          value={
            proposals.data?.data.filter(
              (proposal) => proposal.status === "pending",
            ).length ?? 0
          }
        />
      </div>
      <div className="mt-8 grid gap-6 md:grid-cols-2">
        <Panel title="Upcoming deadlines">
          {d.upcomingDeadlines.length ? (
            d.upcomingDeadlines.map((v) => (
              <Link
                className="block border-b py-2"
                key={v.id}
                to={`/applications/${v.applicationId}`}
              >
                {v.title}
                <span className="block text-xs text-slate-500">
                  {v.deadlineAt
                    ? new Date(v.deadlineAt).toLocaleDateString()
                    : "Unknown"}{" "}
                  · {v.urgency}
                </span>
              </Link>
            ))
          ) : (
            <p className="text-sm text-slate-500">
              No exact deadlines in the next 30 days.
            </p>
          )}
        </Panel>
        <Panel title="Tasks due soon">
          {d.tasksDueSoon.length ? (
            d.tasksDueSoon.map((v) => (
              <Link
                className="block border-b py-2"
                key={v.id}
                to={`/applications/${v.applicationId}`}
              >
                {v.title}
                <span className="block text-xs text-slate-500">
                  {v.dueAt
                    ? new Date(v.dueAt).toLocaleDateString()
                    : "No due date"}
                </span>
              </Link>
            ))
          ) : (
            <p className="text-sm text-slate-500">No tasks due soon.</p>
          )}
        </Panel>
        <Panel title="Applications requiring attention">
          {d.activeApplications.slice(0, 8).map((v) => (
            <Link
              className="block border-b py-2"
              key={v.id}
              to={`/applications/${v.id}`}
            >
              {v.name}
              <span className="block text-xs text-slate-500">
                {v.status} · research {v.researchStatus}
              </span>
            </Link>
          ))}
        </Panel>
        <Panel title="Operational gaps">
          <p className="text-sm">
            {d.incompleteMandatoryRequirements} incomplete mandatory
            requirements
          </p>
          <p className="mt-2 text-sm">
            {d.researchPendingReview} research findings need review
          </p>
          <p className="mt-2 text-xs text-slate-500">
            These counts measure workflow readiness, not admission likelihood.
          </p>
        </Panel>
        <Panel title="Research awaiting review">
          {research.data?.data.length ? (
            research.data.data.map((task) => (
              <Link
                key={task.id}
                className="block border-b py-2"
                to={`/research/${task.id}`}
              >
                {task.title}
                <span className="block text-xs text-slate-500">
                  Review sources, findings, and proposals
                </span>
              </Link>
            ))
          ) : (
            <p className="text-sm text-slate-500">
              No research tasks require review.
            </p>
          )}
        </Panel>
        <Panel title="Missing information">
          {information.data?.data.length ? (
            information.data.data.slice(0, 5).map((request) => (
              <Link
                key={request.id}
                className="block border-b py-2"
                to="/information-requests"
              >
                {request.title}
                <span className="block text-xs text-slate-500">
                  {request.prompt}
                </span>
              </Link>
            ))
          ) : (
            <p className="text-sm text-slate-500">
              No pending information requests.
            </p>
          )}
        </Panel>
      </div>
    </section>
  );
}
function Metric({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg border bg-white p-4">
      <p className="text-2xl font-semibold">{value}</p>
      <p className="text-xs text-slate-500">{label}</p>
    </div>
  );
}
function Panel({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="rounded-lg border bg-white p-5">
      <h2 className="font-semibold">{title}</h2>
      <div className="mt-3">{children}</div>
    </div>
  );
}
