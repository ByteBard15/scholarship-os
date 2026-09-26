export type ApplicationStatus =
  | "discovered"
  | "researching"
  | "research_complete"
  | "eligibility_review"
  | "preparing"
  | "documents_in_progress"
  | "ready_to_submit"
  | "submitted"
  | "waiting"
  | "accepted"
  | "rejected"
  | "withdrawn"
  | "expired";
export type ResearchStatus =
  | "not_started"
  | "queued"
  | "running"
  | "review_required"
  | "complete"
  | "failed"
  | "stale";
export type NamedResource = { id: string; name: string };
export type Application = {
  id: string;
  userId: string;
  applicantProfileId: string;
  institution?: NamedResource;
  programme?: NamedResource;
  scholarship?: NamedResource;
  name: string;
  intake?: string;
  intakeYear?: number;
  country?: string;
  status: ApplicationStatus;
  priority?: number;
  notes?: string;
  researchStatus: ResearchStatus;
  createdAt: string;
  updatedAt: string;
};
export type Requirement = {
  id: string;
  applicationId: string;
  category: string;
  title: string;
  description?: string;
  isMandatory: boolean;
  status: string;
  dueDate?: string;
  sourceId?: string;
  sourceUrl?: string;
  evidenceRequired: boolean;
  notes?: string;
  sortOrder?: number;
  createdAt: string;
  updatedAt: string;
};
export type Deadline = {
  id: string;
  applicationId: string;
  deadlineType: string;
  title: string;
  deadlineAt?: string;
  timezone?: string;
  datePrecision: "exact" | "day" | "month" | "approximate" | "unknown";
  rawDeadlineText?: string;
  isHardDeadline: boolean;
  sourceId?: string;
  sourceUrl?: string;
  verifiedAt?: string;
  notes?: string;
  urgency: string;
  createdAt: string;
  updatedAt: string;
};
export type Funding = {
  id: string;
  applicationId: string;
  fundingType: string;
  currency?: string;
  amount?: number;
  amountPeriod?: string;
  tuitionCoverage?: string;
  stipendAmount?: number;
  stipendPeriod?: string;
  travelCoverage?: string;
  insuranceCoverage?: string;
  accommodationCoverage?: string;
  otherBenefits?: string;
  conditions?: string;
  sourceId?: string;
  sourceUrl?: string;
  createdAt: string;
  updatedAt: string;
};
export type Contact = {
  id: string;
  applicationId: string;
  name?: string;
  role?: string;
  email?: string;
  phone?: string;
  organization?: string;
  contactType?: string;
  url?: string;
  notes?: string;
  sourceId?: string;
};
export type Supervisor = {
  id: string;
  applicationId: string;
  name: string;
  title?: string;
  department?: string;
  institution?: string;
  email?: string;
  profileUrl?: string;
  researchAreas?: string;
  contactStatus?: string;
  notes?: string;
  sourceId?: string;
};
export type ApplicationURL = {
  id: string;
  applicationId: string;
  urlType: string;
  label?: string;
  url: string;
  isOfficial: boolean;
};
export type ApplicationTask = {
  id: string;
  applicationId: string;
  parentTaskId?: string;
  title: string;
  description?: string;
  status: "todo" | "in_progress" | "blocked" | "done" | "cancelled";
  priority?: number;
  dueAt?: string;
  completedAt?: string;
  sortOrder?: number;
  taskType?: string;
  source?: string;
  createdAt: string;
  updatedAt: string;
};
export type ResearchRun = {
  id: string;
  applicationId: string;
  status: string;
  trigger: string;
  researchType: string;
  modelProvider?: string;
  modelName?: string;
  promptVersion?: string;
  startedAt?: string;
  completedAt?: string;
  errorMessage?: string;
  createdAt: string;
  updatedAt: string;
};
export type ResearchSource = {
  id: string;
  researchRunId: string;
  applicationId: string;
  url: string;
  title?: string;
  publisher?: string;
  sourceType: string;
  isOfficial: boolean;
  retrievedAt: string;
  publishedAt?: string;
  contentHash?: string;
  notes?: string;
  createdAt: string;
};
export type ResearchFinding = {
  id: string;
  researchRunId: string;
  applicationId: string;
  sourceId?: string;
  category: string;
  field: string;
  value: unknown;
  confidence?: number;
  verificationStatus: string;
  reviewStatus: "pending" | "accepted" | "rejected" | "auto_accepted";
  rawText?: string;
  notes?: string;
  createdAt: string;
};
export type Readiness = {
  status: string;
  completedItems: number;
  totalItems: number;
  notice: string;
};
export type ApplicationSummary = Application & {
  nextDeadline?: Deadline;
  taskCompletion: { completed: number; total: number };
  requirementCompletion: { completed: number; total: number };
  readiness: Readiness;
};
export type ApplicationDetail = {
  application: Application;
  requirements: Requirement[];
  deadlines: Deadline[];
  funding: Funding[];
  contacts: Contact[];
  supervisors: Supervisor[];
  urls: ApplicationURL[];
  tasks: ApplicationTask[];
  latestResearchRun?: ResearchRun;
};
export type Institution = {
  id: string;
  name: string;
  shortName?: string;
  institutionType?: string;
  country: string;
  city?: string;
  websiteUrl?: string;
  createdAt: string;
  updatedAt: string;
};
export type Programme = {
  id: string;
  institutionId: string;
  name: string;
  degreeLevel?: string;
  fieldOfStudy?: string;
  faculty?: string;
  department?: string;
  programmeUrl?: string;
};
export type Scholarship = {
  id: string;
  institutionId?: string;
  name: string;
  providerName?: string;
  country?: string;
  degreeLevel?: string;
  scholarshipType?: string;
  officialUrl?: string;
};
export type Dashboard = {
  activeApplications: Application[];
  upcomingDeadlines: Deadline[];
  tasksDueSoon: ApplicationTask[];
  applicationsNeedingResearch: number;
  researchPendingReview: number;
  unresolvedResearchConflicts: number;
  incompleteMandatoryRequirements: number;
  statusSummary: Partial<Record<ApplicationStatus, number>>;
};
export type DataEnvelope<T> = { data: T };
export type CollectionEnvelope<T> = { data: T[]; meta: { count: number } };
