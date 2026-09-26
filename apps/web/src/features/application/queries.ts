import { useQuery } from "@tanstack/react-query";
import {
  applicationKeys,
  getApplication,
  getDashboard,
  listApplications,
  listFindings,
  listInstitutions,
  listProgrammes,
  listResearch,
  listScholarships,
  listSources,
} from "./api";
export const useApplications = (query = "") =>
  useQuery({
    queryKey: applicationKeys.list(query),
    queryFn: () => listApplications(query),
  });
export const useApplication = (id: string) =>
  useQuery({
    queryKey: applicationKeys.detail(id),
    queryFn: () => getApplication(id),
    enabled: Boolean(id),
  });
export const useInstitutions = () =>
  useQuery({ queryKey: ["institutions"], queryFn: listInstitutions });
export const useProgrammes = (institution?: string) =>
  useQuery({
    queryKey: ["programmes", institution],
    queryFn: () => listProgrammes(institution),
  });
export const useScholarships = () =>
  useQuery({ queryKey: ["scholarships"], queryFn: listScholarships });
export const useResearch = (id: string) =>
  useQuery({
    queryKey: applicationKeys.research(id),
    queryFn: () => listResearch(id),
    enabled: Boolean(id),
  });
export const useResearchSources = (id: string, run: string) =>
  useQuery({
    queryKey: [...applicationKeys.research(id), run, "sources"],
    queryFn: () => listSources(id, run),
    enabled: Boolean(id && run),
  });
export const useResearchFindings = (id: string, run: string) =>
  useQuery({
    queryKey: applicationKeys.findings(id, run),
    queryFn: () => listFindings(id, run),
    enabled: Boolean(id && run),
  });
export const useDashboard = () =>
  useQuery({ queryKey: applicationKeys.dashboard, queryFn: getDashboard });
