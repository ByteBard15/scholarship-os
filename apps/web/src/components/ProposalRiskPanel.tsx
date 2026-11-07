import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getJSON, requestJSON } from "../api/client";
import { Dialog } from "./Dialog";
import { RequestButton } from "./RequestButton";
import { TaskGuide } from "./TaskGuide";

type Factor = {
  id: string;
  variableName: string;
  value: number | null;
  rationale: string;
  sourceUrl?: string;
  observedAt: string;
  timeVariable: boolean;
  timePredicate?: string;
  validUntil?: string;
  blocking: boolean;
  stale: boolean;
};
type Assessment = {
  id: string;
  riskFactor: number | null;
  computedAt: string;
  stale: boolean;
  unknownCount: number;
  blockingCount: number;
  profileSnapshotId: string;
  notes: string;
  variables: Factor[];
};
type Input = {
  variableName: string;
  value: string;
  rationale: string;
  sourceUrl: string;
  observedAt: string;
  timeVariable: boolean;
  timePredicate: string;
  validUntil: string;
  blocking: boolean;
};
function localNow() {
  return new Date(Date.now() - new Date().getTimezoneOffset() * 60000)
    .toISOString()
    .slice(0, 16);
}
const names = [
  "academic_eligibility",
  "research_fit",
  "procedure_complexity",
  "document_readiness",
  "funding_feasibility",
  "supervisor_accessibility",
];
function newFactor(variableName: string): Input {
  return {
    variableName,
    value: "",
    rationale: "",
    sourceUrl: "",
    observedAt: localNow(),
    timeVariable: false,
    timePredicate: "",
    validUntil: "",
    blocking: false,
  };
}

export function ProposalRiskPanel({
  proposalId,
  profiles,
  defaultProfileId,
}: {
  proposalId: string;
  profiles: { id: string; name: string }[];
  defaultProfileId?: string;
}) {
  const client = useQueryClient();
  const key = ["proposal-risk", proposalId];
  const path = `/api/v1/application-proposals/${proposalId}/risk-assessments`;
  const history = useQuery({
    queryKey: key,
    queryFn: () => getJSON<{ data: Assessment[] }>(path),
  });
  const [open, setOpen] = useState(false);
  const [profileId, setProfileId] = useState("");
  const [notes, setNotes] = useState("");
  const [factors, setFactors] = useState<Input[]>([]);
  const change = (index: number, patch: Partial<Input>) =>
    setFactors((values) =>
      values.map((item, i) => (i === index ? { ...item, ...patch } : item)),
    );
  const save = useMutation({
    mutationFn: () =>
      requestJSON(path, {
        method: "POST",
        body: JSON.stringify({
          profileId,
          notes,
          variables: factors.map((f) => ({
            ...f,
            value: f.value === "" ? null : Number(f.value),
            sourceUrl: f.sourceUrl || undefined,
            observedAt: new Date(f.observedAt).toISOString(),
            timePredicate: f.timePredicate || undefined,
            validUntil: f.validUntil
              ? new Date(f.validUntil).toISOString()
              : undefined,
          })),
        }),
      }),
    onSuccess: () => {
      void client.invalidateQueries({ queryKey: key });
      setOpen(false);
    },
  });
  const latest = history.data?.data?.[0];
  return (
    <section className="mt-4 rounded bg-slate-50 p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h4 className="font-semibold">Risk factor</h4>
        <button
          type="button"
          className="text-sm text-indigo-700"
          onClick={() => {
            setProfileId(defaultProfileId || profiles[0]?.id || "");
            setNotes("");
            setFactors(names.map(newFactor));
            save.reset();
            setOpen(true);
          }}
        >
          New assessment
        </button>
      </div>
      <p className="mt-1 text-xs text-slate-600">
        0 = lower risk; 100 = higher risk. Equal-weight assessment index, not an
        admission probability. Unknowns and hard blockers are shown separately.
      </p>
      {history.isPending && <p className="mt-2 text-sm">Loading assessment…</p>}
      {history.isError && (
        <p role="alert" className="mt-2 text-sm text-red-700">
          {history.error.message}
        </p>
      )}
      {history.data && !latest && (
        <p className="mt-2 text-sm text-slate-500">Not assessed.</p>
      )}
      {latest && (
        <>
          <p className="mt-2 text-sm font-medium">
            {latest.riskFactor == null
              ? "Unscored"
              : `${latest.riskFactor.toFixed(1)}/100`}{" "}
            · Computed {new Date(latest.computedAt).toLocaleString()}
            {latest.stale ? " · STALE — refresh evidence" : ""}
          </p>
          <p className="text-xs text-amber-800">
            {latest.unknownCount} unknown factor(s) · {latest.blockingCount}{" "}
            hard blocker(s). Later credential or evidence changes require a new
            assessment.
          </p>
          <details className="mt-2 text-sm">
            <summary className="cursor-pointer text-indigo-700">
              Factors and assessment history
            </summary>
            {(history.data?.data ?? []).map((assessment) => (
              <article key={assessment.id} className="mt-3 border-t pt-3">
                <p className="font-medium">
                  {new Date(assessment.computedAt).toLocaleString()} ·{" "}
                  {assessment.riskFactor == null
                    ? "Unscored"
                    : `${assessment.riskFactor.toFixed(1)}/100`}
                  {assessment.stale ? " · Stale" : ""}
                </p>
                <p className="mt-1 whitespace-pre-wrap">{assessment.notes}</p>
                <p className="text-xs text-slate-500">
                  Credential snapshot: {assessment.profileSnapshotId}
                </p>
                {assessment.variables.map((factor) => (
                  <div key={factor.id} className="mt-3 border-l-2 pl-3">
                    <p className="font-medium">
                      {factor.variableName}: {factor.value ?? "Unknown"}
                      {factor.blocking ? " · HARD BLOCKER" : ""}
                      {factor.stale ? " · Stale" : ""}
                    </p>
                    <p className="whitespace-pre-wrap">{factor.rationale}</p>
                    <p className="text-xs text-slate-500">
                      Observed {new Date(factor.observedAt).toLocaleString()}
                      {factor.validUntil
                        ? ` · Review by ${new Date(factor.validUntil).toLocaleString()}`
                        : ""}
                    </p>
                    {factor.timePredicate && (
                      <p className="text-xs">
                        Time condition: {factor.timePredicate}
                      </p>
                    )}
                    {factor.sourceUrl && (
                      <TaskGuide description={`Source: ${factor.sourceUrl}`} />
                    )}
                  </div>
                ))}
              </article>
            ))}
            <p className="mt-2 text-xs text-slate-500">
              Up to 50 most recent assessments shown. Historical values remain
              unchanged.
            </p>
          </details>
        </>
      )}
      <Dialog
        open={open}
        title="New proposal risk assessment"
        busy={save.isPending}
        onClose={() => setOpen(false)}
      >
        <form
          className="grid gap-4 p-5"
          onSubmit={(event) => {
            event.preventDefault();
            save.mutate();
          }}
        >
          <p className="text-sm text-slate-600">
            Record evidence-based risk values. Leave unknown values blank and
            explain what is missing. Every assessment creates an immutable
            credential snapshot.
          </p>
          <fieldset disabled={save.isPending} className="grid gap-4">
            <label className="grid gap-1 text-sm">
              Credentials to assess
              <select
                required
                className="rounded border p-2"
                value={profileId}
                onChange={(e) => setProfileId(e.target.value)}
              >
                <option value="">Select profile</option>
                {profiles.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="grid gap-1 text-sm">
              Assessment notes and assumptions
              <textarea
                required
                maxLength={10000}
                className="rounded border p-2"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />
            </label>
            {factors.map((factor, i) => (
              <fieldset key={i} className="grid gap-2 rounded border p-3">
                <legend className="px-1 text-sm font-medium">
                  Factor {i + 1}
                </legend>
                <label className="grid gap-1 text-sm">
                  Variable name
                  <input
                    required
                    pattern="[a-z][a-z0-9_]{0,99}"
                    className="rounded border p-2"
                    value={factor.variableName}
                    onChange={(e) =>
                      change(i, { variableName: e.target.value })
                    }
                  />
                </label>
                <label className="grid gap-1 text-sm">
                  Risk value (0–100; blank = unknown)
                  <input
                    type="number"
                    min={0}
                    max={100}
                    step="any"
                    className="rounded border p-2"
                    value={factor.value}
                    onChange={(e) => change(i, { value: e.target.value })}
                  />
                </label>
                <label className="grid gap-1 text-sm">
                  Evidence and rationale
                  <textarea
                    required
                    maxLength={4000}
                    className="rounded border p-2"
                    value={factor.rationale}
                    onChange={(e) => change(i, { rationale: e.target.value })}
                  />
                </label>
                <label className="grid gap-1 text-sm">
                  Source URL (optional)
                  <input
                    type="url"
                    className="rounded border p-2"
                    value={factor.sourceUrl}
                    onChange={(e) => change(i, { sourceUrl: e.target.value })}
                  />
                </label>
                <label className="grid gap-1 text-sm">
                  Evidence observed at
                  <input
                    required
                    type="datetime-local"
                    className="rounded border p-2"
                    value={factor.observedAt}
                    onChange={(e) => change(i, { observedAt: e.target.value })}
                  />
                </label>
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={factor.blocking}
                    onChange={(e) => change(i, { blocking: e.target.checked })}
                  />
                  Hard requirement blocker
                </label>
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={factor.timeVariable}
                    onChange={(e) =>
                      change(i, { timeVariable: e.target.checked })
                    }
                  />
                  Time-variable evidence
                </label>
                {factor.timeVariable && (
                  <>
                    <label className="grid gap-1 text-sm">
                      Valid only while / temporal condition
                      <textarea
                        required
                        maxLength={2000}
                        className="rounded border p-2"
                        value={factor.timePredicate}
                        onChange={(e) =>
                          change(i, { timePredicate: e.target.value })
                        }
                      />
                    </label>
                    <label className="grid gap-1 text-sm">
                      Review by
                      <input
                        required
                        type="datetime-local"
                        className="rounded border p-2"
                        value={factor.validUntil}
                        onChange={(e) =>
                          change(i, { validUntil: e.target.value })
                        }
                      />
                    </label>
                  </>
                )}
                <button
                  type="button"
                  className="justify-self-start text-sm text-red-700"
                  disabled={factors.length <= 1}
                  onClick={() =>
                    setFactors((v) => v.filter((_, index) => index !== i))
                  }
                >
                  Remove factor
                </button>
              </fieldset>
            ))}
            <button
              type="button"
              className="justify-self-start text-sm text-indigo-700"
              disabled={factors.length >= 30}
              onClick={() => setFactors((v) => [...v, newFactor("")])}
            >
              Add factor
            </button>
          </fieldset>
          {save.isError && (
            <p role="alert" className="text-sm text-red-700">
              {save.error.message}
            </p>
          )}
          <RequestButton
            pending={save.isPending}
            disabled={save.isPending || !profileId}
            type="submit"
            className="rounded bg-indigo-700 px-4 py-2 text-white"
          >
            Compute and save assessment
          </RequestButton>
        </form>
      </Dialog>
    </section>
  );
}
