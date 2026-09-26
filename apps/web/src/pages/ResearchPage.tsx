import { useMutation, useQueryClient } from "@tanstack/react-query";
import { FormEvent, useState } from "react";
import { Link } from "react-router-dom";
import { createResearchTask, workflowKeys } from "../features/workflow/api";
import {
  useResearchContexts,
  useResearchTasks,
} from "../features/workflow/queries";
import { useAuth } from "../features/auth/AuthProvider";
import { useProfiles } from "../features/profile/queries";
import { CollapsibleText } from "../components/CollapsibleText";

export function ResearchPage() {
  const { user } = useAuth();
  const userId = user?.id ?? "";
  const [status, setStatus] = useState("");
  const [taskType, setTaskType] = useState("");
  const [priority, setPriority] = useState("");
  const query = new URLSearchParams(
    Object.fromEntries(
      Object.entries({ userId, status, taskType, priority }).filter(
        ([, value]) => value,
      ),
    ),
  ).toString();
  const tasks = useResearchTasks(query ? `?${query}` : "");
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
  function submit(event: FormEvent) {
    event.preventDefault();
    const submitter = (event.nativeEvent as SubmitEvent).submitter as HTMLButtonElement | null;
    create.mutate(submitter?.value === "queue");
  }
  return (
    <section>
      <h1 className="text-3xl font-semibold">Research Inbox</h1>
      <p className="mt-2 text-slate-600">
        Create controlled research work. Results become proposals and never
        Applications until you approve them.
      </p>
      <form
        onSubmit={submit}
        className="mt-6 grid gap-3 rounded-lg border bg-white p-5 md:grid-cols-2"
      >
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
                <label className="flex gap-2 rounded bg-slate-50 p-3 text-sm" key={context.id}>
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
              <div className="grid gap-2 rounded bg-slate-50 p-3" key={index}>
                <input
                  className="rounded border px-3 py-2"
                  placeholder="Question, e.g. What courses interest you?"
                  required={Boolean(context.question.trim() || context.answer.trim())}
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
                  required={Boolean(context.question.trim() || context.answer.trim())}
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
                        current.filter((_, itemIndex) => itemIndex !== index),
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
          <option value="scholarship_discovery">Scholarship discovery</option>
          <option value="programme_research">Programme research</option>
          <option value="funding_research">Funding research</option>
          <option value="general_research">General research</option>
        </select>
        <select
          className="rounded border px-3 py-2"
          value={taskPriority}
          onChange={(event) => setTaskPriority(event.target.value)}
        >
          <option value="low">Low</option>
          <option value="normal">Normal</option>
          <option value="high">High</option>
          <option value="urgent">Urgent</option>
        </select>
        <div className="flex gap-3 md:col-span-2">
          <button
            type="submit"
            value="draft"
            className="rounded border px-4 py-2"
          >
            Save Draft
          </button>
          <button
            type="submit"
            value="queue"
            className="rounded bg-indigo-700 px-4 py-2 text-white"
          >
            Save &amp; Queue
          </button>
        </div>
        {create.isError && (
          <p className="text-red-700 md:col-span-2">{create.error.message}</p>
        )}
      </form>
      <div className="mt-6 flex flex-wrap gap-2">
        <select
          className="rounded border px-3 py-2"
          value={status}
          onChange={(event) => setStatus(event.target.value)}
        >
          <option value="">All statuses</option>
          <option value="draft">Draft</option>
          <option value="queued">Queued</option>
          <option value="review_required">Review required</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
        </select>
        <input
          className="rounded border px-3 py-2"
          placeholder="Task type filter"
          value={taskType}
          onChange={(event) => setTaskType(event.target.value)}
        />
        <select
          className="rounded border px-3 py-2"
          value={priority}
          onChange={(event) => setPriority(event.target.value)}
        >
          <option value="">All priorities</option>
          <option value="normal">Normal</option>
          <option value="high">High</option>
          <option value="urgent">Urgent</option>
        </select>
      </div>
      <div className="mt-4 grid gap-3">
        {tasks.data?.data.map((task) => (
          <Link
            key={task.id}
            to={`/research/${task.id}`}
            className="rounded-lg border bg-white p-4"
          >
            <div className="flex justify-between">
              <span className="font-medium">{task.title}</span>
              <span className="text-xs uppercase text-indigo-700">
                {task.status.replaceAll("_", " ")}
              </span>
            </div>
            <p className="mt-2 text-sm text-slate-600">
              {task.taskType.replaceAll("_", " ")} · {task.priority ?? "normal"}{" "}
              · {new Date(task.createdAt).toLocaleDateString()}
            </p>
          </Link>
        ))}
      </div>
    </section>
  );
}
