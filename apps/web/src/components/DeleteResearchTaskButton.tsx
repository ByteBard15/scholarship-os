import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  cascadeDeleteResearch,
  previewResearchDeletion,
} from "../features/workflow/api";
import { Dialog } from "./Dialog";
import { RequestButton } from "./RequestButton";

export function DeleteResearchTaskButton({
  id,
  onDeleted,
}: {
  id: string;
  onDeleted?: () => void;
}) {
  const [open, setOpen] = useState(false);
  const [confirmation, setConfirmation] = useState("");
  const client = useQueryClient();
  const preview = useQuery({
    queryKey: ["research-deletion", id],
    queryFn: () => previewResearchDeletion(id),
    enabled: open,
    staleTime: 0,
  });
  const deletion = useMutation({
    mutationFn: () =>
      cascadeDeleteResearch(id, preview.data!.data.confirmationToken),
    onSuccess: () => {
      setOpen(false);
      onDeleted?.();
      void client.invalidateQueries();
    },
    onError: () => {
      setConfirmation("");
      void preview.refetch();
    },
  });
  const data = preview.data?.data;
  return (
    <>
      <button
        type="button"
        className="mt-2 rounded border border-red-200 px-3 py-1 text-sm text-red-700"
        onClick={() => {
          setConfirmation("");
          deletion.reset();
          setOpen(true);
        }}
      >
        Delete research task
      </button>
      <Dialog
        open={open}
        title="Permanently delete research and linked applications"
        busy={deletion.isPending}
        onClose={() => setOpen(false)}
      >
        <div className="space-y-4 p-5">
          <p>
            This permanently deletes the selected research task, its child
            research tasks, proposals, and applications created from those
            proposals, including their tasks, evidence, forms and research
            records.
          </p>
          <p className="text-sm text-slate-600">
            Reusable writing, master and domain profiles, and shared scholarship
            records are retained. Isolated application profiles are archived to
            preserve writing provenance. Applications on university websites are
            unaffected.
          </p>
          {preview.isFetching && <p role="status">Loading deletion preview…</p>}
          {preview.error && <p role="alert">{preview.error.message}</p>}
          {data && (
            <>
              {(
                [
                  ["Research tasks", data.tasks],
                  ["Proposals", data.proposals],
                  ["Applications", data.applications],
                  ["Profiles to archive", data.archivedProfiles],
                ] as const
              ).map(([label, items]) => (
                <div key={label}>
                  <h3 className="font-semibold">
                    {label} ({items.length})
                  </h3>
                  <ul className="max-h-40 list-disc overflow-auto pl-5">
                    {items.map((item) => (
                      <li key={item.id}>{item.name}</li>
                    ))}
                  </ul>
                </div>
              ))}
              {data.blockers.map((message, index) => (
                <p key={index} role="alert" className="text-red-700">
                  {message}
                </p>
              ))}
              <label className="block">
                Type DELETE to confirm
                <input
                  className="mt-1 block rounded border p-2"
                  value={confirmation}
                  onChange={(event) => setConfirmation(event.target.value)}
                  disabled={deletion.isPending}
                  autoComplete="off"
                />
              </label>
              <RequestButton
                className="rounded bg-red-700 px-4 py-2 text-white"
                pending={deletion.isPending}
                disabled={
                  confirmation !== "DELETE" ||
                  data.blockers.length > 0 ||
                  preview.isFetching ||
                  preview.isError
                }
                onClick={() => deletion.mutate()}
              >
                Permanently delete
              </RequestButton>
            </>
          )}
          {deletion.error && (
            <p role="alert" className="text-red-700">
              {deletion.error.message}
            </p>
          )}
        </div>
      </Dialog>
    </>
  );
}
