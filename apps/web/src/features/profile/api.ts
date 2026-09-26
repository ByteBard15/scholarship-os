import { getJSON, requestJSON } from "../../api/client";
import type {
  CollectionEnvelope,
  Completeness,
  DataEnvelope,
  EffectiveProfile,
  ImportCandidate,
  Profile,
  ProfileComparison,
  ProfileDocument,
  ProfileImport,
  ProfileOverride,
  ProfileSnapshot,
  ProfileType,
} from "./types";
export const profileKeys = {
  all: ["profiles"] as const,
  list: (userId: string) => [...profileKeys.all, "list", userId] as const,
  detail: (id: string) => [...profileKeys.all, "detail", id] as const,
};
export const listProfiles = (userId: string) =>
  getJSON<CollectionEnvelope<Profile>>(`/api/v1/users/${userId}/profiles`);
export const getProfile = (id: string) =>
  getJSON<DataEnvelope<EffectiveProfile>>(`/api/v1/profiles/${id}/effective`);
export const createProfile = (
  userId: string,
  input: {
    name: string;
    profileType: ProfileType;
    parentProfileId?: string;
  },
) =>
  requestJSON<DataEnvelope<Profile>>(`/api/v1/users/${userId}/profiles`, {
    method: "POST",
    body: JSON.stringify(input),
  });
export const getCompleteness = (id: string) =>
  getJSON<DataEnvelope<Completeness>>(`/api/v1/profiles/${id}/completeness`);
export const listDocuments = (id: string) =>
  getJSON<CollectionEnvelope<ProfileDocument>>(
    `/api/v1/profiles/${id}/documents`,
  );
export function uploadDocument(id: string, file: File) {
  const body = new FormData();
  body.append("documentType", "cv");
  body.append("file", file);
  return requestJSON<DataEnvelope<ProfileDocument>>(
    `/api/v1/profiles/${id}/documents`,
    { method: "POST", body },
  );
}
export const listImports = (id: string) =>
  getJSON<CollectionEnvelope<ProfileImport>>(`/api/v1/profiles/${id}/imports`);
export const createImport = (id: string, documentId: string) =>
  requestJSON<DataEnvelope<ProfileImport>>(`/api/v1/profiles/${id}/imports`, {
    method: "POST",
    body: JSON.stringify({ documentId }),
  });
export const listCandidates = (profileId: string, importId: string) =>
  getJSON<CollectionEnvelope<ImportCandidate>>(
    `/api/v1/profiles/${profileId}/imports/${importId}/candidates`,
  );
export const reviewCandidate = (
  profileId: string,
  importId: string,
  candidateId: string,
  action: "accept" | "reject" | "merge",
) =>
  requestJSON<DataEnvelope<ImportCandidate>>(
    `/api/v1/profiles/${profileId}/imports/${importId}/candidates/${candidateId}`,
    { method: "PATCH", body: JSON.stringify({ action }) },
  );
export const applyImport = (profileId: string, importId: string) =>
  requestJSON<DataEnvelope<ProfileImport>>(
    `/api/v1/profiles/${profileId}/imports/${importId}/apply`,
    { method: "POST" },
  );
export const listOverrides = (id: string) =>
  getJSON<CollectionEnvelope<ProfileOverride>>(
    `/api/v1/profiles/${id}/overrides`,
  );
export const createOverride = (
  id: string,
  input: {
    entityType: string;
    entityId?: string;
    fieldName: string;
    overrideType: "replace" | "hide" | "append";
    value?: unknown;
    reason?: string;
  },
) =>
  requestJSON<DataEnvelope<ProfileOverride>>(
    `/api/v1/profiles/${id}/overrides`,
    { method: "POST", body: JSON.stringify(input) },
  );
export const deleteOverride = (profileId: string, overrideId: string) =>
  requestJSON<void>(`/api/v1/profiles/${profileId}/overrides/${overrideId}`, {
    method: "DELETE",
  });
export const listSnapshots = (id: string) =>
  getJSON<CollectionEnvelope<ProfileSnapshot>>(
    `/api/v1/profiles/${id}/snapshots`,
  );
export const createSnapshot = (id: string, reason?: string) =>
  requestJSON<DataEnvelope<ProfileSnapshot>>(
    `/api/v1/profiles/${id}/snapshots`,
    { method: "POST", body: JSON.stringify({ reason }) },
  );
export const compareProfiles = (id: string, otherId: string) =>
  getJSON<DataEnvelope<ProfileComparison>>(
    `/api/v1/profiles/${id}/compare/${otherId}`,
  );
