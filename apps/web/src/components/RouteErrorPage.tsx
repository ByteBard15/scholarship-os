import { Link, isRouteErrorResponse, useRouteError } from "react-router-dom";

export function RouteErrorPage() {
  const error = useRouteError();
  const message = isRouteErrorResponse(error)
    ? error.statusText || "The requested page could not be loaded."
    : error instanceof Error
      ? error.message
      : "An unexpected error occurred.";

  return (
    <main className="mx-auto flex min-h-screen max-w-xl items-center px-6">
      <section className="w-full rounded-lg border bg-white p-6 shadow-sm">
        <p className="text-sm font-medium text-red-700">Something went wrong</p>
        <h1 className="mt-2 text-2xl font-semibold">This page could not be displayed</h1>
        <p className="mt-3 text-sm text-slate-600">{message}</p>
        <div className="mt-6 flex gap-3">
          <button
            className="rounded bg-indigo-700 px-4 py-2 text-sm text-white"
            onClick={() => window.location.reload()}
            type="button"
          >
            Try again
          </button>
          <Link className="rounded border px-4 py-2 text-sm" to="/">
            Go to dashboard
          </Link>
        </div>
      </section>
    </main>
  );
}
