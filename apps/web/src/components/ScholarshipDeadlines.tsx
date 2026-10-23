import type { ResearchTask } from "../features/workflow/types";

// Show source wording first: a vague date must never become an invented timestamp.
function deadlineTexts(value: unknown): string[] {
  if (typeof value === "string") return [value];
  if (Array.isArray(value)) return value.flatMap(deadlineTexts);
  if (!value || typeof value !== "object") return [];
  const record = value as Record<string, unknown>;
  for (const key of ["rawDeadlineText", "rawText"]) {
    if (typeof record[key] === "string" && record[key].trim())
      return [record[key]];
  }
  return Object.entries(record).flatMap(([key, item]) => {
    if (item && typeof item === "object") return deadlineTexts(item);
    if (
      typeof item === "string" &&
      /^(date|deadlineAt|deadline|remainingRoundDate|.*Deadline|.*DeadlineDate)$/i.test(
        key,
      )
    )
      return [item];
    return [];
  });
}

export function ScholarshipDeadlines({ task }: { task: ResearchTask }) {
  if (!["review_required", "completed"].includes(task.status)) return null;
  const deadlines = task.scholarshipDeadlines ?? [];
  if (!deadlines.length)
    return (
      <p className="mt-2 text-sm text-slate-500">
        Scholarship deadline: Not established by research.
      </p>
    );
  return (
    <div className="mt-3 text-sm">
      <p className="font-medium text-slate-700">Scholarship deadlines</p>
      <ul className="mt-1 grid gap-2">
        {deadlines.map((deadline) => {
          const texts = [...new Set(deadlineTexts(deadline.value))];
          const text = texts.length
            ? texts.join(" · ")
            : (deadline.rawText ?? "Exact date not established");
          const label = deadline.field
            .replace(/([a-z])([A-Z])/g, "$1 $2")
            .replaceAll("_", " ");
          return (
            <li
              key={deadline.findingId}
              className="border-l-2 border-indigo-200 pl-3"
            >
              <span className="capitalize text-slate-600">{label}: </span>
              <span>{text}</span>
              <a
                href={deadline.sourceUrl}
                target="_blank"
                rel="noreferrer"
                className="ml-2 text-indigo-700 underline"
              >
                Source
              </a>
              <span
                className={
                  deadline.verificationStatus === "conflicting"
                    ? "ml-2 text-amber-700"
                    : "ml-2 text-xs text-slate-500"
                }
              >
                {deadline.verificationStatus}
                {deadline.reviewStatus === "pending"
                  ? " · awaiting review"
                  : ""}
              </span>
              {deadline.verificationStatus === "conflicting" &&
                deadline.notes && (
                  <p className="mt-1 text-xs text-amber-700">
                    {deadline.notes}
                  </p>
                )}
            </li>
          );
        })}
      </ul>
    </div>
  );
}
