import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getJSON, requestJSON } from "../api/client";
import { RequestButton } from "./RequestButton";
import { TaskGuide } from "./TaskGuide";

type Evidence = { id: string; kind: string; label: string; value: string; sourceUrl?: string; recordedByType: string; createdAt: string };
export function TaskEvidencePanel({ applicationId, taskId }: { applicationId: string; taskId: string }) {
  const client = useQueryClient();
  const key = ["applications", applicationId, "tasks", taskId, "evidence"];
  const path = `/api/v1/applications/${applicationId}/tasks/${taskId}/evidence`;
  const evidence = useQuery({ queryKey: key, queryFn: () => getJSON<{ data: Evidence[] }>(path) });
  const [kind, setKind] = useState("note");
  const [label, setLabel] = useState("");
  const [value, setValue] = useState("");
  const [sourceUrl, setSourceUrl] = useState("");
  const save = useMutation({
    mutationFn: () => requestJSON(path, { method: "POST", body: JSON.stringify({ kind, label, value, sourceUrl: sourceUrl || undefined }) }),
    onSuccess: () => { setLabel(""); setValue(""); setSourceUrl(""); void client.invalidateQueries({ queryKey: key }); },
  });
  return <section className="mt-6 border-t pt-4">
    <h3 className="font-semibold">Completion evidence</h3>
    <p className="mt-1 text-sm text-slate-600">Record what happened: an applicant ID, document reference, portal link or confirmation. Evidence does not automatically complete the task. Keep passwords and access tokens out of these fields.</p>
    {evidence.isPending && <p className="mt-3 text-sm">Loading evidence…</p>}
    {evidence.isError && <p role="alert" className="mt-3 text-sm text-red-700">{evidence.error.message}</p>}
    {evidence.data?.data?.length === 0 && <p className="mt-3 text-sm text-slate-500">No evidence recorded yet.</p>}
    <div className="mt-3 space-y-3">{(evidence.data?.data ?? []).map((item) => <article key={item.id} className="rounded border p-3">
      <h4 className="font-medium">{item.label}</h4>
      <p className="text-xs text-slate-500">{item.kind} · Recorded by {item.recordedByType} · {new Date(item.createdAt).toLocaleString()}</p>
      <TaskGuide description={item.value} />
      {item.sourceUrl && <TaskGuide description={`Source: ${item.sourceUrl}`} />}
    </article>)}</div>
    <form className="mt-4 grid gap-3" onSubmit={(event) => { event.preventDefault(); save.mutate(); }}>
      <label className="grid gap-1 text-sm">Evidence type<select className="rounded border p-2" value={kind} onChange={(event) => setKind(event.target.value)}>{["note", "link", "identifier", "document", "confirmation"].map((type) => <option key={type} value={type}>{type}</option>)}</select></label>
      <label className="grid gap-1 text-sm">Label<input className="rounded border p-2" required maxLength={200} value={label} onChange={(event) => setLabel(event.target.value)} placeholder="For example: Calgary UCID" /></label>
      <label className="grid gap-1 text-sm">Value or details<textarea className="rounded border p-2" required maxLength={10000} value={value} onChange={(event) => setValue(event.target.value)} /></label>
      <label className="grid gap-1 text-sm">Source URL (optional)<input className="rounded border p-2" type="url" maxLength={2000} value={sourceUrl} onChange={(event) => setSourceUrl(event.target.value)} /></label>
      <RequestButton pending={save.isPending} disabled={save.isPending} type="submit" className="rounded bg-indigo-700 px-3 py-2 text-white disabled:opacity-50">Add evidence</RequestButton>
      {save.isError && <p role="alert" className="text-sm text-red-700">{save.error.message}</p>}
    </form>
  </section>;
}
