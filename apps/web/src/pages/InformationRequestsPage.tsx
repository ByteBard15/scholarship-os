import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import {
  processInformationRequest,
  respondInformationRequest,
  transitionInformationRequest,
  workflowKeys,
} from "../features/workflow/api";
import { useInformationRequests } from "../features/workflow/queries";
import type { InformationRequest } from "../features/workflow/types";
import { useAuth } from "../features/auth/AuthProvider";

export function InformationRequestsPage() {
  const { user } = useAuth();
  const userId = user?.id ?? "";
  const [status, setStatus] = useState("");
  const query = new URLSearchParams(
    Object.fromEntries(
      Object.entries({ userId, status }).filter(([, value]) => value),
    ),
  ).toString();
  const requests = useInformationRequests(query ? `?${query}` : "");
  const client = useQueryClient(),
    refresh = () => client.invalidateQueries({ queryKey: workflowKeys.all });
  const respond = useMutation({
    mutationFn: ({ id, value }: { id: string; value: string }) =>
      respondInformationRequest(id, value),
    onSuccess: () => void refresh(),
  });
  const process = useMutation({
    mutationFn: processInformationRequest,
    onSuccess: () => void refresh(),
  });
  const reopen = useMutation({
    mutationFn: (id: string) => transitionInformationRequest(id, "reopen"),
    onSuccess: () => void refresh(),
  });
  return (
    <section>
      <h1 className="text-3xl font-semibold">Needs Your Input</h1>
      <p className="mt-2 text-slate-600">
        The agent asks instead of inventing unsupported personal facts.
      </p>
      <select
        className="mt-5 rounded border px-3 py-2"
        value={status}
        onChange={(event) => setStatus(event.target.value)}
      >
        <option value="">All statuses</option>
        <option value="pending">Pending</option>
        <option value="answered">Answered</option>
        <option value="completed">Completed</option>
        <option value="reopened">Reopened</option>
      </select>
      <div className="mt-4 grid gap-4">
        {requests.data?.data.map((request) => (
          <RequestCard
            key={request.id}
            request={request}
            onRespond={(value) => respond.mutate({ id: request.id, value })}
            onProcess={() => process.mutate(request.id)}
            onReopen={() => reopen.mutate(request.id)}
          />
        ))}
      </div>
    </section>
  );
}
function RequestCard({
  request,
  onRespond,
  onProcess,
  onReopen,
}: {
  request: InformationRequest;
  onRespond(value: string): void;
  onProcess(): void;
  onReopen(): void;
}) {
  const [value, setValue] = useState("");
  return (
    <article className="rounded-lg border bg-white p-5">
      <div className="flex justify-between">
        <h2 className="font-medium">{request.title}</h2>
        <span className="text-xs uppercase text-indigo-700">
          {request.status}
        </span>
      </div>
      <p className="mt-2">{request.prompt}</p>
      {request.context && (
        <p className="mt-2 text-sm text-slate-500">Why: {request.context}</p>
      )}
      {["pending", "reopened"].includes(request.status) && (
        <div className="mt-4 flex gap-2">
          <input
            className="flex-1 rounded border px-3 py-2"
            type={request.responseType === "date" ? "date" : "text"}
            value={value}
            onChange={(event) => setValue(event.target.value)}
          />
          <button
            className="rounded bg-indigo-700 px-4 py-2 text-white"
            onClick={() => onRespond(value)}
            disabled={!value}
          >
            Submit Response
          </button>
        </div>
      )}
      {request.status === "answered" && (
        <button
          className="mt-4 rounded bg-indigo-700 px-4 py-2 text-white"
          onClick={onProcess}
        >
          Process Response
        </button>
      )}
      {request.status === "completed" && (
        <div className="mt-4">
          <p className="text-sm text-emerald-700">
            COMPLETED — response applied with provenance.
          </p>
          <button
            className="mt-2 rounded border px-3 py-1.5 text-sm"
            onClick={onReopen}
          >
            Reopen
          </button>
        </div>
      )}
      {request.responses.length > 0 && (
        <details className="mt-4 text-sm">
          <summary>Previous responses ({request.responses.length})</summary>
          {request.responses.map((response) => (
            <p key={response.id} className="mt-2 border-l-2 pl-3">
              {response.responseText ?? JSON.stringify(response.responseValue)}{" "}
              · {new Date(response.createdAt).toLocaleString()}
            </p>
          ))}
        </details>
      )}
    </article>
  );
}
