import type { Metadata } from "next";

/**
 * Minimal placeholder page of the portal.
 *
 * Every destination referenced by `lib/site-structure.ts` must resolve, even
 * while its editorial content is not published yet. This component keeps those
 * routes reachable — and statically compilable — without depending on a
 * message catalog entry that does not exist yet. It is meant to be replaced by
 * a real page (see `theme-page.tsx`) as the content model is built.
 */
export function placeholderMetadata(title: string): Metadata {
  return { title };
}

export function PlaceholderPage({
  title,
  description,
}: {
  title: string;
  /** Optional sentence describing the destination. */
  description?: string;
}) {
  return (
    <section className="gov-section" aria-labelledby="placeholder-title">
      <div className="gov-section__container gov-prose">
        <p className="gov-kicker">CODE</p>
        <h1 id="placeholder-title">{title}</h1>
        <p className="gov-lead">
          {description ?? "Cette page est en cours de préparation."}
        </p>
      </div>
    </section>
  );
}
