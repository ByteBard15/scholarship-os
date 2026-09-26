import { getJSON, requestJSON } from "../../api/client";
import type {
  Application,
  ApplicationDetail,
  ApplicationTask,
  CollectionEnvelope,
  Dashboard,
  DataEnvelope,
  Deadline,
  Institution,
  Programme,
  ResearchFinding,
  ResearchRun,
  ResearchSource,
  Scholarship,
} from "./types";

export const applicationKeys = {
  all: ["applications"] as const,
  list: (filters: string) => ["applications", "list", filters] as const,
  detail: (id: string) => ["applications", "detail", id] as const,
  research: (id: string) => ["applications", id, "research"] as const,
  findings: (id: string, run: string) =>
    ["applications", id, "research", run, "findings"] as const,
  dashboard: ["dashboard"] as const,
};
export const listApplications = (query = "") =>
  getJSON<CollectionEnvelope<Application>>(`/api/v1/applications${query}`);
export const getApplication = (id: string) =>
  getJSON<DataEnvelope<ApplicationDetail>>(`/api/v1/applications/${id}`);
export const createApplication = (input: {
  name: string;
  parentProfileId: string;
  institutionId?: string;
  programmeId?: string;
  scholarshipId?: string;
  intake?: string;
  intakeYear?: number;
  country?: string;
}) =>
  requestJSON<DataEnvelope<Application>>("/api/v1/applications", {
    method: "POST",
    body: JSON.stringify(input),
  });
export const listInstitutions = () =>
  getJSON<CollectionEnvelope<Institution>>("/api/v1/institutions");
export const listProgrammes = (institutionId?: string) =>
  getJSON<CollectionEnvelope<Programme>>(
    `/api/v1/programmes${institutionId ? `?institutionId=${institutionId}` : ""}`,
  );
export const listScholarships = () =>
  getJSON<CollectionEnvelope<Scholarship>>("/api/v1/scholarships");
export const runResearch = (id: string) =>
  requestJSON<DataEnvelope<ResearchRun>>(
    `/api/v1/applications/${id}/research`,
    {
      method: "POST",
      body: JSON.stringify({ trigger: "manual", researchType: "full" }),
    },
  );
export const listResearch = (id: string) =>
  getJSON<CollectionEnvelope<ResearchRun>>(
    `/api/v1/applications/${id}/research`,
  );
export const listSources = (id: string, run: string) =>
  getJSON<CollectionEnvelope<ResearchSource>>(
    `/api/v1/applications/${id}/research/${run}/sources`,
  );
export const listFindings = (id: string, run: string) =>
  getJSON<CollectionEnvelope<ResearchFinding>>(
    `/api/v1/applications/${id}/research/${run}/findings`,
  );
export const reviewFinding = (
  id: string,
  run: string,
  finding: string,
  status: "accepted" | "rejected",
) =>
  requestJSON<DataEnvelope<ResearchFinding>>(
    `/api/v1/applications/${id}/research/${run}/findings/${finding}`,
    { method: "PATCH", body: JSON.stringify({ reviewStatus: status }) },
  );
export const applyResearch = (id: string, run: string) =>
  requestJSON<DataEnvelope<{ status: string }>>(
    `/api/v1/applications/${id}/research/${run}/apply`,
    { method: "POST" },
  );
export const completeTask = (id: string, task: string, complete: boolean) =>
  requestJSON<DataEnvelope<ApplicationTask>>(
    `/api/v1/applications/${id}/tasks/${task}/${complete ? "complete" : "reopen"}`,
    { method: "POST" },
  );
export const upcomingDeadlines = (days = 30) =>
  getJSON<CollectionEnvelope<Deadline>>(
    `/api/v1/deadlines/upcoming?days=${days}`,
  );
export const getDashboard = () =>
  getJSON<DataEnvelope<Dashboard>>("/api/v1/dashboard");
