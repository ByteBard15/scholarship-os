import { useMutation, useQueryClient } from "@tanstack/react-query";
import { FormEvent, useState } from "react";
import { profileKeys, updateProfile } from "./api";
import type { Profile } from "./types";

export function ProfileMetadataForm({ profile }: { profile: Profile }) {
  const client = useQueryClient();
  const [name, setName] = useState(profile.name);
  const [headline, setHeadline] = useState(profile.headline ?? "");
  const [summary, setSummary] = useState(profile.summary ?? "");
  const [isDefault, setIsDefault] = useState(profile.isDefault);
  const mutation = useMutation({
    mutationFn: () =>
      updateProfile(profile.id, { name, headline, summary, isDefault }),
    onSuccess: () =>
      client.invalidateQueries({ queryKey: profileKeys.detail(profile.id) }),
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    mutation.mutate();
  }

  return (
    <form className="mt-4 grid gap-3 md:grid-cols-2" onSubmit={submit}>
      <label className="grid gap-1 text-sm">
        <span>Name</span>
        <input
          className="rounded border px-3 py-2"
          required
          value={name}
          onChange={(event) => setName(event.target.value)}
        />
      </label>
      <label className="grid gap-1 text-sm">
        <span>Headline</span>
        <input
          className="rounded border px-3 py-2"
          value={headline}
          onChange={(event) => setHeadline(event.target.value)}
        />
      </label>
      <label className="grid gap-1 text-sm md:col-span-2">
        <span>Summary</span>
        <textarea
          className="min-h-28 rounded border px-3 py-2"
          value={summary}
          onChange={(event) => setSummary(event.target.value)}
        />
      </label>
      {profile.profileType === "master" && (
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={isDefault}
            onChange={(event) => setIsDefault(event.target.checked)}
          />
          Default master profile
        </label>
      )}
      <div className="flex items-center gap-3 md:col-span-2">
        <button
          className="rounded bg-indigo-700 px-4 py-2 text-white disabled:opacity-50"
          disabled={mutation.isPending}
        >
          {mutation.isPending ? "Saving…" : "Save profile details"}
        </button>
        {mutation.isSuccess && (
          <span className="text-sm text-emerald-700">Saved</span>
        )}
        {mutation.isError && (
          <span className="text-sm text-red-700">{mutation.error.message}</span>
        )}
      </div>
    </form>
  );
}
