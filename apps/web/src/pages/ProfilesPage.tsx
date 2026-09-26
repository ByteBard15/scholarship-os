import { useMutation, useQueryClient } from "@tanstack/react-query";
import { FormEvent, useState } from "react";
import { Link } from "react-router-dom";
import { createProfile, profileKeys } from "../features/profile/api";
import { useCompleteness, useProfiles } from "../features/profile/queries";
import type { Profile, ProfileType } from "../features/profile/types";
import { useAuth } from "../features/auth/AuthProvider";

function ProfileCard({
  profile,
  parent,
}: {
  profile: Profile;
  parent?: Profile;
}) {
  const completeness = useCompleteness(profile.id);
  return (
    <Link
      className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm"
      to={`/profiles/${profile.id}`}
    >
      <div className="flex items-center justify-between">
        <span className="font-medium">{profile.name}</span>
        <span className="rounded bg-slate-100 px-2 py-1 text-xs uppercase">
          {profile.profileType}
        </span>
      </div>
      <p className="mt-2 text-sm text-slate-600">
        {profile.headline ?? "No headline yet"}
      </p>
      {parent && (
        <p className="mt-2 text-xs text-slate-500">
          Inherits from {parent.name}
        </p>
      )}
      <div className="mt-3 flex justify-between text-xs text-slate-500">
        <span>{completeness.data?.data.score ?? "…"}% complete</span>
        <span>Updated {new Date(profile.updatedAt).toLocaleDateString()}</span>
      </div>
    </Link>
  );
}

export function ProfilesPage() {
  const { user } = useAuth();
  const userId = user?.id ?? "";
  const query = useProfiles(userId);
  const client = useQueryClient();
  const [name, setName] = useState("");
  const [type, setType] = useState<ProfileType>("master");
  const [parent, setParent] = useState("");
  const create = useMutation({
    mutationFn: () =>
      createProfile(userId, {
        name,
        profileType: type,
        parentProfileId: type === "master" ? undefined : parent,
      }),
    onSuccess: () => {
      setName("");
      setParent("");
      void client.invalidateQueries({ queryKey: profileKeys.list(userId) });
    },
  });
  if (query.isPending) return <p>Loading profiles…</p>;
  if (query.isError)
    return <p className="text-red-700">{query.error.message}</p>;
  const profiles = query.data.data;
  const parents = profiles.filter((item) =>
    type === "domain"
      ? item.profileType === "master"
      : type === "application"
        ? item.profileType === "domain"
        : false,
  );
  function submit(event: FormEvent) {
    event.preventDefault();
    create.mutate();
  }
  return (
    <section>
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Profiles</h1>
          <p className="mt-1 text-sm text-slate-600">
            Master facts remain canonical; derived profiles inherit and
            customize through overrides.
          </p>
        </div>
      </div>
      <form
        onSubmit={submit}
        className="mt-6 grid gap-3 rounded-lg border bg-white p-4 md:grid-cols-4"
      >
        <input
          className="rounded border px-3 py-2"
          placeholder="Profile name"
          required
          value={name}
          onChange={(event) => setName(event.target.value)}
        />
        <select
          className="rounded border px-3 py-2"
          value={type}
          onChange={(event) => {
            setType(event.target.value as ProfileType);
            setParent("");
          }}
        >
          <option value="master">Master Profile</option>
          <option value="domain">Domain Profile</option>
          <option value="application">Application Profile</option>
        </select>
        <select
          className="rounded border px-3 py-2"
          disabled={type === "master"}
          required={type !== "master"}
          value={parent}
          onChange={(event) => setParent(event.target.value)}
        >
          <option value="">
            {type === "master" ? "No parent" : "Select parent"}
          </option>
          {parents.map((item) => (
            <option value={item.id} key={item.id}>
              {item.name}
            </option>
          ))}
        </select>
        <button
          className="rounded bg-indigo-700 px-4 py-2 text-white disabled:opacity-50"
          disabled={create.isPending || !name || (type !== "master" && !parent)}
        >
          Create profile
        </button>
        {create.isError && (
          <p className="text-sm text-red-700 md:col-span-4">
            {create.error.message}
          </p>
        )}
      </form>
      <div className="mt-6 grid gap-3">
        {profiles.map((profile) => (
          <ProfileCard
            key={profile.id}
            profile={profile}
            parent={profiles.find(
              (item) => item.id === profile.parentProfileId,
            )}
          />
        ))}
      </div>
      {profiles.length === 0 && (
        <p className="mt-4 text-slate-600">No profiles yet.</p>
      )}
    </section>
  );
}
