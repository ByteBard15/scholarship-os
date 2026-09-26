import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import {
  getQuestionnaire,
  listAnswers,
  reviewAnswer,
  workflowKeys,
} from "../features/workflow/api";
import type { Question } from "../features/workflow/types";

export function QuestionnairePage() {
  const { applicationId = "", questionnaireId = "" } = useParams();
  const query = useQuery({
    queryKey: [...workflowKeys.workspace(applicationId), questionnaireId],
    queryFn: () => getQuestionnaire(applicationId, questionnaireId),
  });
  if (query.isPending) return <p>Loading questionnaire…</p>;
  if (query.isError)
    return <p className="text-red-700">{query.error.message}</p>;
  const questionnaire = query.data.data;
  return (
    <section>
      <Link className="text-indigo-700" to={`/applications/${applicationId}`}>
        ← Application
      </Link>
      <h1 className="mt-3 text-3xl font-semibold">{questionnaire.title}</h1>
      <p className="mt-2 text-slate-600">
        {questionnaire.status.replaceAll("_", " ")}
      </p>
      <div className="mt-6 grid gap-4">
        {questionnaire.questions?.map((question) => (
          <QuestionCard
            key={question.id}
            question={question}
            applicationId={applicationId}
          />
        ))}
      </div>
    </section>
  );
}
function QuestionCard({
  question,
  applicationId,
}: {
  question: Question;
  applicationId: string;
}) {
  const answers = useQuery({
    queryKey: [
      ...workflowKeys.workspace(applicationId),
      question.id,
      "answers",
    ],
    queryFn: () => listAnswers(question.id),
  });
  const client = useQueryClient();
  const review = useMutation({
    mutationFn: ({
      id,
      action,
    }: {
      id: string;
      action: "approve" | "reject";
    }) => reviewAnswer(question.id, id, action),
    onSuccess: () =>
      void client.invalidateQueries({
        queryKey: workflowKeys.workspace(applicationId),
      }),
  });
  return (
    <article className="rounded-lg border bg-white p-5">
      <h2 className="font-medium">{question.prompt}</h2>
      <p className="mt-1 text-xs text-slate-500">
        {question.isRequired ? "Required" : "Optional"} ·{" "}
        {question.questionType} · {question.status.replaceAll("_", " ")}
        {question.wordLimit ? ` · ${question.wordLimit} words` : ""}
      </p>
      <div className="mt-3 grid gap-2">
        {answers.data?.data.map((answer) => (
          <div key={answer.id} className="rounded border p-3">
            <p className="whitespace-pre-wrap text-sm">
              {answer.draftText ?? String(answer.value ?? "")}
            </p>
            <p className="mt-2 text-xs text-slate-500">
              {answer.status} · source {answer.answerSource ?? "manual"}
            </p>
            {answer.status === "suggested" && (
              <div className="mt-2 flex gap-2">
                <button
                  className="rounded bg-emerald-700 px-3 py-1 text-sm text-white"
                  onClick={() =>
                    review.mutate({ id: answer.id, action: "approve" })
                  }
                >
                  Approve
                </button>
                <button
                  className="rounded border px-3 py-1 text-sm"
                  onClick={() =>
                    review.mutate({ id: answer.id, action: "reject" })
                  }
                >
                  Reject
                </button>
              </div>
            )}
          </div>
        ))}
      </div>
    </article>
  );
}
