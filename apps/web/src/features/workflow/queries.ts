import { useQuery } from "@tanstack/react-query";
import {
  getPreparation,
  getResearchTask,
  listFields,
  listInformationRequests,
  listProposals,
  listQuestionnaires,
  listResearchTasks,
  listResearchContexts,
  listTaskContexts,
  listTaskFindings,
  listTaskLinks,
  listTaskOutputs,
  listTaskRuns,
  listTaskSources,
  workflowKeys,
} from "./api";

export const useResearchTasks = (query = "") =>
  useQuery({
    queryKey: workflowKeys.tasks(query),
    queryFn: () => listResearchTasks(query),
  });
export const useResearchTask = (id: string) =>
  useQuery({
    queryKey: workflowKeys.task(id),
    queryFn: () => getResearchTask(id),
    enabled: Boolean(id),
  });
export const useResearchContexts = (userId?: string) =>
  useQuery({
    queryKey: [...workflowKeys.all, "research-contexts", userId],
    queryFn: () => listResearchContexts(userId),
    enabled: Boolean(userId),
  });
export const useTaskContexts = (id: string) =>
  useQuery({
    queryKey: [...workflowKeys.task(id), "contexts"],
    queryFn: () => listTaskContexts(id),
    enabled: Boolean(id),
  });
export const useTaskLinks = (id: string) =>
  useQuery({
    queryKey: [...workflowKeys.task(id), "links"],
    queryFn: () => listTaskLinks(id),
    enabled: Boolean(id),
  });
export const useTaskOutputs = (id: string) =>
  useQuery({
    queryKey: [...workflowKeys.task(id), "outputs"],
    queryFn: () => listTaskOutputs(id),
    enabled: Boolean(id),
  });
export const useTaskRuns = (id: string) =>
  useQuery({
    queryKey: [...workflowKeys.task(id), "runs"],
    queryFn: () => listTaskRuns(id),
    enabled: Boolean(id),
  });
export const useTaskSources = (task: string, run: string) =>
  useQuery({
    queryKey: [...workflowKeys.task(task), run, "sources"],
    queryFn: () => listTaskSources(task, run),
    enabled: Boolean(task && run),
  });
export const useTaskFindings = (task: string, run: string) =>
  useQuery({
    queryKey: [...workflowKeys.task(task), run, "findings"],
    queryFn: () => listTaskFindings(task, run),
    enabled: Boolean(task && run),
  });
export const useProposals = (userId?: string) =>
  useQuery({
    queryKey: [...workflowKeys.proposals, userId],
    queryFn: () => listProposals(userId),
  });
export const useInformationRequests = (query = "") =>
  useQuery({
    queryKey: workflowKeys.requests(query),
    queryFn: () => listInformationRequests(query),
  });
export const useApplicationWorkspace = (id: string) => {
  const fields = useQuery({
    queryKey: [...workflowKeys.workspace(id), "fields"],
    queryFn: () => listFields(id),
    enabled: Boolean(id),
  });
  const questionnaires = useQuery({
    queryKey: [...workflowKeys.workspace(id), "questionnaires"],
    queryFn: () => listQuestionnaires(id),
    enabled: Boolean(id),
  });
  const information = useQuery({
    queryKey: [...workflowKeys.workspace(id), "information"],
    queryFn: () => listInformationRequests(`?applicationId=${id}`),
    enabled: Boolean(id),
  });
  const preparation = useQuery({
    queryKey: [...workflowKeys.workspace(id), "preparation"],
    queryFn: () => getPreparation(id),
    enabled: Boolean(id),
  });
  return { fields, questionnaires, information, preparation };
};
