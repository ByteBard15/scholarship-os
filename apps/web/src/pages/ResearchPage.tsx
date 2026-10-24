import { ScholarshipDeadlines } from "../components/ScholarshipDeadlines";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { FormEvent, useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import {
  createResearchTask,
  exportResearchTasks,
  importResearchTasks,
  workflowKeys,
} from "../features/workflow/api";
import {
  useResearchContexts,
  useResearchTasks,
} from "../features/workflow/queries";
import { useAuth } from "../features/auth/AuthProvider";
import { useProfiles } from "../features/profile/queries";
import { CollapsibleText } from "../components/CollapsibleText";

import { Dialog } from "../components/Dialog";
import { RequestButton } from "../components/RequestButton";

export function ResearchPage() {
  const { user } = useAuth();
  const userId = user?.id ?? "";
  const [createOpen, setCreateOpen] = useState(false);
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Math.max(1, Number(searchParams.get("page")) || 1);
  const pageSize = [10, 20, 50, 100].includes(
    Number(searchParams.get("pageSize")),
  )
    ? Number(searchParams.get("pageSize"))
    : 20;
  const sortBy = searchParams.get("sortBy") || "priority";
  const status = searchParams.get("status") || "";
  const taskType = searchParams.get("taskType") || "";
  const priority = searchParams.get("priority") || "";
  const proposalState = searchParams.get("proposalState") || "";
  const [pageAction, setPageAction] = useState<"previous" | "next" | null>(
    null,
  );
  function setFilter(key: string, value: string) {
    setSearchParams(
      (previous) => {
        const next = new URLSearchParams(previous);
        if (value) next.set(key, value);
        else next.delete(key);
        next.set("page", "1");
        return next;
      },
      { replace: true },
    );
  }
  function setPage(value: number) {
    setSearchParams(
      (previous) => {
        const next = new URLSearchParams(previous);
        next.set("page", String(value));
        return next;
      },
      { replace: true },
    );
  }
  const query = new URLSearchParams(
    Object.fromEntries(
      Object.entries({
        userId,
        status,
        taskType,
        priority,
        sortBy,
        proposalState,
      }).filter(([, value]) => value),
    ),
  ).toString();
  const tasks = useResearchTasks(`?${query}&page=${page}&pageSize=${pageSize}`);
  useEffect(() => {
    const totalPages = tasks.data?.meta.totalPages;
    if (totalPages !== undefined && page > Math.max(1, totalPages))
      setPage(Math.max(1, totalPages));
  }, [tasks.data?.meta.totalPages, page]);
  const contexts = useResearchContexts(userId);
  const profiles = useProfiles(userId);
  const client = useQueryClient();
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [instructions, setInstructions] = useState("");
  const [type, setType] = useState("scholarship_research");
  const [taskPriority, setTaskPriority] = useState("normal");
  const [link, setLink] = useState("");
  const [profileId, setProfileId] = useState("");
  const [selectedContextIds, setSelectedContextIds] = useState<string[]>([]);
  const [importText, setImportText] = useState("");
  const [importQueue, setImportQueue] = useState(false);
  const [importMessage, setImportMessage] = useState("");
  const [newContexts, setNewContexts] = useState([
    { question: "", answer: "" },
  ]);
  const create = useMutation({
    mutationFn: (queue: boolean) =>
      createResearchTask(
        {
          userId,
          title,
          description: description || undefined,
          instructions: instructions || undefined,
          taskType: type,
          priority: taskPriority,
          profileId: profileId || undefined,
          links: link ? [{ url: link, linkType: "official" }] : [],
          researchContextIds: selectedContextIds,
          newResearchContexts: newContexts.filter(
            (item) => item.question.trim() && item.answer.trim(),
          ),
        },
        queue,
      ),
    onSuccess: () => {
      setCreateOpen(false);
      setPage(1);
      setTaskPriority("normal");
      setType("scholarship_research");
      setTitle("");
      setDescription("");
      setInstructions("");
      setLink("");
      setProfileId("");
      setSelectedContextIds([]);
      setNewContexts([{ question: "", answer: "" }]);
      void client.invalidateQueries({ queryKey: workflowKeys.all });
    },
  });
  const importTasks = useMutation({
    mutationFn: () => {
      const parsed = JSON.parse(importText) as {
        tasks?: Array<{
          title: string;
          description?: string;
          instructions?: string;
          taskType: string;
          priority?: string;
          dueAt?: string;
          targetApplicationId?: string;
          profileId?: string;
          links?: Array<{ label?: string; url: string; linkType?: string }>;
          researchContexts?: Array<{ question: string; answer: string }>;
          researchContextIds?: string[];
        }>;
      };
      return importResearchTasks({
        userId,
        queue: importQueue,
        tasks: (parsed.tasks ?? []).map((task) => ({
          ...task,
          researchContextIds: task.researchContextIds ?? [],
          researchContexts: task.researchContexts ?? [],
          links: task.links ?? [],
        })),
      });
    },
    onSuccess: (result) => {
      setImportText("");
      setImportMessage(`Imported ${result.meta.count} research task(s).`);
      void client.invalidateQueries({ queryKey: workflowKeys.all });
    },
    onError: (error) => {
      setImportMessage(
        error instanceof Error ? error.message : "Import failed",
      );
    },
  });
  const exportTasks = useMutation({ mutationFn: downloadExport });
  async function downloadExport() {
    const value = await exportResearchTasks(query ? `?${query}` : "");
    const blob = new Blob([JSON.stringify(value.data, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "research-tasks.json";
    anchor.click();
    URL.revokeObjectURL(url);
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    const submitter = (event.nativeEvent as SubmitEvent)
      .submitter as HTMLButtonElement | null;
    if (!create.isPending) create.mutate(submitter?.value === "queue");
  }
  return (
    <section>
      <h1 className="text-3xl font-semibold">Research Inbox</h1>
      <p className="mt-2 text-slate-600">
        Create controlled research work. Results become proposals and never
        Applications until you approve them.
      </p>
      <button
        type="button"
        className="mt-5 rounded bg-indigo-700 px-4 py-2 text-white"
        onClick={() => {
          create.reset();
          setCreateOpen(true);
        }}
      >
        New research task
      </button>
      <Dialog
        open={createOpen}
        title="New research task"
        busy={create.isPending}
        onClose={() => setCreateOpen(false)}
      >
        <form
          onSubmit={submit}
          className="grid gap-3 bg-white p-5 md:grid-cols-2"
        >
          <fieldset disabled={create.isPending} className="contents">
            <input
              required
              className="rounded border px-3 py-2 md:col-span-2"
              placeholder="Research task title"
              value={title}
              onChange={(event) => setTitle(event.target.value)}
            />
            <textarea
              className="rounded border px-3 py-2"
              placeholder="Description"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
            />
            <textarea
              className="rounded border px-3 py-2"
              placeholder="Custom instructions"
              value={instructions}
              onChange={(event) => setInstructions(event.target.value)}
            />
            <input
              className="rounded border px-3 py-2 md:col-span-2"
              type="url"
              placeholder="Official or reference URL (optional)"
              value={link}
              onChange={(event) => setLink(event.target.value)}
            />
            <label className="grid gap-1 text-sm md:col-span-2">
              Applicant profile
              <select
                className="rounded border px-3 py-2"
                required
                value={profileId}
                onChange={(event) => setProfileId(event.target.value)}
              >
                <option value="">Select a profile</option>
                {profiles.data?.data.map((profile) => (
                  <option key={profile.id} value={profile.id}>
                    {profile.name} ({profile.profileType})
                  </option>
                ))}
              </select>
            </label>
            <fieldset className="rounded border p-4 md:col-span-2">
              <legend className="px-2 text-sm font-medium">
                Reusable research context
              </legend>
              <p className="mb-3 text-xs text-slate-500">
                Select saved answers or add new question-and-answer context. New
                entries are saved to your context library and can be reused by
                later research tasks.
              </p>
              {contexts.data?.data.length ? (
                <div className="mb-4 grid gap-2">
                  {contexts.data.data.map((context) => (
                    <label
                      className="flex gap-2 rounded bg-slate-50 p-3 text-sm"
                      key={context.id}
                    >
                      <input
                        checked={selectedContextIds.includes(context.id)}
                        onChange={(event) =>
                          setSelectedContextIds((current) =>
                            event.target.checked
                              ? [...current, context.id]
                              : current.filter((id) => id !== context.id),
                          )
                        }
                        type="checkbox"
                      />
                      <span>
                        <strong>{context.question}</strong>
                        <CollapsibleText
                          className="mt-1 text-slate-600"
                          lines={5}
                        >
                          {context.answer}
                        </CollapsibleText>
                      </span>
                    </label>
                  ))}
                </div>
              ) : null}
              <div className="grid gap-3">
                {newContexts.map((context, index) => (
                  <div
                    className="grid gap-2 rounded bg-slate-50 p-3"
                    key={index}
                  >
                    <input
                      className="rounded border px-3 py-2"
                      placeholder="Question, e.g. What courses interest you?"
                      required={Boolean(
                        context.question.trim() || context.answer.trim(),
                      )}
                      value={context.question}
                      onChange={(event) =>
                        setNewContexts((current) =>
                          current.map((item, itemIndex) =>
                            itemIndex === index
                              ? { ...item, question: event.target.value }
                              : item,
                          ),
                        )
                      }
                    />
                    <textarea
                      className="rounded border px-3 py-2"
                      placeholder="Answer or existing personal-statement material"
                      required={Boolean(
                        context.question.trim() || context.answer.trim(),
                      )}
                      value={context.answer}
                      onChange={(event) =>
                        setNewContexts((current) =>
                          current.map((item, itemIndex) =>
                            itemIndex === index
                              ? { ...item, answer: event.target.value }
                              : item,
                          ),
                        )
                      }
                    />
                    {newContexts.length > 1 && (
                      <button
                        className="justify-self-start text-xs text-red-700"
                        onClick={() =>
                          setNewContexts((current) =>
                            current.filter(
                              (_, itemIndex) => itemIndex !== index,
                            ),
                          )
                        }
                        type="button"
                      >
                        Remove
                      </button>
                    )}
                  </div>
                ))}
              </div>
              <button
                className="mt-3 rounded border px-3 py-2 text-sm"
                onClick={() =>
                  setNewContexts((current) => [
                    ...current,
                    { question: "", answer: "" },
                  ])
                }
                type="button"
              >
                Add context question
              </button>
            </fieldset>
            <select
              className="rounded border px-3 py-2"
              value={type}
              onChange={(event) => setType(event.target.value)}
            >
              <option value="scholarship_research">Scholarship research</option>
              <option value="scholarship_discovery">
                Scholarship discovery
              </option>
              <option value="programme_research">Programme research</option>
              <option value="funding_research">Funding research</option>
              <option value="general_research">General research</option>
            </select>
            <select
              className="rounded border px-3 py-2"
              aria-label="Research priority"
              value={taskPriority}
              onChange={(event) => setTaskPriority(event.target.value)}
            >
              <option value="low">Low</option>
              <option value="normal">Normal</option>
              <option value="high">High</option>
              <option value="urgent">Urgent</option>
            </select>
            <div className="flex gap-3 md:col-span-2">
              <RequestButton
                pending={create.isPending && !create.variables}
                disabled={create.isPending}
                type="submit"
                value="draft"
                className="rounded border px-4 py-2"
              >
                Save Draft
              </RequestButton>
              <RequestButton
                pending={create.isPending && create.variables}
                disabled={create.isPending}
                type="submit"
                value="queue"
                className="rounded bg-indigo-700 px-4 py-2 text-white"
              >
                Save &amp; Queue
              </RequestButton>
            </div>
            {create.isError && (
              <p className="text-red-700 md:col-span-2">
                {create.error.message}
              </p>
            )}
          </fieldset>
        </form>
      </Dialog>
      <section className="mt-6 rounded-lg border bg-white p-5">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold">Import / Export</h2>
            <p className="text-sm text-slate-600">
              Bulk move research task definitions as JSON. Research results are
              not included.
            </p>
          </div>
          <RequestButton
            pending={exportTasks.isPending}
            className="rounded border px-3 py-2 text-sm"
            onClick={() => exportTasks.mutate()}
            type="button"
          >
            Export filtered tasks
          </RequestButton>
        </div>
        {exportTasks.isError && (
          <p role="alert" className="mt-3 text-sm text-red-700">
            {exportTasks.error.message}
          </p>
        )}
        <textarea
          className="mt-4 min-h-36 w-full rounded border px-3 py-2 font-mono text-sm"
          placeholder='{"tasks":[{"title":"Example Scholarship 2027","taskType":"scholarship_research","profileId":"...","links":[{"url":"https://example.edu"}],"researchContexts":[{"question":"Course interests","answer":"Biomedical imaging"}]}]}'
          value={importText}
          onChange={(event) => setImportText(event.target.value)}
        />
        <div className="mt-3 flex flex-wrap items-center gap-3">
          <label className="flex items-center gap-2 text-sm">
            <input
              checked={importQueue}
              onChange={(event) => setImportQueue(event.target.checked)}
              type="checkbox"
            />
            Queue imported tasks
          </label>
          <RequestButton
            pending={importTasks.isPending}
            className="rounded bg-slate-900 px-4 py-2 text-sm text-white disabled:opacity-50"
            disabled={!importText.trim() || importTasks.isPending}
            onClick={() => importTasks.mutate()}
            type="button"
          >
            Import tasks
          </RequestButton>
          {importMessage && (
            <span className="text-sm text-slate-600">{importMessage}</span>
          )}
        </div>
      </section>
      <div className="mt-6 flex flex-wrap gap-2">
        <select
          className="rounded border px-3 py-2"
          aria-label="Filter by status"
          value={status}
          onChange={(event) => {
            setFilter("status", event.target.value);
          }}
        >
          <option value="">All statuses</option>
          <option value="draft">Draft</option>
          <option value="queued">Queued</option>
          <option value="running">Running</option>
          <option value="cancelled">Cancelled</option>
          <option value="review_required">Review required</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
        </select>
        <input
          className="rounded border px-3 py-2"
          placeholder="Task type filter"
          value={taskType}
          onChange={(event) => {
            setFilter("taskType", event.target.value);
          }}
        />
        <select
          className="rounded border px-3 py-2"
          aria-label="Filter by priority"
          value={priority}
          onChange={(event) => {
            setFilter("priority", event.target.value);
          }}
        >
          <option value="">All priorities</option>
          <option value="low">Low</option>
          <option value="normal">Normal</option>
          <option value="high">High</option>
          <option value="urgent">Urgent</option>
        </select>
        <select
          className="rounded border px-3 py-2"
          aria-label="Filter by proposals"
          value={proposalState}
          onChange={(event) => setFilter("proposalState", event.target.value)}
        >
          <option value="">All proposal states</option>
          <option value="any">Has proposals</option>
          <option value="none">No proposals</option>
          <option value="approved">Has approved proposals</option>
        </select>
        <label className="flex items-center gap-2 text-sm">
          Sort by
          <select
            className="rounded border px-3 py-2"
            value={sortBy}
            onChange={(event) => {
              setFilter("sortBy", event.target.value);
            }}
          >
            <option value="priority">Priority</option>
            <option value="createdAt">Newest</option>
          </select>
        </label>
      </div>
      {tasks.isPending && (
        <p className="mt-4" role="status">
          Loading research tasks…
        </p>
      )}
      {tasks.isError && (
        <p className="mt-4 text-red-700" role="alert">
          {tasks.error.message}
        </p>
      )}
      {tasks.data?.data.length === 0 && (
        <p className="mt-4 text-slate-600">
          No research tasks match these filters.
        </p>
      )}
      <div className="mt-4 grid gap-3">
        {tasks.data?.data.map((task) => (
          <article key={task.id} className="rounded-lg border bg-white p-4">
            <Link
              to={`/research/${task.id}?${searchParams.toString()}`}
              className="block"
            >
              <div className="flex justify-between">
                <span className="font-medium">{task.title}</span>
                <span className="text-xs uppercase text-indigo-700">
                  {task.status.replaceAll("_", " ")}
                </span>
              </div>
              <p className="mt-2 text-sm text-slate-600">
                {task.taskType.replaceAll("_", " ")} ·{" "}
                {task.priority ?? "normal"} · Created{" "}
                {new Date(task.createdAt).toLocaleDateString()}
              </p>
            </Link>
            <ScholarshipDeadlines task={task} />
          </article>
        ))}
      </div>
      <nav
        aria-label="Research task pagination"
        className="mt-5 flex flex-wrap items-center justify-between gap-3 text-sm"
      >
        <label className="flex items-center gap-2">
          Tasks per page
          <select
            className="rounded border px-2 py-1"
            value={pageSize}
            onChange={(event) => {
              setFilter("pageSize", event.target.value);
            }}
          >
            {[10, 20, 50, 100].map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
        </label>
        <span>
          {tasks.data
            ? `${tasks.data.meta.total} task(s) · Page ${page} of ${Math.max(1, tasks.data.meta.totalPages)}`
            : "Loading…"}
        </span>
        <div className="flex gap-2">
          <RequestButton
            type="button"
            className="rounded border px-3 py-2 disabled:opacity-50"
            disabled={page === 1 || tasks.isFetching}
            pending={tasks.isFetching && pageAction === "previous"}
            onClick={() => {
              setPageAction("previous");
              setPage(page - 1);
            }}
          >
            Previous
          </RequestButton>
          <RequestButton
            type="button"
            className="rounded border px-3 py-2 disabled:opacity-50"
            disabled={
              !tasks.data ||
              page >= tasks.data.meta.totalPages ||
              tasks.isFetching
            }
            pending={tasks.isFetching && pageAction === "next"}
            onClick={() => {
              setPageAction("next");
              setPage(page + 1);
            }}
          >
            Next
          </RequestButton>
        </div>
      </nav>
    </section>
  );
}
