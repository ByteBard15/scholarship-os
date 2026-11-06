import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Dialog } from "./Dialog";
import { RequestButton } from "./RequestButton";
import { updateWritingSample, writingKeys } from "../features/writing/api";
import type {
  WritingDocumentType,
  WritingSample,
} from "../features/writing/types";

export function WritingEditDialog({
  sample,
  onClose,
}: {
  sample: WritingSample;
  onClose: () => void;
}) {
  const client = useQueryClient();
  const [title, setTitle] = useState(sample.title);
  const [documentType, setDocumentType] = useState(sample.documentType);
  const [description, setDescription] = useState(sample.description ?? "");
  const [content, setContent] = useState(sample.content);
  const [tags, setTags] = useState(
    sample.tags.map((tag) => tag.name).join(", "),
  );
  const save = useMutation({
    mutationFn: () =>
      updateWritingSample(sample.id, {
        title: title.trim(),
        documentType,
        description,
        content,
        tagNames: [
          ...new Set(
            tags
              .split(",")
              .map((tag) => tag.trim())
              .filter(Boolean),
          ),
        ],
      }),
    onSuccess: () => {
      void client.invalidateQueries({ queryKey: writingKeys.all });
      onClose();
    },
  });
  return (
    <Dialog
      open
      title="Edit writing sample"
      busy={save.isPending}
      onClose={onClose}
    >
      <form
        className="grid gap-4 p-5"
        onSubmit={(event) => {
          event.preventDefault();
          save.mutate();
        }}
      >
        <p className="text-sm text-slate-500">
          Saving preserves this sample’s application, source provenance and
          review status.
        </p>
        <fieldset disabled={save.isPending} className="grid gap-4">
          <label className="grid gap-1 text-sm">
            Title
            <input
              className="rounded border p-2"
              required
              value={title}
              onChange={(event) => setTitle(event.target.value)}
            />
          </label>
          <label className="grid gap-1 text-sm">
            Document type
            <select
              className="rounded border p-2"
              value={documentType}
              onChange={(event) =>
                setDocumentType(event.target.value as WritingDocumentType)
              }
            >
              {(
                [
                  "personal_statement",
                  "motivation_letter",
                  "statement_of_purpose",
                  "recommendation_letter",
                  "essay",
                  "cover_letter",
                  "email",
                  "other",
                ] as const
              ).map((type) => (
                <option key={type} value={type}>
                  {type.replaceAll("_", " ")}
                </option>
              ))}
            </select>
          </label>
          <label className="grid gap-1 text-sm">
            Description
            <textarea
              className="rounded border p-2"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
            />
          </label>
          <label className="grid gap-1 text-sm">
            Tags
            <input
              className="rounded border p-2"
              value={tags}
              onChange={(event) => setTags(event.target.value)}
            />
            <span className="text-xs text-slate-500">
              Separate tags with commas. Clear this field to remove all tags.
            </span>
          </label>
          <label className="grid gap-1 text-sm">
            Writing
            <textarea
              className="min-h-80 rounded border p-2"
              required
              value={content}
              onChange={(event) => setContent(event.target.value)}
            />
          </label>
        </fieldset>
        {save.isError && (
          <p role="alert" className="text-sm text-red-700">
            {save.error.message}
          </p>
        )}
        <div className="flex justify-end gap-3">
          <button
            type="button"
            className="rounded border px-4 py-2"
            disabled={save.isPending}
            onClick={onClose}
          >
            Cancel
          </button>
          <RequestButton
            type="submit"
            className="rounded bg-indigo-700 px-4 py-2 text-white disabled:opacity-50"
            pending={save.isPending}
            disabled={save.isPending || !title.trim() || !content.trim()}
          >
            Save changes
          </RequestButton>
        </div>
      </form>
    </Dialog>
  );
}
