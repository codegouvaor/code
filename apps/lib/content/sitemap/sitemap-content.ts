import { legalPaths } from "@/lib/site-structure";

/**
 * Content configuration of the sitemap page (`/sitemap`).
 *
 * The main part of the page is driven by `primaryNavigation`
 * (`lib/site-structure.ts`): the same configuration that builds the header and
 * the footer, so the sitemap can never drift apart from the navigation. Only
 * the transversal pages (utility pages reachable from the footer) live here.
 * Every display string is resolved from the message catalogs
 * (`pages.sitemap.*`).
 */
export const sitemapTransversalLinks = [
  { key: "contact", href: "/contact" },
  { key: "accessibility", href: legalPaths.accessibility },
  { key: "privacy", href: legalPaths.privacy },
  { key: "terms", href: legalPaths.terms },
  { key: "cookies", href: legalPaths.cookies },
  { key: "publications", href: "/publications-officielles" },
  { key: "documentsOpposables", href: "/documents-opposables" },
] as const;