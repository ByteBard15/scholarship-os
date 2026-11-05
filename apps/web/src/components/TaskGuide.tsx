import { Fragment, type ReactNode } from "react";

// React escapes text; only HTTP(S) references become navigable links.
function references(text: string): ReactNode[] {
  const pattern = /\[([^\]\n]+)\]\((https?:\/\/[^\s)]+)\)|https?:\/\/[^\s<>]+/g;
  const parts: ReactNode[] = [];
  let cursor = 0;
  for (const match of text.matchAll(pattern)) {
    const index = match.index ?? 0;
    parts.push(text.slice(cursor, index));
    const rawUrl = match[2] ?? match[0];
    const url = match[2] ? rawUrl : rawUrl.replace(/[.,;:!?\])]+$/, "");
    const label = match[1] ?? url;
    parts.push(
      <Fragment key={index}>
        <a
          className="break-words font-medium text-indigo-700 underline underline-offset-2 hover:text-indigo-900"
          href={url}
          rel="noopener noreferrer"
          target="_blank"
        >
          {label}
        </a>
        {match[2] ? "" : rawUrl.slice(url.length)}
      </Fragment>,
    );
    cursor = index + match[0].length;
  }
  parts.push(text.slice(cursor));
  return parts;
}

export function TaskGuide({ description }: { description: string }) {
  return (
    <div className="mt-3 whitespace-pre-wrap break-words border-t border-slate-200 pt-3 text-sm leading-6 text-slate-700">
      {references(description)}
    </div>
  );
}
