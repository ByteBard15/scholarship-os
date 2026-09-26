import { useMutation, useQueryClient } from "@tanstack/react-query";
import { FormEvent, useState } from "react";
import { profileKeys, putPersonalInfo } from "./api";
import type { PersonalInfo } from "./types";

const fields: { key: keyof PersonalInfo; label: string; type?: string }[] = [
  { key: "firstName", label: "First name" },
  { key: "middleName", label: "Middle name" },
  { key: "lastName", label: "Last name" },
  { key: "preferredName", label: "Preferred name" },
  { key: "phone", label: "Phone", type: "tel" },
  { key: "city", label: "City" },
  { key: "stateOrRegion", label: "State or region" },
  { key: "country", label: "Country" },
  { key: "nationality", label: "Nationality" },
  { key: "linkedInUrl", label: "LinkedIn URL", type: "url" },
  { key: "githubUrl", label: "GitHub URL", type: "url" },
  { key: "websiteUrl", label: "Website URL", type: "url" },
];

export function PersonalInfoForm({
  profileId,
  personalInfo,
}: {
  profileId: string;
  personalInfo: PersonalInfo | null;
}) {
  const client = useQueryClient();
  const [values, setValues] = useState<Record<string, string>>(() =>
    Object.fromEntries(
      fields.map(({ key }) => [key, String(personalInfo?.[key] ?? "")]),
    ),
  );
  const mutation = useMutation({
    mutationFn: () =>
      putPersonalInfo(
        profileId,
        Object.fromEntries(
          Object.entries(values).filter(
            ([key, value]) =>
              value !== "" || key === "firstName" || key === "lastName",
          ),
        ) as PersonalInfo,
      ),
    onSuccess: () =>
      client.invalidateQueries({ queryKey: profileKeys.detail(profileId) }),
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    mutation.mutate();
  }

  return (
    <form className="mt-4 grid gap-3 md:grid-cols-2" onSubmit={submit}>
      {personalInfo?.inherited && (
        <p className="rounded bg-indigo-50 p-3 text-sm text-indigo-800 md:col-span-2">
          These values are inherited. Saving creates personal information for
          this profile and does not change its parent.
        </p>
      )}
      {fields.map((field) => (
        <label className="grid gap-1 text-sm" key={field.key}>
          <span>{field.label}</span>
          <input
            className="rounded border px-3 py-2"
            type={field.type ?? "text"}
            required={field.key === "firstName" || field.key === "lastName"}
            value={values[field.key] ?? ""}
            onChange={(event) =>
              setValues((current) => ({
                ...current,
                [field.key]: event.target.value,
              }))
            }
          />
        </label>
      ))}
      <div className="flex items-center gap-3 md:col-span-2">
        <button
          className="rounded bg-indigo-700 px-4 py-2 text-white disabled:opacity-50"
          disabled={mutation.isPending}
        >
          {mutation.isPending ? "Saving…" : "Save personal information"}
        </button>
        {mutation.isError && (
          <span className="text-sm text-red-700">{mutation.error.message}</span>
        )}
      </div>
    </form>
  );
}
