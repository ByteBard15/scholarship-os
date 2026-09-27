import { getJSON, requestJSON } from "../../api/client";
import type {
  Answer,
  ApplicationField,
  ApplicationProposal,
  Collection,
  Envelope,
  InformationRequest,
  PreparationSummary,
  Questionnaire,
  ResearchFinding,
  ResearchContext,
  ResearchRun,
  ResearchSource,
  ResearchTask,
  ResearchTaskLink,
  ResearchTaskOutput,
} from "./types";

export const workflowKeys = {
  all: ["workflow"] as const,
  tasks: (query = "") => ["workflow", "tasks", query] as const,
  task: (id: string) => ["workflow", "task", id] as const,
  proposals: ["workflow", "proposals"] as const,
  requests: (query = "") => ["workflow", "information", query] as const,
  workspace: (id: string) => ["workflow", "application", id] as const,
};

export const listResearchTasks = (query = "") =>
  getJSON<Collection<ResearchTask>>(`/api/v1/research-tasks${query}`);
export const getResearchTask = (id: string) =>
  getJSON<Envelope<ResearchTask>>(`/api/v1/research-tasks/${id}`);
export const createResearchTask = (
  input: {
    userId: string;
    title: string;
    description?: string;
    instructions?: string;
    taskType: string;
    priority?: string;
    targetApplicationId?: string;
    profileId?: string;
    links: Array<{ label?: string; url: string; linkType?: string }>;
    researchContextIds: string[];
    newResearchContexts: Array<{ question: string; answer: string }>;
  },
  queue: boolean,
) =>
  requestJSON<Envelope<ResearchTask>>(
    `/api/v1/research-tasks${queue ? "?queue=true" : ""}`,
    { method: "POST", body: JSON.stringify(input) },
  );
export const transitionResearchTask = (
  id: string,
  action: "queue" | "cancel" | "retry",
) =>
  requestJSON<Envelope<ResearchTask>>(
    `/api/v1/research-tasks/${id}/${action}`,
    { method: "POST" },
  );
export const updateResearchTask = (
  id: string,
  input: {
    title?: string;
    description?: string;
    instructions?: string;
    taskType?: string;
    priority?: string;
    profileId?: string;
    clearProfile?: boolean;
  },
) =>
  requestJSON<Envelope<ResearchTask>>(`/api/v1/research-tasks/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
export const attachTaskContexts = (id: string, researchContextIds: string[]) =>
  requestJSON<Collection<ResearchContext>>(
    `/api/v1/research-tasks/${id}/contexts`,
    { method: "POST", body: JSON.stringify({ researchContextIds }) },
  );
export const detachTaskContext = (taskId: string, contextId: string) =>
  requestJSON<void>(`/api/v1/research-tasks/${taskId}/contexts/${contextId}`, {
    method: "DELETE",
  });
export const listTaskLinks = (id: string) =>
  getJSON<Collection<ResearchTaskLink>>(`/api/v1/research-tasks/${id}/links`);
export const listResearchContexts = (userId?: string) =>
  getJSON<Collection<ResearchContext>>(
    `/api/v1/research-contexts${userId ? `?userId=${userId}` : ""}`,
  );
export const createResearchContext = (input: {
  userId: string;
  question: string;
  answer: string;
}) =>
  requestJSON<Envelope<ResearchContext>>("/api/v1/research-contexts", {
    method: "POST",
    body: JSON.stringify(input),
  });
export const listTaskContexts = (id: string) =>
  getJSON<Collection<ResearchContext>>(`/api/v1/research-tasks/${id}/contexts`);
export const listTaskOutputs = (id: string) =>
  getJSON<Collection<ResearchTaskOutput>>(
    `/api/v1/research-tasks/${id}/outputs`,
  );
export const listTaskRuns = (id: string) =>
  getJSON<Collection<ResearchRun>>(`/api/v1/research-tasks/${id}/runs`);
export const listTaskSources = (task: string, run: string) =>
  getJSON<Collection<ResearchSource>>(
    `/api/v1/research-tasks/${task}/runs/${run}/sources`,
  );
export const listTaskFindings = (task: string, run: string) =>
  getJSON<Collection<ResearchFinding>>(
    `/api/v1/research-tasks/${task}/runs/${run}/findings`,
  );

export const listProposals = (userId?: string) =>
  getJSON<Collection<ApplicationProposal>>(
    `/api/v1/application-proposals${userId ? `?userId=${userId}` : ""}`,
  );
export const getProposal = (id: string) =>
  getJSON<Envelope<ApplicationProposal>>(`/api/v1/application-proposals/${id}`);
export const approveProposal = (id: string, parentProfileId: string) =>
  requestJSON<Envelope<ApplicationProposal>>(
    `/api/v1/application-proposals/${id}/approve`,
    { method: "POST", body: JSON.stringify({ parentProfileId }) },
  );
export const reviewProposal = (id: string, action: "reject" | "reopen") =>
  requestJSON<Envelope<ApplicationProposal>>(
    `/api/v1/application-proposals/${id}/${action}`,
    { method: "POST" },
  );

export const listFields = (applicationId: string) =>
  getJSON<Collection<ApplicationField>>(
    `/api/v1/applications/${applicationId}/fields`,
  );
export const listQuestionnaires = (applicationId: string) =>
  getJSON<Collection<Questionnaire>>(
    `/api/v1/applications/${applicationId}/questionnaires`,
  );
export const getQuestionnaire = (
  applicationId: string,
  questionnaireId: string,
) =>
  getJSON<Envelope<Questionnaire>>(
    `/api/v1/applications/${applicationId}/questionnaires/${questionnaireId}`,
  );
export const listAnswers = (questionId: string) =>
  getJSON<Collection<Answer>>(`/api/v1/questions/${questionId}/answers`);
export const reviewAnswer = (
  questionId: string,
  answerId: string,
  action: "approve" | "reject",
) =>
  requestJSON<Envelope<Answer>>(
    `/api/v1/questions/${questionId}/answers/${answerId}/${action}`,
    { method: "POST" },
  );
export const runPrefill = (applicationId: string, regenerate = false) =>
  requestJSON<Envelope<{ id: string; status: string }>>(
    `/api/v1/applications/${applicationId}/prefill`,
    {
      method: "POST",
      body: JSON.stringify({
        trigger: regenerate ? "regeneration" : "manual",
        regenerate,
      }),
    },
  );
export const getPreparation = (applicationId: string) =>
  getJSON<Envelope<PreparationSummary>>(
    `/api/v1/applications/${applicationId}/preparation`,
  );

export const listInformationRequests = (query = "") =>
  getJSON<Collection<InformationRequest>>(
    `/api/v1/information-requests${query}`,
  );
export const getInformationRequest = (id: string) =>
  getJSON<Envelope<InformationRequest>>(`/api/v1/information-requests/${id}`);
export const respondInformationRequest = (id: string, responseText: string) =>
  requestJSON<Envelope<InformationRequest>>(
    `/api/v1/information-requests/${id}/respond`,
    {
      method: "POST",
      body: JSON.stringify({ responseText, submittedBy: "user" }),
    },
  );
export const processInformationRequest = (id: string) =>
  requestJSON<Envelope<InformationRequest>>(
    `/api/v1/information-requests/${id}/process`,
    { method: "POST" },
  );
export const transitionInformationRequest = (
  id: string,
  action: "reopen" | "cancel",
) =>
  requestJSON<Envelope<InformationRequest>>(
    `/api/v1/information-requests/${id}/${action}`,
    { method: "POST" },
  );
