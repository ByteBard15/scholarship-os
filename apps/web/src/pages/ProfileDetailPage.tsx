import { useParams } from "react-router-dom";
import { useProfile } from "../features/profile/queries";

export function ProfileDetailPage() {
  const { id = "" } = useParams();
  const query = useProfile(id);
  if (query.isPending) return <p>Loading profile…</p>;
  if (query.isError)
    return <p className="text-red-700">{query.error.message}</p>;
  const profile = query.data.data;
  const sectionCount =
    profile.education.length +
    profile.employment.length +
    profile.projects.length +
    profile.publications.length +
    profile.articles.length +
    profile.skills.length +
    profile.researchInterests.length +
    profile.careerGoals.length +
    profile.certifications.length +
    profile.awards.length +
    profile.volunteering.length;
  return (
    <section>
      <p className="text-sm text-slate-500">Profile overview</p>
      <h1 className="mt-1 text-3xl font-semibold">{profile.name}</h1>
      <p className="mt-3 text-slate-600">
        {profile.headline ?? "No headline yet"}
      </p>
      <div className="mt-8 rounded-lg border border-slate-200 bg-white p-5">
        <p className="text-sm text-slate-500">Saved section entries</p>
        <p className="mt-1 text-2xl font-semibold">{sectionCount}</p>
      </div>
    </section>
  );
}
