import {
  useRef,
  useState,
  type ButtonHTMLAttributes,
  type MouseEvent,
} from "react";

type Props = Omit<ButtonHTMLAttributes<HTMLButtonElement>, "onClick"> & {
  pending?: boolean;
  pendingLabel?: string;
  onClick?: (event: MouseEvent<HTMLButtonElement>) => void | Promise<unknown>;
};

export function RequestButton({
  pending = false,
  pendingLabel = "Working…",
  disabled,
  children,
  className = "",
  onClick,
  ...props
}: Props) {
  const [requestPending, setRequestPending] = useState(false);
  const [error, setError] = useState("");
  const active = useRef(false);
  const busy = pending || requestPending;
  return (
    <>
      <button
        {...props}
        className={`relative disabled:cursor-not-allowed disabled:opacity-60 ${className}`}
        disabled={disabled || busy}
        aria-busy={busy}
        onClick={async (event) => {
          if (active.current || pending) return;
          setError("");
          active.current = true;
          try {
            const result = onClick?.(event);
            if (result && typeof result.then === "function") {
              setRequestPending(true);
              await result;
            }
          } catch (cause) {
            setError(
              cause instanceof Error
                ? cause.message
                : "Request failed. Please try again.",
            );
          } finally {
            active.current = false;
            setRequestPending(false);
          }
        }}
      >
        <span className={busy ? "invisible" : undefined}>{children}</span>
        {busy && (
          <span
            className="absolute inset-0 flex items-center justify-center"
            role="status"
          >
            <span
              className="h-4 w-4 animate-spin rounded-full border-2 border-current border-r-transparent motion-reduce:animate-none"
              aria-hidden="true"
            />
            <span className="sr-only">{pendingLabel}</span>
          </span>
        )}
      </button>
      {error && (
        <span role="alert" className="text-sm text-red-700">
          {error}
        </span>
      )}
    </>
  );
}
