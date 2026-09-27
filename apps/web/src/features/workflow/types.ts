export type Envelope<T> = { data: T };
export type Collection<T> = { data: T[]; meta: { count: number } };

export type ResearchTaskStatus =
  | "draft"
  | "ready"
  | "queued"
  | "running"
  | "review_required"
  | "completed"
  | "failed"
  | "cancelled";

export interface ResearchTask {
  id: string;
  userId: string;
  parentTaskId?: string;
  targetApplicationId?: string;
  profileId?: string;
  title: string;
  description?: string;
  instructions?: string;
  taskType: string;
  status: ResearchTaskStatus;
  priority?: string;
  dueAt?: string;
  researchConfig?: unknown;
  startedAt?: string;
  completedAt?: string;
  failedAt?: string;
  failureReason?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ResearchTaskLink {
  id: string;
  researchTaskId: string;
  label?: string;
  url: string;
  linkType?: string;
}

export interface ResearchContext {
  id: string;
  userId: string;
  question: string;
  answer: string;
  createdAt: string;
  updatedAt: string;
}

export interface ResearchTaskOutput {
  id: string;
  researchTaskId: string;
  outputType: string;
  entityId: string;
  createdAt: string;
}

export interface ResearchRun {
  id: string;
  researchTaskId?: string;
  applicationId?: string;
  status: string;
  modelProvider?: string;
  modelName?: string;
  createdAt: string;
}

export interface ResearchSource {
  id: string;
  researchRunId: string;
  url: string;
  title?: string;
  publisher?: string;
  sourceType: string;
  isOfficial: boolean;
  retrievedAt: string;
}

export interface ResearchFinding {
  id: string;
  researchRunId: string;
  sourceId?: string;
  category: string;
  field: string;
  value: unknown;
  confidence?: number;
  verificationStatus: string;
  reviewStatus: string;
  rawText?: string;
}

export interface ApplicationProposal {
  id: string;
  userId: string;
  researchTaskId: string;
  researchRunId?: string;
  name: string;
  country?: string;
  intake?: string;
  intakeYear?: number;
  summary?: string;
  status: "pending" | "approved" | "rejected" | "superseded";
  confidence?: number;
  reasoningSummary?: string;
  proposedInstitution?: {
    name: string;
    shortName?: string;
    institutionType?: string;
    country: string;
    city?: string;
    websiteUrl?: string;
  };
  proposedProgramme?: {
    name: string;
    degreeLevel?: string;
    fieldOfStudy?: string;
    faculty?: string;
    department?: string;
    durationMonths?: number;
    mode?: string;
    language?: string;
    programmeUrl?: string;
    description?: string;
  };
  proposedScholarship?: {
    name: string;
    providerName?: string;
    description?: string;
    country?: string;
    degreeLevel?: string;
    officialUrl?: string;
    scholarshipType?: string;
    isRecurring?: boolean;
  };
  sources: Array<{
    id: string;
    researchSourceId?: string;
    sourceRole?: string;
    url: string;
    title?: string;
    isOfficial: boolean;
  }>;
  createdAt: string;
}

export interface ApplicationField {
  id: string;
  applicationId: string;
  key: string;
  label: string;
  value?: unknown;
  valueType: string;
  status: string;
  sourceType?: string;
}

export interface Question {
  id: string;
  questionnaireId: string;
  key?: string;
  prompt: string;
  questionType: string;
  isRequired: boolean;
  wordLimit?: number;
  characterLimit?: number;
  status: string;
}

export interface Questionnaire {
  id: string;
  applicationId: string;
  title: string;
  questionnaireType?: string;
  status: string;
  sourceUrl?: string;
  questions?: Question[];
}

export interface Answer {
  id: string;
  questionId: string;
  value?: unknown;
  draftText?: string;
  status: string;
  answerSource?: string;
  confidence?: number;
}

export interface InformationRequest {
  id: string;
  userId: string;
  applicationId?: string;
  questionnaireId?: string;
  questionId?: string;
  applicationFieldId?: string;
  requestType: string;
  title: string;
  prompt: string;
  context?: string;
  status: string;
  responseType: string;
  responseValue?: unknown;
  responseText?: string;
  responses: Array<{
    id: string;
    responseText?: string;
    responseValue?: unknown;
    submittedBy: string;
    createdAt: string;
  }>;
  createdAt: string;
}

export interface PreparationSummary {
  fieldsCompleted: number;
  fieldsTotal: number;
  questionnairesCompleted: number;
  questionnairesTotal: number;
  pendingInformation: number;
  requiredQuestionsAnswered: number;
  requiredQuestionsTotal: number;
}
