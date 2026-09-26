import { useMutation, useQueryClient } from "@tanstack/react-query";
import { FormEvent, useState } from "react";
import {
  createSectionEntry,
  deleteSectionEntry,
  profileKeys,
  updateSectionEntry,
} from "./api";
import type {
  ProfileEntry,
  ProfileSectionInput,
  ProfileSectionKey,
} from "./types";

type FieldType = "text" | "textarea" | "date" | "number" | "url" | "checkbox";
type Field = {
  key: keyof ProfileSectionInput;
  label: string;
  type?: FieldType;
  required?: boolean;
};
export type SectionDefinition = {
  responseKey:
    | "education"
    | "employment"
    | "projects"
    | "publications"
    | "articles"
    | "skills"
    | "researchInterests"
    | "careerGoals"
    | "certifications"
    | "awards"
    | "volunteering";
  endpoint: ProfileSectionKey;
  label: string;
  fields: Field[];
};

const text = (
  key: keyof ProfileSectionInput,
  label: string,
  required = false,
): Field => ({ key, label, required });
const date = (key: keyof ProfileSectionInput, label: string): Field => ({
  key,
  label,
  type: "date",
});
const area = (
  key: keyof ProfileSectionInput,
  label: string,
  required = false,
): Field => ({ key, label, type: "textarea", required });

export const profileSections: SectionDefinition[] = [
  {
    responseKey: "education",
    endpoint: "education",
    label: "Education",
    fields: [
      text("institution", "Institution", true),
      text("degree", "Degree", true),
      text("fieldOfStudy", "Field of study", true),
      date("startDate", "Start date"),
      date("endDate", "End date"),
      { key: "isCurrent", label: "Currently studying", type: "checkbox" },
      text("grade", "Grade"),
      text("gradeScale", "Grade scale"),
      text("classification", "Classification"),
      text("city", "City"),
      text("country", "Country"),
      area("description", "Description"),
    ],
  },
  {
    responseKey: "employment",
    endpoint: "employment",
    label: "Employment",
    fields: [
      text("organization", "Organization", true),
      text("jobTitle", "Job title", true),
      text("employmentType", "Employment type"),
      date("startDate", "Start date"),
      date("endDate", "End date"),
      { key: "isCurrent", label: "Current role", type: "checkbox" },
      text("city", "City"),
      text("country", "Country"),
      area("description", "Description"),
    ],
  },
  {
    responseKey: "projects",
    endpoint: "projects",
    label: "Projects",
    fields: [
      text("title", "Title", true),
      text("organization", "Organization"),
      date("startDate", "Start date"),
      date("endDate", "End date"),
      area("description", "Description"),
      { key: "url", label: "Project URL", type: "url" },
      { key: "repositoryUrl", label: "Repository URL", type: "url" },
    ],
  },
  {
    responseKey: "publications",
    endpoint: "publications",
    label: "Publications",
    fields: [
      text("title", "Title", true),
      text("publicationType", "Publication type"),
      text("publisher", "Publisher"),
      date("publicationDate", "Publication date"),
      text("doi", "DOI"),
      { key: "url", label: "URL", type: "url" },
      area("description", "Description"),
    ],
  },
  {
    responseKey: "articles",
    endpoint: "articles",
    label: "Articles",
    fields: [
      text("title", "Title", true),
      date("publicationDate", "Publication date"),
      { key: "url", label: "URL", type: "url" },
      area("description", "Description"),
    ],
  },
  {
    responseKey: "skills",
    endpoint: "skills",
    label: "Skills",
    fields: [
      text("name", "Name", true),
      text("category", "Category"),
      text("proficiency", "Proficiency"),
      { key: "yearsExperience", label: "Years of experience", type: "number" },
    ],
  },
  {
    responseKey: "researchInterests",
    endpoint: "research-interests",
    label: "Research Interests",
    fields: [
      text("name", "Name", true),
      area("description", "Description"),
      { key: "priority", label: "Priority", type: "number" },
    ],
  },
  {
    responseKey: "careerGoals",
    endpoint: "career-goals",
    label: "Career Goals",
    fields: [
      text("title", "Title", true),
      area("description", "Description", true),
      text("goalType", "Goal type"),
      { key: "priority", label: "Priority", type: "number" },
    ],
  },
  {
    responseKey: "certifications",
    endpoint: "certifications",
    label: "Certifications",
    fields: [
      text("name", "Name", true),
      text("issuer", "Issuer"),
      date("issueDate", "Issue date"),
      date("expirationDate", "Expiration date"),
      text("credentialId", "Credential ID"),
      { key: "url", label: "URL", type: "url" },
    ],
  },
  {
    responseKey: "awards",
    endpoint: "awards",
    label: "Awards",
    fields: [
      text("title", "Title", true),
      text("issuer", "Issuer"),
      date("awardDate", "Award date"),
      area("description", "Description"),
    ],
  },
  {
    responseKey: "volunteering",
    endpoint: "volunteering",
    label: "Volunteering",
    fields: [
      text("organization", "Organization", true),
      text("role", "Role", true),
      date("startDate", "Start date"),
      date("endDate", "End date"),
      { key: "isCurrent", label: "Current role", type: "checkbox" },
      area("description", "Description"),
    ],
  },
];

function titleOf(item: ProfileEntry) {
  return (
    item.title ??
    item.name ??
    item.institution ??
    item.organization ??
    item.role ??
    "Untitled entry"
  );
}

function initialValues(definition: SectionDefinition, item?: ProfileEntry) {
  return Object.fromEntries(
    definition.fields.map((field) => {
      const value = item?.[field.key as keyof ProfileEntry];
      if (field.type === "checkbox") return [field.key, Boolean(value)];
      if (field.type === "date" && typeof value === "string")
        return [field.key, value.slice(0, 10)];
      return [
        field.key,
        value === undefined || value === null ? "" : String(value),
      ];
    }),
  ) as Record<string, string | boolean>;
}

function payload(
  definition: SectionDefinition,
  values: Record<string, string | boolean>,
) {
  const result: Record<string, unknown> = {};
  for (const field of definition.fields) {
    const value = values[field.key];
    if (field.type === "checkbox") {
      result[field.key] = Boolean(value);
    } else if (field.type === "date") {
      if (value)
        result[field.key] = new Date(`${value}T00:00:00.000Z`).toISOString();
    } else if (field.type === "number") {
      if (value !== "") result[field.key] = Number(value);
    } else {
      if (value !== "" || field.required) {
        result[field.key] = String(value ?? "");
      }
    }
  }
  return result as ProfileSectionInput;
}

function EntryForm({
  profileId,
  definition,
  item,
  onDone,
}: {
  profileId: string;
  definition: SectionDefinition;
  item?: ProfileEntry;
  onDone: () => void;
}) {
  const [values, setValues] = useState(() => initialValues(definition, item));
  const mutation = useMutation({
    mutationFn: () =>
      item
        ? updateSectionEntry(
            profileId,
            definition.endpoint,
            item.id,
            payload(definition, values),
          )
        : createSectionEntry(
            profileId,
            definition.endpoint,
            payload(definition, values),
          ),
    onSuccess: onDone,
  });
  function submit(event: FormEvent) {
    event.preventDefault();
    mutation.mutate();
  }
  return (
    <form
      className="mt-3 grid gap-3 rounded border bg-slate-50 p-4 md:grid-cols-2"
      onSubmit={submit}
    >
      {definition.fields.map((field) =>
        field.type === "checkbox" ? (
          <label className="flex items-center gap-2 text-sm" key={field.key}>
            <input
              type="checkbox"
              checked={Boolean(values[field.key])}
              onChange={(event) =>
                setValues((current) => ({
                  ...current,
                  [field.key]: event.target.checked,
                }))
              }
            />
            {field.label}
          </label>
        ) : (
          <label
            className={`grid gap-1 text-sm ${field.type === "textarea" ? "md:col-span-2" : ""}`}
            key={field.key}
          >
            <span>{field.label}</span>
            {field.type === "textarea" ? (
              <textarea
                className="min-h-24 rounded border px-3 py-2"
                required={field.required}
                value={String(values[field.key] ?? "")}
                onChange={(event) =>
                  setValues((current) => ({
                    ...current,
                    [field.key]: event.target.value,
                  }))
                }
              />
            ) : (
              <input
                className="rounded border px-3 py-2"
                type={field.type ?? "text"}
                step={field.type === "number" ? "any" : undefined}
                required={field.required}
                value={String(values[field.key] ?? "")}
                onChange={(event) =>
                  setValues((current) => ({
                    ...current,
                    [field.key]: event.target.value,
                  }))
                }
              />
            )}
          </label>
        ),
      )}
      <div className="flex items-center gap-2 md:col-span-2">
        <button
          className="rounded bg-indigo-700 px-3 py-2 text-sm text-white disabled:opacity-50"
          disabled={mutation.isPending}
        >
          {mutation.isPending
            ? "Saving…"
            : item
              ? "Save changes"
              : `Add ${definition.label.toLowerCase()} entry`}
        </button>
        <button
          className="rounded border px-3 py-2 text-sm"
          type="button"
          onClick={onDone}
        >
          Cancel
        </button>
        {mutation.isError && (
          <span className="text-sm text-red-700">{mutation.error.message}</span>
        )}
      </div>
    </form>
  );
}

export function ProfileSectionEditor({
  profileId,
  definition,
  entries,
  onHideInherited,
}: {
  profileId: string;
  definition: SectionDefinition;
  entries: ProfileEntry[];
  onHideInherited: (entry: ProfileEntry) => void;
}) {
  const client = useQueryClient();
  const [editing, setEditing] = useState<ProfileEntry>();
  const [creating, setCreating] = useState(false);
  const refresh = () => {
    setCreating(false);
    setEditing(undefined);
    void client.invalidateQueries({ queryKey: profileKeys.detail(profileId) });
  };
  const remove = useMutation({
    mutationFn: (id: string) =>
      deleteSectionEntry(profileId, definition.endpoint, id),
    onSuccess: refresh,
  });
  return (
    <div
      id={definition.responseKey}
      className="mt-5 rounded-lg border bg-white p-5"
    >
      <div className="flex items-center justify-between gap-3">
        <h2 className="font-semibold">{definition.label}</h2>
        <button
          className="rounded border px-3 py-1 text-sm"
          onClick={() => {
            setEditing(undefined);
            setCreating(true);
          }}
        >
          Add entry
        </button>
      </div>
      {creating && (
        <EntryForm
          profileId={profileId}
          definition={definition}
          onDone={refresh}
        />
      )}
      <div className="mt-3 grid gap-2">
        {entries.map((item) => (
          <div className="rounded border p-3" key={item.id}>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <span>{titleOf(item)}</span>
                <p className="mt-1 text-xs text-slate-500">
                  {item.inherited
                    ? `Inherited from ${item.inheritedFromProfileId}`
                    : "Owned by this profile"}
                  {item.modifiedInProfileId && " · Modified here"}
                </p>
              </div>
              <div className="flex gap-2 text-sm">
                {item.inherited ? (
                  <button
                    className="text-red-700"
                    onClick={() => onHideInherited(item)}
                  >
                    Hide inherited item
                  </button>
                ) : (
                  <>
                    <button
                      className="text-indigo-700"
                      onClick={() => {
                        setCreating(false);
                        setEditing(item);
                      }}
                    >
                      Edit
                    </button>
                    <button
                      className="text-red-700"
                      disabled={remove.isPending}
                      onClick={() => remove.mutate(item.id)}
                    >
                      Delete
                    </button>
                  </>
                )}
              </div>
            </div>
            {editing?.id === item.id && (
              <EntryForm
                profileId={profileId}
                definition={definition}
                item={item}
                onDone={refresh}
              />
            )}
          </div>
        ))}
        {entries.length === 0 && (
          <p className="text-sm text-slate-500">No entries.</p>
        )}
      </div>
    </div>
  );
}
