export type Profile = {
  id: string;
  userId: string;
  name: string;
  headline?: string;
  summary?: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
};

export type ProfileFull = Profile & {
  personalInfo: Record<string, unknown> | null;
  education: Record<string, unknown>[];
  employment: Record<string, unknown>[];
  projects: Record<string, unknown>[];
  publications: Record<string, unknown>[];
  articles: Record<string, unknown>[];
  skills: Record<string, unknown>[];
  researchInterests: Record<string, unknown>[];
  careerGoals: Record<string, unknown>[];
  certifications: Record<string, unknown>[];
  awards: Record<string, unknown>[];
  volunteering: Record<string, unknown>[];
};

export type DataEnvelope<T> = { data: T };
export type CollectionEnvelope<T> = { data: T[]; meta: { count: number } };
