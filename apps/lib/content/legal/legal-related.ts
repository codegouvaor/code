import { legalPaths } from "@/lib/site-structure";

/**
 * Cross-links of the legal information zone of the portal, shared by every
 * legal page and by the legal landing page.
 *
 * This module centralises the *structure* of the legal navigation (their keys
 * and hrefs) — not their editorial content, which stays in each page's
 * dedicated `{page}-content.ts` file and in the message catalogs. Labels and
 * descriptions are resolved from `pages.legal.related.<key>.*` by
 * `lib/theme-localize.ts`.
 */
const legalPages = [
  { key: "accessibility", href: legalPaths.accessibility },
  { key: "privacy", href: legalPaths.privacy },
  { key: "terms", href: legalPaths.terms },
  { key: "cookies", href: legalPaths.cookies },
  { key: "sitemap", href: legalPaths.sitemap },
] as const;

export type LegalRelatedLink = (typeof legalPages)[number];

/** The legal documents of the portal, as cross-links. */
export function legalRelatedDestinations(): ReadonlyArray<LegalRelatedLink> {
  return legalPages;
}