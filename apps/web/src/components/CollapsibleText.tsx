import { useLayoutEffect, useRef, useState } from "react";

export function CollapsibleText({
  children,
  className = "",
  lines = 5,
}: {
  children: string;
  className?: string;
  lines?: number;
}) {
  const textRef = useRef<HTMLParagraphElement>(null);
  const [expanded, setExpanded] = useState(false);
  const [overflowing, setOverflowing] = useState(false);

  useLayoutEffect(() => {
    const element = textRef.current;
    if (!element || expanded) return;
    const measure = () => setOverflowing(element.scrollHeight > element.clientHeight + 1);
    measure();
    window.addEventListener("resize", measure);
    return () => window.removeEventListener("resize", measure);
  }, [children, expanded, lines]);

  return (
    <div className={className}>
      <p
        ref={textRef}
        className="whitespace-pre-wrap"
        style={
          expanded
            ? undefined
            : {
                display: "-webkit-box",
                WebkitBoxOrient: "vertical",
                WebkitLineClamp: lines,
                overflow: "hidden",
              }
        }
      >
        {children}
      </p>
      {(overflowing || expanded) && (
        <button
          aria-expanded={expanded}
          className="mt-1 text-xs font-medium text-indigo-700"
          onClick={(event) => {
            event.preventDefault();
            event.stopPropagation();
            setExpanded((value) => !value);
          }}
          type="button"
        >
          {expanded ? "View less" : "View more"}
        </button>
      )}
    </div>
  );
}
