import { useMutation, useQueryClient } from "@tanstack/react-query";
import { FormEvent, useState } from "react";
import { Link } from "react-router-dom";
import {
  applicationKeys,
  createApplication,
} from "../features/application/api";
import {
  useApplications,
  useInstitutions,
  useProgrammes,
  useScholarships,
} from "../features/application/queries";
import { useProfiles } from "../features/profile/queries";
import { useAuth } from "../features/auth/AuthProvider";

export function ApplicationsPage() {
  const { user } = useAuth();
  const userId = user?.id ?? "";
  const [status, setStatus] = useState("");
  const [country, setCountry] = useState("");
  const [year, setYear] = useState("");
  const queryString = new URLSearchParams(
    Object.fromEntries(
      Object.entries({ status, country, intakeYear: year }).filter(
        ([, v]) => v,
      ),
    ),
  ).toString();
  const applications = useApplications(queryString ? `?${queryString}` : "");
  const profiles = useProfiles(userId);
  const institutions = useInstitutions();
  const scholarships = useScholarships();
  const [name, setName] = useState("");
  const [parent, setParent] = useState("");
  const [institution, setInstitution] = useState("");
  const [programme, setProgramme] = useState("");
  const [scholarship, setScholarship] = useState("");
  const [intakeYear, setIntakeYear] = useState("");
  const programmes = useProgrammes(institution || undefined);
  const client = useQueryClient();
  const create = useMutation({
    mutationFn: () =>
      createApplication({
        name,
        parentProfileId: parent,
        institutionId: institution || undefined,
        programmeId: programme || undefined,
        scholarshipId: scholarship || undefined,
        intakeYear: intakeYear ? Number(intakeYear) : undefined,
      }),
    onSuccess: () => {
      setName("");
      setParent("");
      void client.invalidateQueries({ queryKey: applicationKeys.all });
    },
  });
  function submit(event: FormEvent) {
    event.preventDefault();
    create.mutate();
  }
  return (
    <section>
      <h1 className="text-2xl font-semibold">Applications</h1>
      <p className="mt-1 text-sm text-slate-600">
        Concrete opportunities linked to isolated application profiles.
      </p>
      <form
        onSubmit={submit}
        className="mt-6 grid gap-3 rounded-lg border bg-white p-4 md:grid-cols-3"
      >
        <input
          className="rounded border px-3 py-2"
          required
          placeholder="Application name"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <select
          className="rounded border px-3 py-2"
          required
          value={parent}
          onChange={(e) => setParent(e.target.value)}
        >
          <option value="">Parent master/domain profile</option>
          {profiles.data?.data
            .filter((p) => p.profileType !== "application")
            .map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
        </select>
        <select
          className="rounded border px-3 py-2"
          value={institution}
          onChange={(e) => {
            setInstitution(e.target.value);
            setProgramme("");
          }}
        >
          <option value="">No institution yet</option>
          {institutions.data?.data.map((v) => (
            <option key={v.id} value={v.id}>
              {v.name}
            </option>
          ))}
        </select>
        <select
          className="rounded border px-3 py-2"
          value={programme}
          onChange={(e) => setProgramme(e.target.value)}
        >
          <option value="">No programme yet</option>
          {programmes.data?.data.map((v) => (
            <option key={v.id} value={v.id}>
              {v.name}
            </option>
          ))}
        </select>
        <select
          className="rounded border px-3 py-2"
          value={scholarship}
          onChange={(e) => setScholarship(e.target.value)}
        >
          <option value="">No scholarship yet</option>
          {scholarships.data?.data.map((v) => (
            <option key={v.id} value={v.id}>
              {v.name}
            </option>
          ))}
        </select>
        <input
          className="rounded border px-3 py-2"
          type="number"
          min="2000"
          max="2200"
          placeholder="Intake year"
          value={intakeYear}
          onChange={(e) => setIntakeYear(e.target.value)}
        />
        <button
          className="rounded bg-indigo-700 px-4 py-2 text-white disabled:opacity-50"
          disabled={!name || !parent || create.isPending}
        >
          Create application
        </button>
        {create.isError && (
          <p className="text-sm text-red-700 md:col-span-3">
            {create.error.message}
          </p>
        )}
      </form>
      <div className="mt-6 flex flex-wrap gap-3 rounded border bg-white p-3 text-sm">
        <select
          className="rounded border px-2 py-1"
          value={status}
          onChange={(e) => setStatus(e.target.value)}
        >
          <option value="">All statuses</option>
          <option value="discovered">Discovered</option>
          <option value="researching">Researching</option>
          <option value="preparing">Preparing</option>
          <option value="submitted">Submitted</option>
          <option value="waiting">Waiting</option>
        </select>
        <input
          className="rounded border px-2 py-1"
          placeholder="Country"
          value={country}
          onChange={(e) => setCountry(e.target.value)}
        />
        <input
          className="w-32 rounded border px-2 py-1"
          placeholder="Intake year"
          value={year}
          onChange={(e) => setYear(e.target.value)}
        />
      </div>
      {applications.isPending ? (
        <p className="mt-6">Loading applications…</p>
      ) : applications.isError ? (
        <p className="mt-6 text-red-700">{applications.error.message}</p>
      ) : (
        <div className="mt-6 grid gap-3">
          {applications.data.data.map((v) => (
            <Link
              key={v.id}
              to={`/applications/${v.id}`}
              className="rounded-lg border bg-white p-4 shadow-sm"
            >
              <div className="flex justify-between gap-4">
                <span className="font-medium">{v.name}</span>
                <span className="text-xs uppercase text-indigo-700">
                  {v.status}
                </span>
              </div>
              <p className="mt-2 text-sm text-slate-600">
                {v.institution?.name ?? "Institution not set"} ·{" "}
                {v.country ?? "Country unknown"} ·{" "}
                {v.intakeYear ?? v.intake ?? "Intake unknown"}
              </p>
              <p className="mt-2 text-xs text-slate-500">
                Research: {v.researchStatus} · Updated{" "}
                {new Date(v.updatedAt).toLocaleDateString()}
              </p>
            </Link>
          ))}
        </div>
      )}
    </section>
  );
}
