import { getJSON } from "../../api/client";
import type {
  CollectionEnvelope,
  DataEnvelope,
  Profile,
  ProfileFull,
} from "./types";

export const profileKeys = {
  all: ["profiles"] as const,
  list: (userId: string) => [...profileKeys.all, "list", userId] as const,
  detail: (id: string) => [...profileKeys.all, "detail", id] as const,
};

export function listProfiles(userId: string) {
  return getJSON<CollectionEnvelope<Profile>>(
    `/api/v1/users/${userId}/profiles`,
  );
}

export function getProfile(id: string) {
  return getJSON<DataEnvelope<ProfileFull>>(`/api/v1/profiles/${id}/full`);
}
