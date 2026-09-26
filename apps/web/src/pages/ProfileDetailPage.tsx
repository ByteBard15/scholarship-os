import { useMutation, useQueryClient } from "@tanstack/react-query";
import { FormEvent, useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  createImport,
  createOverride,
  createSnapshot,
  deleteOverride,
  profileKeys,
  uploadDocument,
} from "../features/profile/api";
import {
  useCompleteness,
  useDocuments,
  useImports,
  useOverrides,
  useProfile,
  useProfiles,
  useSnapshots,
} from "../features/profile/queries";
import type { ProfileEntry } from "../features/profile/types";
import { useAuth } from "../features/auth/AuthProvider";

const sections = [
  { key: "education", label: "Education" },
  { key: "employment", label: "Employment" },
  { key: "projects", label: "Projects" },
  { key: "publications", label: "Publications" },
  { key: "articles", label: "Articles" },
  { key: "skills", label: "Skills" },
  { key: "researchInterests", label: "Research Interests" },
  { key: "careerGoals", label: "Career Goals" },
  { key: "certifications", label: "Certifications" },
  { key: "awards", label: "Awards" },
  { key: "volunteering", label: "Volunteering" },
] as const;
function titleOf(item: ProfileEntry) {
  return String(
    item.title ??
      item.name ??
      item.institution ??
      item.organization ??
      item.role ??
      "Untitled entry",
  );
}

export function ProfileDetailPage() {
  const { user } = useAuth();
  const { id = "" } = useParams();
  const query = useProfile(id);
  const completeness = useCompleteness(id);
  const documents = useDocuments(id);
  const imports = useImports(id);
  const overrides = useOverrides(id);
  const snapshots = useSnapshots(id);
  const profiles = useProfiles(user?.id ?? "");
  const client = useQueryClient();
  const [file, setFile] = useState<File>();
  const [reason, setReason] = useState("");
  const [overrideEntity, setOverrideEntity] = useState("");
  const [overrideField, setOverrideField] = useState("description");
  const [overrideValue, setOverrideValue] = useState("");
  const [compareWith, setCompareWith] = useState("");
  const refresh = () =>
    client.invalidateQueries({ queryKey: profileKeys.detail(id) });
  const upload = useMutation({
    mutationFn: () => uploadDocument(id, file!),
    onSuccess: () => {
      setFile(undefined);
      void refresh();
    },
  });
  const startImport = useMutation({
    mutationFn: (documentId: string) => createImport(id, documentId),
    onSuccess: () => void refresh(),
  });
  const snapshot = useMutation({
    mutationFn: () => createSnapshot(id, reason || undefined),
    onSuccess: () => {
      setReason("");
      void refresh();
    },
  });
  const removeOverride = useMutation({
    mutationFn: (overrideId: string) => deleteOverride(id, overrideId),
    onSuccess: () => void refresh(),
  });
  const addOverride = useMutation({
    mutationFn: (input: {
      entityType: string;
      entityId: string;
      fieldName: string;
      overrideType: "replace" | "hide";
      value?: unknown;
    }) => createOverride(id, input),
    onSuccess: () => {
      setOverrideValue("");
      void refresh();
    },
  });
  if (query.isPending) return <p>Loading profile…</p>;
  if (query.isError)
    return <p className="text-red-700">{query.error.message}</p>;
  const profile = query.data.data;
  const allEntries = sections.flatMap((section) =>
    profile[section.key].map((item) => ({ ...item, entityType: section.key })),
  );
  function submitOverride(event: FormEvent) {
    event.preventDefault();
    const selected = allEntries.find((entry) => entry.id === overrideEntity);
    if (selected)
      addOverride.mutate({
        entityType: selected.entityType,
        entityId: selected.id,
        fieldName: overrideField,
        overrideType: "replace",
        value: overrideValue,
      });
  }
  return (
    <section>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="text-sm uppercase text-indigo-700">
            {profile.profileType} profile
          </p>
          <h1 className="mt-1 text-3xl font-semibold">{profile.name}</h1>
          <p className="mt-2 text-slate-600">
            {profile.headline ?? "No headline yet"}
          </p>
          <p className="mt-2 text-xs text-slate-500">
            Lineage: {profile.lineage.map((item) => item.name).join(" → ")}
          </p>
          <div className="mt-3 flex gap-2">
            <select
              className="rounded border px-2 py-1 text-sm"
              value={compareWith}
              onChange={(event) => setCompareWith(event.target.value)}
            >
              <option value="">Compare with…</option>
              {profiles.data?.data
                .filter((item) => item.id !== id)
                .map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name}
                  </option>
                ))}
            </select>
            {compareWith && (
              <Link
                className="rounded border px-3 py-1 text-sm"
                to={`/profiles/${id}/compare/${compareWith}`}
              >
                Compare
              </Link>
            )}
          </div>
        </div>
        <div className="rounded-lg border bg-white p-4 text-center">
          <p className="text-2xl font-semibold">
            {completeness.data?.data.score ?? "…"}%
          </p>
          <p className="text-xs text-slate-500">profile completeness</p>
        </div>
      </div>
      <nav className="mt-8 flex flex-wrap gap-3 text-sm">
        <a href="#overview">Overview</a>
        <a href="#personal">Personal</a>
        {sections.map((section) => (
          <a key={section.key} href={`#${section.key}`}>
            {section.label}
          </a>
        ))}
        <a href="#documents">Documents</a>
        <a href="#imports">Imports</a>
        <a href="#overrides">Overrides</a>
        <a href="#snapshots">Snapshots</a>
      </nav>
      <div id="personal" className="mt-8 rounded-lg border bg-white p-5">
        <h2 className="font-semibold">Personal</h2>
        <p className="mt-2 text-sm text-slate-600">
          {profile.personalInfo
            ? `${profile.personalInfo.firstName ?? ""} ${profile.personalInfo.lastName ?? ""}`
            : "No personal information"}
        </p>
        {profile.personalInfo?.inherited && (
          <p className="mt-1 text-xs text-indigo-700">
            Inherited canonical content
          </p>
        )}
      </div>
      {sections.map((section) => (
        <div
          id={section.key}
          key={section.key}
          className="mt-5 rounded-lg border bg-white p-5"
        >
          <h2 className="font-semibold">{section.label}</h2>
          <div className="mt-3 grid gap-2">
            {profile[section.key].map((item) => (
              <div key={item.id} className="rounded border p-3">
                <div className="flex justify-between gap-3">
                  <span>{titleOf(item)}</span>
                  {item.inherited && (
                    <button
                      className="text-xs text-red-700"
                      onClick={() =>
                        addOverride.mutate({
                          entityType: section.key,
                          entityId: item.id,
                          fieldName: "visibility",
                          overrideType: "hide",
                        })
                      }
                    >
                      Hide inherited item
                    </button>
                  )}
                </div>
                <p className="mt-1 text-xs text-slate-500">
                  {item.inherited
                    ? `Inherited from ${item.inheritedFromProfileId}`
                    : "Canonical/profile-specific content"}
                  {item.modifiedInProfileId && ` · Modified here`}
                </p>
              </div>
            ))}
            {profile[section.key].length === 0 && (
              <p className="text-sm text-slate-500">No entries.</p>
            )}
          </div>
        </div>
      ))}
      <div id="documents" className="mt-5 rounded-lg border bg-white p-5">
        <h2 className="font-semibold">Documents</h2>
        <div className="mt-3 flex gap-2">
          <input
            type="file"
            accept=".txt,.docx,.pdf"
            onChange={(event) => setFile(event.target.files?.[0])}
          />
          <button
            className="rounded bg-indigo-700 px-3 py-2 text-sm text-white disabled:opacity-50"
            disabled={!file || upload.isPending}
            onClick={() => upload.mutate()}
          >
            Upload CV
          </button>
        </div>
        <div className="mt-4 grid gap-2">
          {documents.data?.data.map((document) => (
            <div
              className="flex items-center justify-between rounded border p-3"
              key={document.id}
            >
              <span>
                {document.originalFilename} · {document.status}
              </span>
              <button
                className="text-sm text-indigo-700"
                onClick={() => startImport.mutate(document.id)}
              >
                Import profile data
              </button>
            </div>
          ))}
        </div>
      </div>
      <div id="imports" className="mt-5 rounded-lg border bg-white p-5">
        <h2 className="font-semibold">Imports</h2>
        <div className="mt-3 grid gap-2">
          {imports.data?.data.map((item) => (
            <Link
              className="rounded border p-3"
              key={item.id}
              to={`/profiles/${id}/imports/${item.id}`}
            >
              {item.importType.toUpperCase()} import · {item.status}
            </Link>
          ))}
        </div>
      </div>
      <div id="overrides" className="mt-5 rounded-lg border bg-white p-5">
        <h2 className="font-semibold">Overrides</h2>
        {profile.profileType !== "master" && (
          <form
            className="mt-3 grid gap-2 md:grid-cols-4"
            onSubmit={submitOverride}
          >
            <select
              className="rounded border px-2 py-2"
              required
              value={overrideEntity}
              onChange={(event) => setOverrideEntity(event.target.value)}
            >
              <option value="">Inherited entity</option>
              {allEntries
                .filter((item) => item.inherited)
                .map((item) => (
                  <option value={item.id} key={item.id}>
                    {titleOf(item)}
                  </option>
                ))}
            </select>
            <input
              className="rounded border px-2 py-2"
              value={overrideField}
              onChange={(event) => setOverrideField(event.target.value)}
              placeholder="Field name"
            />
            <input
              className="rounded border px-2 py-2"
              value={overrideValue}
              onChange={(event) => setOverrideValue(event.target.value)}
              placeholder="Replacement text"
            />
            <button className="rounded bg-indigo-700 px-3 py-2 text-white">
              Create override
            </button>
          </form>
        )}
        <div className="mt-3 grid gap-2">
          {overrides.data?.data.map((item) => (
            <div
              className="flex justify-between rounded border p-3"
              key={item.id}
            >
              <span>
                {item.overrideType} {item.entityType}.{item.fieldName}
              </span>
              <button
                className="text-sm text-red-700"
                onClick={() => removeOverride.mutate(item.id)}
              >
                Restore inherited value
              </button>
            </div>
          ))}
        </div>
      </div>
      <div id="snapshots" className="mt-5 rounded-lg border bg-white p-5">
        <h2 className="font-semibold">Snapshots</h2>
        <div className="mt-3 flex gap-2">
          <input
            className="rounded border px-3 py-2"
            value={reason}
            onChange={(event) => setReason(event.target.value)}
            placeholder="Snapshot reason"
          />
          <button
            className="rounded bg-indigo-700 px-3 py-2 text-white"
            onClick={() => snapshot.mutate()}
          >
            Create snapshot
          </button>
        </div>
        <div className="mt-3 grid gap-2">
          {snapshots.data?.data.map((item) => (
            <p className="rounded border p-3" key={item.id}>
              Version {item.version} ·{" "}
              {new Date(item.createdAt).toLocaleString()}{" "}
              {item.reason && `· ${item.reason}`}
            </p>
          ))}
        </div>
      </div>
    </section>
  );
}
