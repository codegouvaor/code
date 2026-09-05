/**
 * Content of the CODE.GOUV.AOR homepage.
 *
 * The homepage is rendered from structured data, never from hard-coded JSX
 * strings. Labels and descriptions live in the message catalogs (`nav.panel`
 * for shared ecosystem nouns, `home.*` for homepage copy); this module only
 * carries the message keys, the destination hrefs and the structure.
 *
 * The platform content model is still being built: hrefs below point to the
 * routes the platform is organised around (the same plan the header
 * navigation and the footer use). Once real content exists, these entries
 * will be fed by the content backend instead of static data.
 */
import { codeRepositoryUrl } from "./site-structure";

/**
 * “I want to build…” — intention-oriented entry points for visitors who do
 * not know the organisation of CODE yet. Each intention points at the
 * lifecycle resource that matches its starting point.
 */
export const homeIntentions: ReadonlyArray<{ key: string; href: string }> = [
  { key: "website", href: "/construire/frontend" },
  { key: "application", href: "/concevoir/conception-de-services" },
  { key: "api", href: "/construire/apis" },
  { key: "adminService", href: "/concevoir/conception-de-services" },
  { key: "infrastructure", href: "/deployer/infrastructure" },
  { key: "package", href: "/construire/packages" },
];

/**
 * Foundations — the cross-cutting references of CODE. The `key` doubles as
 * the `nav.panel` label key (Architecture, Sécurité, APIs, Données,
 * Accessibilité, Open Source); the description comes from
 * `home.foundations.items.<key>`.
 */
export const homeFoundations: ReadonlyArray<{ key: string; href: string }> = [
  { key: "architecture", href: "/concevoir/architecture" },
  { key: "securite", href: "/definir/securite" },
  { key: "apis", href: "/concevoir/apis" },
  { key: "donnees", href: "/definir/donnees" },
  { key: "accessibilite", href: "/definir/accessibilite" },
  { key: "openSource", href: "/definir/open-source" },
];

/**
 * Technical resources of the ecosystem. Labels come from `nav.panel` (via
 * `labelKey`), descriptions from `home.resources.items.<key>`. The Design
 * System entry is the one external link that already exists (ADS, the
 * `@codegouvaor/react-ads` repository).
 */
export const homeResources: ReadonlyArray<
  { key: string; href: string; labelKey: string; external?: boolean }
> = [
  { key: "packages", labelKey: "packages", href: "/construire/packages" },
  { key: "sdks", labelKey: "sdk", href: "/construire/sdk" },
  { key: "cli", labelKey: "cli", href: "/construire/cli" },
  { key: "templates", labelKey: "templates", href: "/construire/templates" },
  { key: "apis", labelKey: "apis", href: "/construire/apis" },
  {
    key: "designSystem",
    labelKey: "designSystem",
    href: codeRepositoryUrl,
    external: true,
  },
];

/**
 * Community actions of the “Take part in public digital technology” section.
 * Labels and descriptions live under `home.community.items.<key>`.
 */
export const homeCommunityActions: ReadonlyArray<{ key: string; href: string }> = [
  { key: "contribute", href: "/faire-evoluer/contributions" },
  { key: "standard", href: "/standards" },
  { key: "rfc", href: "/faire-evoluer/rfc" },
  { key: "package", href: "/construire/packages" },
  { key: "explore", href: "/projets" },
];

/**
 * “Pour / Par / Avec les agents” doorways — the editorial heart of the
 * homepage (mirrors the reference portal art direction). Each doorway
 * answers a single intention: use the standards, publish your code, join
 * the community. Labels/descriptions live under `home.doors.items.<key>`.
 */
export const homeDoors: ReadonlyArray<{ key: string; href: string }> = [
  { key: "forAdmin", href: "/standards" },
  { key: "byAdmin", href: "/faire-evoluer/open-source" },
  { key: "community", href: "/faire-evoluer/communaute" },
];

/**
 * “Vos contributions sont les bienvenues” — the accordion questions of the
 * homepage. Each opens an answer whose `href` carries the action to take;
 * the anchor label lives inside the message (a `<link>` tag).
 */
export const homeContributions: ReadonlyArray<{ key: string; href: string }> = [
  { key: "standard", href: "/faire-evoluer/rfc" },
  { key: "repositories", href: "/faire-evoluer/open-source" },
  { key: "events", href: "/faire-evoluer/communaute" },
];

/**
 * External destinations of the “follow” band: the public source repository
 * of the platform (and its releases feed), used as the news channel of the
 * platform while no editorial/blog flow exists yet.
 */
export const homeFollowLinks = {
  githubUrl: codeRepositoryUrl,
  releasesUrl: `${codeRepositoryUrl}/releases`,
} as const;
