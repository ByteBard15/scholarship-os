import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { CollapsibleText } from "../components/CollapsibleText";
import {
  createWritingSample,
  deleteWritingSample,
  listWritingSamples,
  listWritingTags,
  writingKeys,
} from "../features/writing/api";
import type { WritingDocumentType } from "../features/writing/types";

const documentTypes: Array<{ value: WritingDocumentType; label: string }> = [
  { value: "personal_statement", label: "Personal statement" },
  { value: "motivation_letter", label: "Motivation letter" },
  { value: "statement_of_purpose", label: "Statement of purpose" },
  { value: "recommendation_letter", label: "Recommendation letter" },
  { value: "essay", label: "Essay" },
  { value: "cover_letter", label: "Cover letter" },
  { value: "other", label: "Other" },
];

export function WritingLibraryPage() {
  const client = useQueryClient();
  const samples = useQuery({
    queryKey: writingKeys.samples,
    queryFn: listWritingSamples,
  });
  const tags = useQuery({
    queryKey: writingKeys.tags,
    queryFn: listWritingTags,
  });
  const [title, setTitle] = useState("");
  const [documentType, setDocumentType] =
    useState<WritingDocumentType>("personal_statement");
  const [description, setDescription] = useState("");
  const [tagText, setTagText] = useState("");
  const [content, setContent] = useState("");
  const create = useMutation({
    mutationFn: createWritingSample,
    onSuccess: () => {
      setTitle("");
      setDescription("");
      setTagText("");
      setContent("");
      void client.invalidateQueries({ queryKey: writingKeys.all });
    },
  });
  const remove = useMutation({
    mutationFn: deleteWritingSample,
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: writingKeys.all }),
  });
  const availableTagNames = (tags.data?.data ?? [])
    .map((tag) => tag.name)
    .join(", ");

  return (
    <section>
      <h1 className="text-3xl font-semibold">Writing Library</h1>
      <p className="mt-2 max-w-3xl text-sm text-slate-600">
        Add factual source writing for application agents to find by tag. Agent
        drafts remain suggestions and preserve links to the source samples they
        used.
      </p>

      <form
        className="mt-6 grid gap-3 rounded-lg border bg-white p-5 md:grid-cols-2"
        onSubmit={(event) => {
          event.preventDefault();
          create.mutate({
            title,
            documentType,
            content,
            description: description || undefined,
            tagNames: tagText
              .split(",")
              .map((tag) => tag.trim())
              .filter(Boolean),
          });
        }}
      >
        <h2 className="font-semibold md:col-span-2">Add source writing</h2>
        <label className="grid gap-1 text-sm">
          Title
          <input
            className="rounded border px-3 py-2"
            required
            value={title}
            onChange={(event) => setTitle(event.target.value)}
          />
        </label>
        <label className="grid gap-1 text-sm">
          Document type
          <select
            className="rounded border px-3 py-2"
            value={documentType}
            onChange={(event) =>
              setDocumentType(event.target.value as WritingDocumentType)
            }
          >
            {documentTypes.map((type) => (
              <option key={type.value} value={type.value}>
                {type.label}
              </option>
            ))}
          </select>
        </label>
        <label className="grid gap-1 text-sm md:col-span-2">
          Description
          <input
            className="rounded border px-3 py-2"
            placeholder="What this writing is useful for"
            value={description}
            onChange={(event) => setDescription(event.target.value)}
          />
        </label>
        <label className="grid gap-1 text-sm md:col-span-2">
          Tags
          <input
            className="rounded border px-3 py-2"
            placeholder="biomedical engineering, leadership, healthcare access"
            value={tagText}
            onChange={(event) => setTagText(event.target.value)}
          />
          <span className="text-xs text-slate-500">
            Separate tags with commas. Existing tags:{" "}
            {availableTagNames || "none"}.
          </span>
        </label>
        <label className="grid gap-1 text-sm md:col-span-2">
          Writing
          <textarea
            className="min-h-64 rounded border px-3 py-2"
            required
            value={content}
            onChange={(event) => setContent(event.target.value)}
          />
        </label>
        <button
          className="rounded bg-indigo-700 px-4 py-2 text-white md:col-span-2 disabled:opacity-50"
          disabled={create.isPending}
          type="submit"
        >
          {create.isPending ? "Saving…" : "Save writing sample"}
        </button>
        {create.isError && (
          <p className="text-sm text-red-700 md:col-span-2">
            {create.error.message}
          </p>
        )}
      </form>

      <div className="mt-8 flex items-center justify-between">
        <h2 className="text-xl font-semibold">Available samples</h2>
        <span className="text-sm text-slate-500">
          {samples.data?.meta.count ?? 0} samples
        </span>
      </div>
      {samples.isPending && <p className="mt-4">Loading writing samples…</p>}
      {samples.isError && (
        <p className="mt-4 text-red-700">{samples.error.message}</p>
      )}
      <div className="mt-4 grid gap-4">
        {(samples.data?.data ?? []).map((sample) => (
          <article className="rounded-lg border bg-white p-5" key={sample.id}>
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h3 className="font-medium">{sample.title}</h3>
                <p className="mt-1 text-xs uppercase text-slate-500">
                  {sample.documentType.replaceAll("_", " ")} · {sample.origin} ·{" "}
                  {sample.status}
                </p>
              </div>
              <button
                className="text-sm text-red-700"
                disabled={remove.isPending}
                onClick={() => remove.mutate(sample.id)}
                type="button"
              >
                Delete
              </button>
            </div>
            {sample.description && (
              <p className="mt-3 text-sm text-slate-600">
                {sample.description}
              </p>
            )}
            <div className="mt-3 flex flex-wrap gap-2">
              {(sample.tags ?? []).map((tag) => (
                <span
                  className="rounded-full bg-indigo-50 px-2 py-1 text-xs text-indigo-700"
                  key={tag.id}
                >
                  {tag.name}
                </span>
              ))}
            </div>
            <CollapsibleText
              className="mt-4 whitespace-pre-wrap text-sm text-slate-700"
              lines={5}
            >
              {sample.content}
            </CollapsibleText>
            {sample.sourceSampleIds?.length ? (
              <p className="mt-3 text-xs text-slate-500">
                Derived from {sample.sourceSampleIds.length} source sample(s)
              </p>
            ) : null}
          </article>
        ))}
      </div>
    </section>
  );
}
