import { useQuery } from "@tanstack/react-query";
import { getProfile, listProfiles, profileKeys } from "./api";

export function useProfiles(userId: string) {
  return useQuery({
    queryKey: profileKeys.list(userId),
    queryFn: () => listProfiles(userId),
    enabled: Boolean(userId),
  });
}

export function useProfile(id: string) {
  return useQuery({
    queryKey: profileKeys.detail(id),
    queryFn: () => getProfile(id),
    enabled: Boolean(id),
  });
}
