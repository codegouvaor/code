/**
 * URL structure of CODE — the public developer platform of Astoria's
 * administration.
 *
 * Hrefs are locale-agnostic pathnames: the next-intl Link (registered as the
 * ADS link renderer) prefixes the active locale automatically. Labels are
 * never stored here — they come from the message catalogs through the key
 * provided by each entry.
 *
 * The header navigation (`primaryNavigation`) is the spine of the platform,
 * organised on a **7 × 4 × 4** information architecture:
 *
 *   7 grands thèmes → jusqu'à 4 sections par thème → jusqu'à 4 entrées par section
 *
 * Each theme answers one principal intention of a Code visitor (discover,
 * build, document, contribute…), each section groups the related
 * destinations, and each entry is a single route. The same `primaryNavigation`
 * drives the desktop mega-menus, the hierarchical mobile drawer and the
 * sitemap, so the whole platform shares a single taxonomy.
 *
 * NOTE: most destinations below are provisional — the platform content model
 * is still being built. Entries point to routes that exist today, to the
 * canonical routes the platform is organised around, or to explicit external
 * destinations (GitHub releases, open data). No link is invented just to fill
 * a 4 × 4 cell.
 */
export const PORTAL_HOME = "/";

/**
 * Top-level hub pages of the portal information architecture, referenced by
 * the header navigation, the footer and the homepage. Like every platform
 * destination below they are provisional routes: the hub pages are not all
 * published yet.
 */
export const portalPaths = {
  documentation: "/documentation",
  standards: "/standards",
  ressources: "/ressources",
  projets: "/projets",
  communaute: "/communaute",
} as const;

/** Real platform routes (the GitHub-compatible platform of Code). */
export const platformPaths = {
  repositories: "/repositories",
  issues: "/issues",
  pulls: "/pulls",
  dashboard: "/dashboard",
  marketplace: "/marketplace",
  notifications: "/notifications",
  newRepo: "/new/repo",
  newOrg: "/new/organization",
  newProject: "/new/project",
  importRepo: "/new/import",
} as const;

export const legalPaths = {
  accessibility: "/legal/accessibility",
  privacy: "/legal/privacy",
  terms: "/legal/terms",
  cookies: "/legal/cookies",
  sitemap: "/sitemap",
} as const;

export const searchPath = "/search";

/** External service-status page of the platform (public roadmap/status). */
export const serviceStatusPath = "/statut-du-service";

/** Official open-data portal of the Republic of Astoria. */
export const openDataUrl = "https://data.gouv.aor/";

/** Source repository of the platform (external link shown in header/footer). */
export const codeRepositoryUrl = "https://github.com/codegouvaor/code";

/** Releases feed of the source repository — the news channel of the platform
 *  while no editorial flow exists yet. */
export const codeRepositoryReleasesUrl = `${codeRepositoryUrl}/releases`;

/** DOM ids used as skip-link targets. */
export const pageAnchors = {
  content: "main-content",
  footer: "main-footer",
} as const;

/** A single destination; its label resolves under `nav.panel`. */
export type PrimaryNavLink = {
  labelKey: string;
  href: string;
  external?: boolean;
};

/**
 * One entry of a mega-menu section. The heading label resolves under
 * `nav.panel`; when `href` is set the heading is itself a destination (ADS
 * `categoryMainLink`), otherwise it is plain text (`categoryMainText`).
 */
export type NavigationItem = {
  /** Message key (`nav.panel`) of the section heading. */
  labelKey: string;
  /** Destination of the section heading (its hub). */
  href: string;
  /** Up to four entries of the section. */
  links: ReadonlyArray<PrimaryNavLink>;
};

/** The leader band shown on top of a theme mega-menu panel. */
export type NavigationLeader = {
  /** Message key (`nav.panel`) of the band title. */
  titleKey: string;
  /** Message key (`nav.panel`) of the band paragraph. */
  paragraphKey: string;
  /** Main action of the band (“Voir tout”…). */
  link: PrimaryNavLink;
};

/** One of the seven themes of the Code header navigation. */
export type NavigationSection = {
  /** Message key (`nav.primary`) of the theme label; also the `nav.panel`
   *  `<labelKey>.title` key used by the sitemap heading. */
  labelKey: string;
  /** Route matched when marking the theme current. */
  href: string;
  leader: NavigationLeader;
  /** Up to four sections, each with up to four entries. */
  primaryItems: ReadonlyArray<NavigationItem>;
};

/**
 * The 7 × 4 × 4 navigation of Code.
 *
 * The seven themes are chosen from the real intentions of a developer on the
 * platform, not from an aesthetic list:
 *
 *   Explorer       → discover the ecosystem (projects, organisations, subjects)
 *   Projets        → the live project & repository platform of Code
 *   Organisations  → the organisations, teams, developers and communities
 *   Développement  → the tools and integrations to build public services
 *   Documentation  → the lifecycle of a public digital service
 *   Ressources     → the reusable building blocks (APIs, SDKs, components)
 *   Communauté     → contribute, propose and debate
 *
 * Every entry points to an existing route, a canonical route of the platform
 * or an explicit external destination. Top-level labels resolve under
 * `nav.primary`, everything nested under `nav.panel`.
 */
export const primaryNavigation: ReadonlyArray<NavigationSection> = [
  {
    labelKey: "explorer",
    href: portalPaths.projets,
    leader: {
      titleKey: "explorerLeaderTitle",
      paragraphKey: "explorerLeaderParagraph",
      link: { labelKey: "explorerLeaderLink", href: portalPaths.projets },
    },
    primaryItems: [
      {
        labelKey: "explorerCategoryDécouvrir",
        href: portalPaths.projets,
        links: [
          { labelKey: "projets", href: portalPaths.projets },
          { labelKey: "organisations", href: portalPaths.communaute },
          { labelKey: "communautes", href: portalPaths.communaute },
          { labelKey: "openSource", href: "/definir/open-source" },
        ],
      },
      {
        labelKey: "explorerCategoryActivite",
        href: codeRepositoryReleasesUrl,
        links: [
          { labelKey: "releases", href: codeRepositoryReleasesUrl, external: true },
          { labelKey: "nouveautes", href: portalPaths.projets },
          { labelKey: "openData", href: openDataUrl, external: true },
          { labelKey: "standards", href: portalPaths.standards },
        ],
      },
      {
        labelKey: "explorerCategorySujets",
        href: portalPaths.standards,
        links: [
          { labelKey: "standards", href: portalPaths.standards },
          { labelKey: "apis", href: "/construire/apis" },
          { labelKey: "donnees", href: "/definir/donnees" },
          { labelKey: "securite", href: "/definir/securite" },
        ],
      },
      {
        labelKey: "explorerCategoryRecherche",
        href: searchPath,
        links: [
          { labelKey: "rechercheGlobale", href: searchPath },
          { labelKey: "projets", href: portalPaths.projets },
          { labelKey: "documentation", href: portalPaths.documentation },
          { labelKey: "organisations", href: portalPaths.communaute },
        ],
      },
    ],
  },
  {
    labelKey: "projets",
    href: platformPaths.repositories,
    leader: {
      titleKey: "projetsLeaderTitle",
      paragraphKey: "projetsLeaderParagraph",
      link: { labelKey: "projetsLeaderLink", href: platformPaths.repositories },
    },
    primaryItems: [
      {
        labelKey: "projetsCategoryPlateforme",
        href: platformPaths.repositories,
        links: [
          { labelKey: "repositories", href: platformPaths.repositories },
          { labelKey: "issues", href: platformPaths.issues },
          { labelKey: "pulls", href: platformPaths.pulls },
          { labelKey: "notifications", href: platformPaths.notifications },
        ],
      },
      {
        labelKey: "projetsCategoryPublier",
        href: platformPaths.newRepo,
        links: [
          { labelKey: "newRepo", href: platformPaths.newRepo },
          { labelKey: "newOrg", href: platformPaths.newOrg },
          { labelKey: "newProject", href: platformPaths.newProject },
          { labelKey: "importRepo", href: platformPaths.importRepo },
        ],
      },
      {
        labelKey: "projetsCategoryCatalogue",
        href: portalPaths.projets,
        links: [
          { labelKey: "projets", href: portalPaths.projets },
          { labelKey: "marketplace", href: platformPaths.marketplace },
          { labelKey: "templates", href: "/construire/templates" },
          { labelKey: "packages", href: "/construire/packages" },
        ],
      },
      {
        labelKey: "projetsCategoryVersions",
        href: codeRepositoryReleasesUrl,
        links: [
          { labelKey: "releases", href: codeRepositoryReleasesUrl, external: true },
          { labelKey: "versions", href: "/faire-evoluer/versions" },
          { labelKey: "roadmap", href: "/faire-evoluer/roadmap" },
          { labelKey: "migrations", href: "/faire-evoluer/migrations" },
        ],
      },
    ],
  },
  {
    labelKey: "organisations",
    href: portalPaths.communaute,
    leader: {
      titleKey: "organisationsLeaderTitle",
      paragraphKey: "organisationsLeaderParagraph",
      link: { labelKey: "organisationsLeaderLink", href: portalPaths.communaute },
    },
    primaryItems: [
      {
        labelKey: "orgCategoryOrganisations",
        href: portalPaths.communaute,
        links: [
          { labelKey: "organisations", href: portalPaths.communaute },
          { labelKey: "equipes", href: portalPaths.communaute },
          { labelKey: "developpeurs", href: portalPaths.communaute },
          { labelKey: "contributions", href: "/contributions" },
        ],
      },
      {
        labelKey: "orgCategoryCommunautes",
        href: portalPaths.communaute,
        links: [
          { labelKey: "communautes", href: portalPaths.communaute },
          { labelKey: "openSource", href: "/definir/open-source" },
          { labelKey: "github", href: codeRepositoryUrl, external: true },
          { labelKey: "openData", href: openDataUrl, external: true },
        ],
      },
      {
        labelKey: "orgCategoryGouvernance",
        href: "/definir/gouvernance-technique",
        links: [
          { labelKey: "gouvernanceTechnique", href: "/definir/gouvernance-technique" },
          { labelKey: "interoperabilite", href: "/definir/interoperabilite" },
          { labelKey: "securite", href: "/definir/securite" },
          { labelKey: "accessibilite", href: "/definir/accessibilite" },
        ],
      },
      {
        labelKey: "orgCategoryDecouvrir",
        href: portalPaths.projets,
        links: [
          { labelKey: "projets", href: portalPaths.projets },
          { labelKey: "communaute", href: portalPaths.communaute },
          { labelKey: "standards", href: portalPaths.standards },
          { labelKey: "documentation", href: portalPaths.documentation },
        ],
      },
    ],
  },
  {
    labelKey: "developpement",
    href: "/construire",
    leader: {
      titleKey: "developpementLeaderTitle",
      paragraphKey: "developpementLeaderParagraph",
      link: { labelKey: "developpementLeaderLink", href: "/construire" },
    },
    primaryItems: [
      {
        labelKey: "devCategoryOutils",
        href: "/construire/cli",
        links: [
          { labelKey: "cli", href: "/construire/cli" },
          { labelKey: "packages", href: "/construire/packages" },
          { labelKey: "templates", href: "/construire/templates" },
          { labelKey: "ads", href: "/construire/ads" },
        ],
      },
      {
        labelKey: "devCategoryIntegration",
        href: "/construire/apis",
        links: [
          { labelKey: "apis", href: "/construire/apis" },
          { labelKey: "sdk", href: "/construire/sdk" },
          { labelKey: "webhooks", href: "/construire/webhooks" },
          { labelKey: "backend", href: "/construire/backend" },
        ],
      },
      {
        labelKey: "devCategoryFrontend",
        href: "/construire/frontend",
        links: [
          { labelKey: "frontend", href: "/construire/frontend" },
          { labelKey: "backend", href: "/construire/backend" },
          { labelKey: "accessibilite", href: "/definir/accessibilite" },
          { labelKey: "tests", href: "/construire/tests" },
        ],
      },
      {
        labelKey: "devCategoryPlateformes",
        href: "/definir/interoperabilite",
        links: [
          { labelKey: "github", href: codeRepositoryUrl, external: true },
          { labelKey: "gitlab", href: "/definir/interoperabilite" },
          { labelKey: "giteria", href: "/definir/interoperabilite" },
          { labelKey: "interoperabilite", href: "/definir/interoperabilite" },
        ],
      },
    ],
  },
  {
    labelKey: "documentation",
    href: portalPaths.documentation,
    leader: {
      titleKey: "documentationLeaderTitle",
      paragraphKey: "documentationLeaderParagraph",
      link: { labelKey: "documentationLeaderLink", href: portalPaths.documentation },
    },
    primaryItems: [
      {
        labelKey: "docCategoryDefinir",
        href: "/definir",
        links: [
          { labelKey: "principes", href: "/definir/principes" },
          { labelKey: "standards", href: "/definir/standards" },
          { labelKey: "securite", href: "/definir/securite" },
          { labelKey: "accessibilite", href: "/definir/accessibilite" },
        ],
      },
      {
        labelKey: "docCategoryConcevoir",
        href: "/concevoir",
        links: [
          { labelKey: "conceptionServices", href: "/concevoir/conception-de-services" },
          { labelKey: "architecture", href: "/concevoir/architecture" },
          { labelKey: "apis", href: "/concevoir/apis" },
          { labelKey: "identite", href: "/concevoir/identite" },
        ],
      },
      {
        labelKey: "docCategoryConstruire",
        href: "/construire",
        links: [
          { labelKey: "frontend", href: "/construire/frontend" },
          { labelKey: "backend", href: "/construire/backend" },
          { labelKey: "packages", href: "/construire/packages" },
          { labelKey: "sdk", href: "/construire/sdk" },
        ],
      },
      {
        labelKey: "docCategoryCycle",
        href: "/faire-evoluer",
        links: [
          { labelKey: "deployer", href: "/deployer" },
          { labelKey: "exploiter", href: "/exploiter" },
          { labelKey: "faireEvoluer", href: "/faire-evoluer" },
          { labelKey: "roadmap", href: "/faire-evoluer/roadmap" },
        ],
      },
    ],
  },
  {
    labelKey: "ressources",
    href: portalPaths.ressources,
    leader: {
      titleKey: "ressourcesLeaderTitle",
      paragraphKey: "ressourcesLeaderParagraph",
      link: { labelKey: "ressourcesLeaderLink", href: portalPaths.ressources },
    },
    primaryItems: [
      {
        labelKey: "resCategoryApis",
        href: "/construire/apis",
        links: [
          { labelKey: "apis", href: "/construire/apis" },
          { labelKey: "donnees", href: "/definir/donnees" },
          { labelKey: "openData", href: openDataUrl, external: true },
          { labelKey: "interoperabilite", href: "/definir/interoperabilite" },
        ],
      },
      {
        labelKey: "resCategoryComposants",
        href: "/construire/packages",
        links: [
          { labelKey: "packages", href: "/construire/packages" },
          { labelKey: "sdk", href: "/construire/sdk" },
          { labelKey: "cli", href: "/construire/cli" },
          { labelKey: "templates", href: "/construire/templates" },
        ],
      },
      {
        labelKey: "resCategoryDesign",
        href: "/construire/ads",
        links: [
          { labelKey: "ads", href: "/construire/ads" },
          { labelKey: "frontend", href: "/construire/frontend" },
          { labelKey: "accessibilite", href: "/definir/accessibilite" },
          { labelKey: "referencesConception", href: "/concevoir/references-de-conception" },
        ],
      },
      {
        labelKey: "resCategoryInfrastructure",
        href: "/deployer/infrastructure",
        links: [
          { labelKey: "cloud", href: "/deployer/cloud" },
          { labelKey: "reseaux", href: "/deployer/reseaux" },
          { labelKey: "dns", href: "/deployer/dns" },
          { labelKey: "containers", href: "/deployer/containers" },
        ],
      },
    ],
  },
  {
    labelKey: "communaute",
    href: portalPaths.communaute,
    leader: {
      titleKey: "communauteLeaderTitle",
      paragraphKey: "communauteLeaderParagraph",
      link: { labelKey: "communauteLeaderLink", href: portalPaths.communaute },
    },
    primaryItems: [
      {
        labelKey: "comCategoryContribuer",
        href: "/contributions",
        links: [
          { labelKey: "contributions", href: "/contributions" },
          { labelKey: "openSource", href: "/definir/open-source" },
          { labelKey: "rfc", href: "/contributions/rfc" },
          { labelKey: "communaute", href: portalPaths.communaute },
        ],
      },
      {
        labelKey: "comCategoryProposer",
        href: portalPaths.standards,
        links: [
          { labelKey: "standards", href: portalPaths.standards },
          { labelKey: "gouvernanceTechnique", href: "/definir/gouvernance-technique" },
          { labelKey: "innovation", href: "/propose-and-discuss/innovation" },
          { labelKey: "recherche", href: "/propose-and-discuss/recherche" },
        ],
      },
      {
        labelKey: "comCategoryEvolutions",
        href: "/faire-evoluer/roadmap",
        links: [
          { labelKey: "roadmap", href: "/faire-evoluer/roadmap" },
          { labelKey: "versions", href: "/faire-evoluer/versions" },
          { labelKey: "migrations", href: "/faire-evoluer/migrations" },
          { labelKey: "depreciations", href: "/faire-evoluer/depreciations" },
        ],
      },
      {
        labelKey: "comCategoryDecouvrir",
        href: portalPaths.projets,
        links: [
          { labelKey: "communaute", href: portalPaths.communaute },
          { labelKey: "projets", href: portalPaths.projets },
          { labelKey: "github", href: codeRepositoryUrl, external: true },
          { labelKey: "releases", href: codeRepositoryReleasesUrl, external: true },
        ],
      },
    ],
  },
];

/**
 * Platform-level destinations shown in the footer bottom bar: they answer
 * “what can I do on CODE?” (documentation, catalogs, projects, community)
 * and give access to the source repository.
 *
 * Labels resolve under `nav.panel`. `github` points outside the platform and
 * is rendered as an external link by the footer.
 */
export const footerPlatformLinks: ReadonlyArray<PrimaryNavLink> = [
  { labelKey: "documentation", href: portalPaths.documentation },
  { labelKey: "standards", href: portalPaths.standards },
  { labelKey: "ressources", href: portalPaths.ressources },
  { labelKey: "projets", href: portalPaths.projets },
  { labelKey: "communaute", href: portalPaths.communaute },
  { labelKey: "github", href: codeRepositoryUrl },
];

/**
 * Bottom bar of the footer — institutional and legal information of the
 * platform. `legalPaths` and the external destinations above feed the
 * component; labels resolve under `footer.bottom`.
 */
export const footerLegalLinks: ReadonlyArray<PrimaryNavLink & { external?: boolean }> = [
  { labelKey: "openData", href: openDataUrl, external: true },
  { labelKey: "serviceStatus", href: serviceStatusPath },
  { labelKey: "privacy", href: legalPaths.privacy },
  { labelKey: "cookies", href: legalPaths.cookies },
];