export type ProfileType = "master" | "domain" | "application";
export type Profile = {
  id: string;
  userId: string;
  name: string;
  profileType: ProfileType;
  parentProfileId?: string;
  headline?: string;
  summary?: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
};
export type ProfileSectionKey =
  | "education"
  | "employment"
  | "projects"
  | "publications"
  | "articles"
  | "skills"
  | "research-interests"
  | "career-goals"
  | "certifications"
  | "awards"
  | "volunteering";
export type ProfileEntry = {
  id: string;
  profileId: string;
  inherited?: boolean;
  inheritedFromProfileId?: string;
  modifiedInProfileId?: string;
  institution?: string;
  degree?: string;
  fieldOfStudy?: string;
  organization?: string;
  jobTitle?: string;
  employmentType?: string;
  title?: string;
  name?: string;
  role?: string;
  publicationType?: string;
  publisher?: string;
  publicationDate?: string;
  startDate?: string;
  endDate?: string;
  isCurrent?: boolean;
  grade?: string;
  gradeScale?: string;
  classification?: string;
  city?: string;
  country?: string;
  description?: string;
  url?: string;
  repositoryUrl?: string;
  doi?: string;
  category?: string;
  proficiency?: string;
  yearsExperience?: number;
  priority?: number;
  goalType?: string;
  issuer?: string;
  issueDate?: string;
  expirationDate?: string;
  credentialId?: string;
  awardDate?: string;
  createdAt?: string;
  updatedAt?: string;
};
export type PersonalInfo = {
  id?: string;
  profileId?: string;
  inherited?: boolean;
  inheritedFromProfileId?: string;
  firstName?: string;
  middleName?: string;
  lastName?: string;
  preferredName?: string;
  phone?: string;
  city?: string;
  stateOrRegion?: string;
  country?: string;
  nationality?: string;
  linkedInUrl?: string;
  githubUrl?: string;
  websiteUrl?: string;
};
export type ProfileSectionInput = Omit<
  ProfileEntry,
  | "id"
  | "profileId"
  | "inherited"
  | "inheritedFromProfileId"
  | "modifiedInProfileId"
  | "createdAt"
  | "updatedAt"
>;
export type ProfileBundle = {
  schemaVersion: "1.0";
  exportedAt?: string;
  profile: {
    name: string;
    profileType: ProfileType;
    parentProfileId?: string;
    headline?: string;
    summary?: string;
    isDefault: boolean;
    personalInfo?: PersonalInfo;
    education: ProfileSectionInput[];
    employment: ProfileSectionInput[];
    projects: ProfileSectionInput[];
    publications: ProfileSectionInput[];
    articles: ProfileSectionInput[];
    skills: ProfileSectionInput[];
    researchInterests: ProfileSectionInput[];
    careerGoals: ProfileSectionInput[];
    certifications: ProfileSectionInput[];
    awards: ProfileSectionInput[];
    volunteering: ProfileSectionInput[];
  };
};
export type ProfileFull = Profile & {
  personalInfo: PersonalInfo | null;
  education: ProfileEntry[];
  employment: ProfileEntry[];
  projects: ProfileEntry[];
  publications: ProfileEntry[];
  articles: ProfileEntry[];
  skills: ProfileEntry[];
  researchInterests: ProfileEntry[];
  careerGoals: ProfileEntry[];
  certifications: ProfileEntry[];
  awards: ProfileEntry[];
  volunteering: ProfileEntry[];
};
export type LineageItem = { id: string; name: string; type: ProfileType };
export type EffectiveProfile = ProfileFull & { lineage: LineageItem[] };
export type Completeness = {
  score: number;
  sections: { name: string; status: string }[];
  notice: string;
};
export type ProfileDocument = {
  id: string;
  profileId: string;
  documentType: string;
  originalFilename: string;
  storageProvider: string;
  mimeType: string;
  fileSize: number;
  sha256?: string;
  status: string;
  createdAt: string;
  updatedAt: string;
};
export type ProfileImport = {
  id: string;
  profileId: string;
  documentId: string;
  importType: string;
  status: string;
  errorMessage?: string;
  createdAt: string;
  updatedAt: string;
};
export type ImportCandidate = {
  id: string;
  importId: string;
  sectionType: string;
  candidateData: Record<string, unknown>;
  sourceText?: string;
  confidence?: number;
  status: string;
  matchedEntityId?: string;
  existing?: Record<string, unknown>;
  changes?: { field: string; existing: unknown; candidate: unknown }[];
};
export type ProfileOverride = {
  id: string;
  profileId: string;
  entityType: string;
  entityId?: string;
  fieldName: string;
  overrideType: "replace" | "hide" | "append";
  value?: unknown;
  reason?: string;
  createdAt: string;
};
export type ProfileSnapshot = {
  id: string;
  profileId: string;
  version: number;
  createdAt: string;
  createdBy?: string;
  reason?: string;
};
export type ProfileComparison = {
  baseProfile: Profile;
  comparedProfile: Profile;
  inheritedEntities: ProfileEntry[];
  hiddenEntities: ProfileOverride[];
  modifiedFields: ProfileOverride[];
  appendedEntities: ProfileEntry[];
};
export type DataEnvelope<T> = { data: T };
export type CollectionEnvelope<T> = { data: T[]; meta: { count: number } };
