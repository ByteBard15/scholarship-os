import { useQuery } from "@tanstack/react-query";
import {
  getCompleteness,
  getProfile,
  listCandidates,
  listDocuments,
  listImports,
  listOverrides,
  listProfiles,
  listSnapshots,
  profileKeys,
} from "./api";
export const useProfiles = (userId: string) =>
  useQuery({
    queryKey: profileKeys.list(userId),
    queryFn: () => listProfiles(userId),
    enabled: Boolean(userId),
  });
export const useProfile = (id: string) =>
  useQuery({
    queryKey: profileKeys.detail(id),
    queryFn: () => getProfile(id),
    enabled: Boolean(id),
  });
export const useCompleteness = (id: string) =>
  useQuery({
    queryKey: [...profileKeys.detail(id), "completeness"],
    queryFn: () => getCompleteness(id),
    enabled: Boolean(id),
  });
export const useDocuments = (id: string) =>
  useQuery({
    queryKey: [...profileKeys.detail(id), "documents"],
    queryFn: () => listDocuments(id),
    enabled: Boolean(id),
  });
export const useImports = (id: string) =>
  useQuery({
    queryKey: [...profileKeys.detail(id), "imports"],
    queryFn: () => listImports(id),
    enabled: Boolean(id),
  });
export const useOverrides = (id: string) =>
  useQuery({
    queryKey: [...profileKeys.detail(id), "overrides"],
    queryFn: () => listOverrides(id),
    enabled: Boolean(id),
  });
export const useSnapshots = (id: string) =>
  useQuery({
    queryKey: [...profileKeys.detail(id), "snapshots"],
    queryFn: () => listSnapshots(id),
    enabled: Boolean(id),
  });
export const useImportCandidates = (profileId: string, importId: string) =>
  useQuery({
    queryKey: [
      ...profileKeys.detail(profileId),
      "imports",
      importId,
      "candidates",
    ],
    queryFn: () => listCandidates(profileId, importId),
    enabled: Boolean(profileId && importId),
  });
