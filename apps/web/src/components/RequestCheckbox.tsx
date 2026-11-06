import { useRef, useState } from "react";

export function RequestCheckbox({ checked, label, onChange }: {
  checked: boolean;
  label: string;
  onChange: () => Promise<unknown>;
}) {
  const active = useRef(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");
  return <div className="mt-1">
    <span className="relative inline-flex h-4 w-4 items-center justify-center">
      <input type="checkbox" checked={checked} disabled={pending} aria-label={label} aria-busy={pending}
        className={pending ? "opacity-0" : ""}
        onChange={async () => {
          if (active.current) return;
          active.current = true;
          setPending(true);
          setError("");
          try { await onChange(); }
          catch (cause) { setError(cause instanceof Error ? cause.message : "Unable to update task. Please try again."); }
          finally { active.current = false; setPending(false); }
        }} />
      {pending && <span role="status" className="pointer-events-none absolute inset-0 flex items-center justify-center text-indigo-700">
        <span aria-hidden="true" className="h-4 w-4 animate-spin rounded-full border-2 border-current border-r-transparent motion-reduce:animate-none" />
        <span className="sr-only">Updating task…</span>
      </span>}
    </span>
    {error && <p role="alert" className="mt-1 max-w-xs text-sm text-red-700">{error}</p>}
  </div>;
}
