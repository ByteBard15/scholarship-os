import { Link } from "react-router-dom";
import { useProfiles } from "../features/profile/queries";

export function ProfilesPage() {
  const userId = import.meta.env.VITE_DEMO_USER_ID ?? "";
  const query = useProfiles(userId);
  if (!userId)
    return (
      <section>
        <h1 className="text-2xl font-semibold">Profiles</h1>
        <p className="mt-3 text-slate-600">
          Set <code>VITE_DEMO_USER_ID</code> in <code>.env</code> to load
          profiles while authentication is not yet implemented.
        </p>
      </section>
    );
  if (query.isPending) return <p>Loading profiles…</p>;
  if (query.isError)
    return <p className="text-red-700">{query.error.message}</p>;
  return (
    <section>
      <h1 className="text-2xl font-semibold">Profiles</h1>
      <div className="mt-6 grid gap-3">
        {query.data.data.map((profile) => (
          <Link
            className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm"
            key={profile.id}
            to={`/profiles/${profile.id}`}
          >
            <span className="font-medium">{profile.name}</span>
            {profile.isDefault && (
              <span className="ml-2 rounded bg-indigo-50 px-2 py-1 text-xs text-indigo-700">
                Default
              </span>
            )}
            <p className="mt-1 text-sm text-slate-600">
              {profile.headline ?? "No headline yet"}
            </p>
          </Link>
        ))}
      </div>
      {query.data.meta.count === 0 && (
        <p className="mt-4 text-slate-600">No profiles yet.</p>
      )}
    </section>
  );
}
