import { getJSON, requestJSON } from "../../api/client";
import type { CreateWritingSample, WritingSample, WritingTag } from "./types";

type Envelope<T> = { data: T };
type Collection<T> = { data: T[]; meta: { count: number } };
type NullableCollection<T> = {
  data: T[] | null;
  meta?: { count: number } | null;
};

function normalizeCollection<T>(collection: NullableCollection<T>): Collection<T> {
  const data = collection.data ?? [];
  return {
    data,
    meta: { count: collection.meta?.count ?? data.length },
  };
}

export const writingKeys = {
  all: ["writing"] as const,
  samples: ["writing", "samples"] as const,
  tags: ["writing", "tags"] as const,
};

export const listWritingSamples = async () =>
  normalizeCollection(
    await getJSON<NullableCollection<WritingSample>>("/api/v1/writing-samples"),
  );

export const createWritingSample = (request: CreateWritingSample) =>
  requestJSON<Envelope<WritingSample>>("/api/v1/writing-samples", {
    method: "POST",
    body: JSON.stringify(request),
  });

export const deleteWritingSample = (id: string) =>
  requestJSON<void>(`/api/v1/writing-samples/${id}`, { method: "DELETE" });

export const listWritingTags = async () =>
  normalizeCollection(
    await getJSON<NullableCollection<WritingTag>>("/api/v1/writing-tags"),
  );
