export type WritingDocumentType =
  | "personal_statement"
  | "motivation_letter"
  | "recommendation_letter"
  | "statement_of_purpose"
  | "essay"
  | "cover_letter"
  | "other";

export interface WritingTag {
  id: string;
  name: string;
  sampleCount: number;
}

export interface WritingSample {
  id: string;
  userId: string;
  applicationId?: string;
  title: string;
  documentType: WritingDocumentType;
  content: string;
  origin: "user" | "agent" | "import";
  status: "source" | "draft" | "suggested" | "approved" | "archived";
  description?: string;
  createdBy?: string;
  tags: WritingTag[];
  sourceSampleIds?: string[];
  createdAt: string;
  updatedAt: string;
}

export interface CreateWritingSample {
  title: string;
  documentType: WritingDocumentType;
  content: string;
  description?: string;
  tagNames: string[];
}
