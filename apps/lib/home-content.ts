/**
 * Content of the CODE homepage — the public developer platform of Astoria's
 * administration.
 *
 * The homepage is rendered from structured data, never from hard-coded JSX
 * strings. Titles, descriptions and leads live in the message catalogs
 * (`home.*`); this module only carries the message keys, the destination
 * hrefs, the DSFR icons and the structure.
 *
 * CODE is a *developer platform*, not a Git forge: it presents projects,
 * organisations, documentation, standards, APIs, SDKs and technical
 * resources in a single common interface while staying interoperable with
 * the development platforms it references (GitHub, GitLab, Giteria…). The
 * conceptual model is `Code Project → Repository Binding → GitHub / GitLab /
 * Giteria`: a project is independent from the forge that hosts its code.
 *
 * Most destinations below are the canonical routes of the platform (the same
 * plan the header navigation and the footer use, see `site-structure.ts`).
 * Once real content exists, these entries will be fed by the content backend
 * instead of static data.
 */
import {
  codeRepositoryUrl,
  codeRepositoryReleasesUrl,
  portalPaths,
} from "./site-structure";

/** DSFR icon used by a card/entry of the homepage. */
export type CodeHomeIconId = string;

/** One entry of a homepage grid: a message key, an optional destination and
 *  a DSFR icon. */
export type CodeHomeEntry = {
  /** Message key resolved under `home.<section>.items.<key>`. */
  key: string;
  href?: string;
  iconId: CodeHomeIconId;
};

/** A homepage entry that is always a link (href is mandatory). */
export type CodeHomeLink = CodeHomeEntry & { href: string };

/* -------------------------------------------------------------------------- *
 * 02 — Explorer Code
 * -------------------------------------------------------------------------- */

export const homeExploreDoors: ReadonlyArray<CodeHomeLink> = [
  { key: "projets", href: portalPaths.projets, iconId: "fr-icon-folder-2-line" },
  { key: "organisations", href: portalPaths.communaute, iconId: "fr-icon-building-line" },
  { key: "documentation", href: portalPaths.documentation, iconId: "fr-icon-book-2-line" },
  { key: "standards", href: portalPaths.standards, iconId: "fr-icon-flag-line" },
  { key: "apis", href: "/construire/apis", iconId: "fr-icon-links-line" },
  { key: "sdks", href: "/construire/sdk", iconId: "fr-icon-code-s-slash-line" },
];

/* -------------------------------------------------------------------------- *
 * 03 — Les projets publics
 * -------------------------------------------------------------------------- *
 * No project is invented: while the public referencing process is not open,
 * the section explains the interoperable model of CODE (Projet → Liaison de
 * dépôt → Forge) and points to the projects catalog.
 */

export const homeProjectsModel: ReadonlyArray<CodeHomeEntry> = [
  { key: "project", iconId: "fr-icon-folder-2-line" },
  { key: "binding", iconId: "fr-icon-links-line" },
  { key: "forge", iconId: "fr-icon-code-s-slash-line" },
];

export const homeProjectsCtaHref = portalPaths.projets;

/* -------------------------------------------------------------------------- *
 * 04 — Un écosystème ouvert
 * -------------------------------------------------------------------------- */

export const homeEcosystemPlatforms: ReadonlyArray<{
  key: string;
  href?: string;
  external?: boolean;
}> = [
  { key: "github", href: codeRepositoryUrl, external: true },
  { key: "gitlab" },
  { key: "giteria" },
];

/* -------------------------------------------------------------------------- *
 * 05 — Développer les services publics
 * -------------------------------------------------------------------------- */

export const homeResourcesItems: ReadonlyArray<CodeHomeLink> = [
  { key: "apis", href: "/construire/apis", iconId: "fr-icon-links-line" },
  { key: "sdks", href: "/construire/sdk", iconId: "fr-icon-code-s-slash-line" },
  { key: "composants", href: "/construire/frontend", iconId: "fr-icon-settings-5-line" },
  { key: "outils", href: "/construire/cli", iconId: "fr-icon-tools-line" },
  { key: "services", href: "/construire/backend", iconId: "fr-icon-server-line" },
  { key: "standards", href: portalPaths.standards, iconId: "fr-icon-flag-line" },
];

export const homeResourcesCtaHref = portalPaths.ressources;

/* -------------------------------------------------------------------------- *
 * 06 — La documentation
 * -------------------------------------------------------------------------- */

export const homeDocumentationItems: ReadonlyArray<CodeHomeLink> = [
  { key: "guides", href: portalPaths.documentation, iconId: "fr-icon-book-2-line" },
  { key: "referencesApi", href: "/construire/apis", iconId: "fr-icon-links-line" },
  { key: "technique", href: "/construire/sdk", iconId: "fr-icon-code-s-slash-line" },
  { key: "standards", href: portalPaths.standards, iconId: "fr-icon-flag-line" },
  { key: "projectsDoc", href: portalPaths.projets, iconId: "fr-icon-folder-2-line" },
];

export const homeDocumentationCtaHref = portalPaths.documentation;

/* -------------------------------------------------------------------------- *
 * 07 — Organisations et communautés
 * -------------------------------------------------------------------------- */

export const homeOrganizationItems: ReadonlyArray<CodeHomeEntry> = [
  { key: "organisations", iconId: "fr-icon-building-line" },
  { key: "equipes", iconId: "fr-icon-team-line" },
  { key: "developpeurs", iconId: "fr-icon-user-line" },
  { key: "communautes", iconId: "fr-icon-community-line" },
  { key: "contributions", iconId: "fr-icon-git-merge-line" },
];

export const homeOrganizationsCtaHref = portalPaths.communaute;

/* -------------------------------------------------------------------------- *
 * 08 — Contribuer
 * -------------------------------------------------------------------------- */

export const homeContributeItems: ReadonlyArray<CodeHomeEntry> = [
  { key: "code", iconId: "fr-icon-code-s-slash-line" },
  { key: "issues", iconId: "fr-icon-alert-line" },
  { key: "pullRequests", iconId: "fr-icon-git-pull-request-line" },
  { key: "documentation", iconId: "fr-icon-book-2-line" },
  { key: "standards", iconId: "fr-icon-flag-line" },
  { key: "securite", iconId: "fr-icon-shield-line" },
  { key: "plateforme", iconId: "fr-icon-settings-5-line" },
  { key: "ecosysteme", iconId: "fr-icon-share-line" },
];

export const homeContributeCtaHref = "/contributions";

/* -------------------------------------------------------------------------- *
 * 09 — Actualités et activité
 * -------------------------------------------------------------------------- */

export const homeActivityItems: ReadonlyArray<CodeHomeEntry> = [
  { key: "releases", iconId: "fr-icon-arrow-up-circle-line" },
  { key: "projets", iconId: "fr-icon-folder-2-line" },
  { key: "misesAJour", iconId: "fr-icon-refresh-line" },
  { key: "annonces", iconId: "fr-icon-information-line" },
  { key: "activite", iconId: "fr-icon-community-line" },
];

/** News channel of the platform: the public repository releases feed, used
 *  while no editorial flow exists yet. */
export const homeActivityHref = codeRepositoryReleasesUrl;
export const homeActivityExternal = true;

/* -------------------------------------------------------------------------- *
 * 10 — CTA final
 * -------------------------------------------------------------------------- */

export const homeFinalCtaLinks: ReadonlyArray<{ key: string; href: string }> = [
  { key: "cta1", href: portalPaths.projets },
  { key: "cta2", href: portalPaths.documentation },
  { key: "cta3", href: "/contributions" },
];

/* -------------------------------------------------------------------------- *
 * External destinations reused by the homepage
 * -------------------------------------------------------------------------- */

export const homeExternalLinks = {
  githubUrl: codeRepositoryUrl,
} as const;