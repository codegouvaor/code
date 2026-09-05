/**
 * URL structure of CODE.GOUV.AOR — the public platform of reference to
 * define, design, build, deploy, operate and evolve Astoria's public digital
 * infrastructure.
 *
 * Hrefs are locale-agnostic pathnames: the next-intl Link (registered as the
 * ADS link renderer) prefixes the active locale automatically. Labels are
 * never stored here — they come from the message catalogs through the key
 * provided by each entry (`nav.primary.*` for top-level entries,
 * `nav.panel.*` for everything nested).
 *
 * The header navigation (`headerNavigation`) is a government-portal
 * information architecture answering « Que puis-je trouver sur
 * CODE.GOUV.AOR ? »: Accueil and Actualités are simple links (homepage and
 * news channel), and the four other sections — Communauté, Documentation,
 * Ressources and À propos — each open a mega-menu of exactly four sections
 * with at least five links per section. The six lifecycle themes
 * (`lifecycleThemes` below) are the methodological structure of the
 * *documentation* — answering « Comment construire le numérique public ? » —
 * and they are grouped into the four sections of the Documentation
 * mega-menu.
 *
 * NOTE: most destinations below are provisional. The platform content model
 * is still being built; each theme hub (`/definir`, `/concevoir`, …) and its
 * child resources are the URLs this platform will publish.
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

/**
 * The six lifecycle themes — the methodological structure of the CODE
 * documentation. They form a loop: DÉFINIR → CONCEVOIR → CONSTRUIRE →
 * DÉPLOYER → EXPLOITER → FAIRE ÉVOLUER → ↺. Each theme is one stage of the
 * lifecycle of a public digital service and answers a single intention:
 *
 *   Définir        → Quelles sont les règles, principes et références qui
 *                     définissent le numérique public ?
 *   Concevoir      → Comment transforme-t-on un besoin public en solution
 *                     numérique ?
 *   Construire     → Comment réalise-t-on concrètement un service numérique ?
 *   Déployer       → Comment met-on une solution en production ?
 *   Exploiter      → Comment fait-on fonctionner durablement un service
 *                     public ?
 *   Faire évoluer  → Comment améliore-t-on et fait-on évoluer collectivement
 *                     le numérique public ?
 *
 * The header does NOT expose the themes as six top-level entries anymore:
 * they are grouped in the four sections of the Documentation mega-menu (see
 * `headerNavigation`), and the homepage renders them as its « Explore the
 * platform » cards.
 */
export const themePaths = {
  definir: "/definir",
  concevoir: "/concevoir",
  construire: "/construire",
  deployer: "/deployer",
  exploiter: "/exploiter",
  faireEvoluer: "/faire-evoluer",
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
 *  (the homepage “follow” band uses it while no editorial flow exists yet). */
export const codeRepositoryReleasesUrl = `${codeRepositoryUrl}/releases`;

/** DOM ids used as skip-link targets. */
export const pageAnchors = {
  content: "main-content",
  footer: "main-footer",
} as const;

/** A link inside a mega-menu panel; its label is a `nav.panel` message key. */
export type PrimaryNavLink = {
  labelKey: string;
  href: string;
};

/**
 * One column of a theme mega-menu. The heading is plain text (a
 * `nav.panel` message key rendered by ADS as `categoryMainText`), never a
 * destination — the leader link is the main entry of the theme.
 */
export type ThemeCategory = {
  /** Message key (`nav.panel`) of the column heading. */
  titleKey: string;
  links: ReadonlyArray<PrimaryNavLink>;
};

/**
 * One of the six lifecycle themes — the methodological structure of the
 * documentation. It feeds the Documentation mega-menu columns (heading +
 * deep links, via `headerNavigation`) and the homepage lifecycle cards. The
 * top-level label resolves under `nav.primary`, every other label under
 * `nav.panel`. The footer deliberately does not repeat the themes.
 */
export type LifecycleTheme = {
  /** Message key (`nav.primary`) of the theme label. */
  labelKey: string;
  /** Theme hub route. */
  href: string;
  /** Leader band shown on top of the header mega-menu panel. */
  leader: {
    /** Message key (`nav.panel`) of the paragraph answering the theme question. */
    paragraphKey: string;
    /** Message key (`nav.panel`) of the leader action (“Tout Définir”…). */
    linkLabelKey: string;
  };
  /** Mega-menu columns. The mega-menu is not the sitemap of the theme: only
   *  the resources that matter to the user journey of this phase. */
  categories: ReadonlyArray<ThemeCategory>;
};

/**
 * The six lifecycle themes of the platform. Top-level labels resolve under
 * `nav.primary`, panel content (paragraphs, categories, links) under
 * `nav.panel`.
 */
export const lifecycleThemes: ReadonlyArray<LifecycleTheme> = [
  {
    labelKey: "definir",
    href: themePaths.definir,
    leader: {
      paragraphKey: "definirParagraph",
      linkLabelKey: "definirAllLink",
    },
    categories: [
      {
        titleKey: "definirCategoryRegles",
        links: [
          { labelKey: "principes", href: "/definir/principes" },
          { labelKey: "standards", href: "/definir/standards" },
          { labelKey: "architectureReference", href: "/definir/architecture-de-reference" },
          { labelKey: "gouvernanceTechnique", href: "/definir/gouvernance-technique" },
        ],
      },
      {
        titleKey: "definirCategoryExigences",
        links: [
          { labelKey: "securite", href: "/definir/securite" },
          { labelKey: "accessibilite", href: "/definir/accessibilite" },
          { labelKey: "interoperabilite", href: "/definir/interoperabilite" },
          { labelKey: "souveraineteNumerique", href: "/definir/souverainete-numerique" },
        ],
      },
      {
        titleKey: "definirCategoryMatieres",
        links: [
          { labelKey: "donnees", href: "/definir/donnees" },
          { labelKey: "openSource", href: "/definir/open-source" },
        ],
      },
    ],
  },
  {
    labelKey: "concevoir",
    href: themePaths.concevoir,
    leader: {
      paragraphKey: "concevoirParagraph",
      linkLabelKey: "concevoirAllLink",
    },
    categories: [
      {
        titleKey: "concevoirCategoryServices",
        links: [
          { labelKey: "conceptionServices", href: "/concevoir/conception-de-services" },
          { labelKey: "parcoursUtilisateur", href: "/concevoir/parcours-utilisateur" },
          { labelKey: "modelisation", href: "/concevoir/modelisation" },
          { labelKey: "choixTechnologiques", href: "/concevoir/choix-technologiques" },
        ],
      },
      {
        titleKey: "concevoirCategoryArchitecture",
        links: [
          { labelKey: "architecture", href: "/concevoir/architecture" },
          { labelKey: "apis", href: "/concevoir/apis" },
          { labelKey: "donnees", href: "/concevoir/donnees" },
          { labelKey: "identite", href: "/concevoir/identite" },
        ],
      },
      {
        titleKey: "concevoirCategoryQualite",
        links: [
          { labelKey: "resilience", href: "/concevoir/resilience" },
          { labelKey: "referencesConception", href: "/concevoir/references-de-conception" },
        ],
      },
    ],
  },
  {
    labelKey: "construire",
    href: themePaths.construire,
    leader: {
      paragraphKey: "construireParagraph",
      linkLabelKey: "construireAllLink",
    },
    categories: [
      {
        titleKey: "construireCategoryDeveloppement",
        links: [
          { labelKey: "frontend", href: "/construire/frontend" },
          { labelKey: "backend", href: "/construire/backend" },
          { labelKey: "apis", href: "/construire/apis" },
          { labelKey: "basesDeDonnees", href: "/construire/bases-de-donnees" },
        ],
      },
      {
        titleKey: "construireCategoryReutilisables",
        links: [
          { labelKey: "packages", href: "/construire/packages" },
          { labelKey: "sdk", href: "/construire/sdk" },
          { labelKey: "cli", href: "/construire/cli" },
          { labelKey: "ads", href: "/construire/ads" },
        ],
      },
      {
        titleKey: "construireCategoryLivraison",
        links: [
          { labelKey: "tests", href: "/construire/tests" },
          { labelKey: "ciCd", href: "/construire/ci-cd" },
        ],
      },
      {
        titleKey: "construireCategoryInfrastructure",
        links: [
          { labelKey: "infrastructureAsCode", href: "/construire/infrastructure-as-code" },
          { labelKey: "containers", href: "/construire/containers" },
        ],
      },
    ],
  },
  {
    labelKey: "deployer",
    href: themePaths.deployer,
    leader: {
      paragraphKey: "deployerParagraph",
      linkLabelKey: "deployerAllLink",
    },
    categories: [
      {
        titleKey: "deployerCategoryInfrastructure",
        links: [
          { labelKey: "infrastructure", href: "/deployer/infrastructure" },
          { labelKey: "cloud", href: "/deployer/cloud" },
          { labelKey: "reseaux", href: "/deployer/reseaux" },
          { labelKey: "dns", href: "/deployer/dns" },
          { labelKey: "environnements", href: "/deployer/environnements" },
        ],
      },
      {
        titleKey: "deployerCategoryLivraison",
        links: [
          { labelKey: "ciCd", href: "/deployer/ci-cd" },
          { labelKey: "containers", href: "/deployer/containers" },
          { labelKey: "securite", href: "/deployer/securite" },
        ],
      },
      {
        titleKey: "deployerCategoryRobustesse",
        links: [
          { labelKey: "hauteDisponibilite", href: "/deployer/haute-disponibilite" },
          { labelKey: "scalabilite", href: "/deployer/scalabilite" },
          { labelKey: "disasterRecovery", href: "/deployer/disaster-recovery" },
        ],
      },
    ],
  },
  {
    labelKey: "exploiter",
    href: themePaths.exploiter,
    leader: {
      paragraphKey: "exploiterParagraph",
      linkLabelKey: "exploiterAllLink",
    },
    categories: [
      {
        titleKey: "exploiterCategoryObservabilite",
        links: [
          { labelKey: "monitoring", href: "/exploiter/monitoring" },
          { labelKey: "observabilite", href: "/exploiter/observabilite" },
          { labelKey: "logs", href: "/exploiter/logs" },
          { labelKey: "metriques", href: "/exploiter/metriques" },
          { labelKey: "traces", href: "/exploiter/traces" },
        ],
      },
      {
        titleKey: "exploiterCategorySupervision",
        links: [
          { labelKey: "alerting", href: "/exploiter/alerting" },
          { labelKey: "incidents", href: "/exploiter/incidents" },
          { labelKey: "performance", href: "/exploiter/performance" },
          { labelKey: "slaSlo", href: "/exploiter/sla-slo" },
        ],
      },
      {
        titleKey: "exploiterCategoryContinuite",
        links: [
          { labelKey: "maintenance", href: "/exploiter/maintenance" },
          { labelKey: "continuiteService", href: "/exploiter/continuite-de-service" },
        ],
      },
    ],
  },
  {
    labelKey: "faireEvoluer",
    href: themePaths.faireEvoluer,
    leader: {
      paragraphKey: "evoluerParagraph",
      linkLabelKey: "evoluerAllLink",
    },
    categories: [
      {
        titleKey: "evoluerCategoryContribuer",
        links: [
          { labelKey: "contributions", href: "/faire-evoluer/contributions" },
          { labelKey: "rfc", href: "/faire-evoluer/rfc" },
          { labelKey: "openSource", href: "/faire-evoluer/open-source" },
          { labelKey: "communaute", href: "/faire-evoluer/communaute" },
        ],
      },
      {
        titleKey: "evoluerCategoryEcosysteme",
        links: [
          { labelKey: "roadmap", href: "/faire-evoluer/roadmap" },
          { labelKey: "retoursExperience", href: "/faire-evoluer/retours-experience" },
          { labelKey: "innovation", href: "/faire-evoluer/innovation" },
          { labelKey: "recherche", href: "/faire-evoluer/recherche" },
        ],
      },
      {
        titleKey: "evoluerCategoryCycleVie",
        links: [
          { labelKey: "versions", href: "/faire-evoluer/versions" },
          { labelKey: "migrations", href: "/faire-evoluer/migrations" },
          { labelKey: "depreciations", href: "/faire-evoluer/depreciations" },
        ],
      },
    ],
  },
];

/* -------------------------------------------------------------------------- *
 * Header navigation — data model
 * -------------------------------------------------------------------------- */

/** One row of a mega-menu column. The label resolves under `nav.panel` by
 *  default. `external` rows open in a new window. `primary` labels resolve
 *  under `nav.primary` instead (lifecycle names). */
export type HeaderNavLink = {
  labelKey: string;
  href: string;
  external?: boolean;
  primary?: boolean;
};

/** One column of a header mega-menu. The heading is a `nav.panel` (or
 *  `nav.primary` when `titlePrimary`) message key. When `href` is set the
 *  heading is itself a destination (ADS `categoryMainLink`). */
export type HeaderNavCategory = {
  titleKey: string;
  titlePrimary?: boolean;
  href?: string;
  external?: boolean;
  links: ReadonlyArray<HeaderNavLink>;
};

/** A direct-link entry of the header (Accueil, Actualités). Its label is a
 *  `nav.primary` key; next-intl prefixes the active locale to internal `href`
 *  automatically. `external` entries open in a new window. */
export type HeaderNavLinkEntry = {
  kind: "link";
  labelKey: string;
  href: string;
  external?: boolean;
  /** Hrefs considered when marking the entry current. Defaults to `[href]`. */
  matchHrefs?: ReadonlyArray<string>;
};

/** A mega-menu entry of the header (Communauté, Documentation, Ressources,
 *  À propos). A mega-menu has no leader band (no title nor description): it
 *  only exposes exactly four sections, each with at least five links, and
 *  the “Fermer” button rendered by the ADS MegaMenu. The top-level label is
 *  a `nav.primary` key. */
export type HeaderNavMegaEntry = {
  kind: "megaMenu";
  labelKey: string;
  categories: ReadonlyArray<HeaderNavCategory>;
  /** Hrefs considered when marking the section current. When absent, every
   *  href reachable from the panel is considered. Sections whose panel is a
   *  set of shortcuts (Actualités) pass an explicit list to stay neutral. */
  matchHrefs?: ReadonlyArray<string>;
};

/** One entry of the header navigation bar. */
export type HeaderNavEntry = HeaderNavLinkEntry | HeaderNavMegaEntry;

/**
 * Navigation of the CODE.GOUV.AOR header. It answers « Que puis-je trouver
 * sur CODE.GOUV.AOR ? » with six entries on the bar:
 *
 *   Accueil       — simple link to the homepage (the portal entry point).
 *   Actualités    — simple link to the news channel of the platform (the
 *                   public repository releases feed, external).
 *   Communauté    — mega-menu (4 sections): who builds public digital
 *                   technology, and how to take part.
 *   Documentation — mega-menu (4 sections): « Comment construire le numérique
 *                   public ? », with the six lifecycle themes (Définir → … →
 *                   Faire évoluer) grouped into the four sections.
 *   Ressources    — mega-menu (4 sections): the concrete building blocks.
 *   À propos      — mega-menu (4 sections): institutional pages.
 *
 * Each of the four mega-menus exposes exactly four sections, and every
 * section carries at least five links. All of it is defined explicitly in
 * the `categories` arrays below.
 *
 * Extensibility rule: entries only point to destinations that exist or are
 * canonical routes of the platform. Families without a page yet (BlueHats 🧢,
 * Discussions, Événements, Proposer une ressource, Gouvernance de CODE,
 * Contact…) are intentionally absent and are added here — nothing else — as
 * soon as their route ships. The header columns below are organised in a
 * data-first way so that extending a panel never requires touching the
 * Header component.
 */
export const headerNavigation: ReadonlyArray<HeaderNavEntry> = [
  {
    kind: "link",
    labelKey: "home",
    href: PORTAL_HOME,
  },
  {
    kind: "link",
    labelKey: "actualites",
    // No editorial page exists yet: the news channel of the platform is the
    // public repository releases feed (external link, new window).
    href: codeRepositoryReleasesUrl,
    external: true,
  },
  {
    kind: "megaMenu",
    labelKey: "communaute",
    // 4 sections, each with at least five links.
    categories: [
      {
        titleKey: "contribuer",
        links: [
          { labelKey: "contributions", href: "/faire-evoluer/contributions" },
          { labelKey: "openSource", href: "/faire-evoluer/open-source" },
          { labelKey: "packages", href: "/construire/packages" },
          { labelKey: "sdk", href: "/construire/sdk" },
          { labelKey: "cli", href: "/construire/cli" },
          { labelKey: "ads", href: "/construire/ads" },
        ],
      },
      {
        titleKey: "communauteCategoryProposer",
        links: [
          { labelKey: "rfc", href: "/faire-evoluer/rfc" },
          { labelKey: "standards", href: portalPaths.standards },
          { labelKey: "gouvernanceTechnique", href: "/definir/gouvernance-technique" },
          { labelKey: "innovation", href: "/faire-evoluer/innovation" },
          { labelKey: "recherche", href: "/faire-evoluer/recherche" },
        ],
      },
      {
        titleKey: "communauteCategoryEvolutions",
        links: [
          { labelKey: "roadmap", href: "/faire-evoluer/roadmap" },
          { labelKey: "versions", href: "/faire-evoluer/versions" },
          { labelKey: "migrations", href: "/faire-evoluer/migrations" },
          { labelKey: "depreciations", href: "/faire-evoluer/depreciations" },
          { labelKey: "retoursExperience", href: "/faire-evoluer/retours-experience" },
        ],
      },
      {
        titleKey: "decouvrir",
        links: [
          { labelKey: "communaute", href: portalPaths.communaute },
          { labelKey: "projets", href: portalPaths.projets },
          { labelKey: "github", href: codeRepositoryUrl, external: true },
          { labelKey: "releases", href: codeRepositoryReleasesUrl, external: true },
          { labelKey: "openData", href: openDataUrl, external: true },
        ],
      },
    ],
  },
  {
    kind: "megaMenu",
    labelKey: "documentation",
    // 4 sections, each with at least five links. The lifecycle themes are
    // grouped into four stages so the bar keeps one balanced Documentation
    // mega-menu instead of six theme entries.
    categories: [
      {
        titleKey: "definir",
        titlePrimary: true,
        href: themePaths.definir,
        links: [
          { labelKey: "principes", href: "/definir/principes" },
          { labelKey: "standards", href: "/definir/standards" },
          { labelKey: "architectureReference", href: "/definir/architecture-de-reference" },
          { labelKey: "gouvernanceTechnique", href: "/definir/gouvernance-technique" },
          { labelKey: "securite", href: "/definir/securite" },
          { labelKey: "accessibilite", href: "/definir/accessibilite" },
          { labelKey: "interoperabilite", href: "/definir/interoperabilite" },
          { labelKey: "souveraineteNumerique", href: "/definir/souverainete-numerique" },
          { labelKey: "donnees", href: "/definir/donnees" },
          { labelKey: "openSource", href: "/definir/open-source" },
        ],
      },
      {
        titleKey: "concevoir",
        titlePrimary: true,
        href: themePaths.concevoir,
        links: [
          { labelKey: "conceptionServices", href: "/concevoir/conception-de-services" },
          { labelKey: "parcoursUtilisateur", href: "/concevoir/parcours-utilisateur" },
          { labelKey: "modelisation", href: "/concevoir/modelisation" },
          { labelKey: "choixTechnologiques", href: "/concevoir/choix-technologiques" },
          { labelKey: "architecture", href: "/concevoir/architecture" },
          { labelKey: "apis", href: "/concevoir/apis" },
          { labelKey: "identite", href: "/concevoir/identite" },
          { labelKey: "resilience", href: "/concevoir/resilience" },
          { labelKey: "referencesConception", href: "/concevoir/references-de-conception" },
        ],
      },
      {
        titleKey: "construire",
        titlePrimary: true,
        href: themePaths.construire,
        links: [
          { labelKey: "frontend", href: "/construire/frontend" },
          { labelKey: "backend", href: "/construire/backend" },
          { labelKey: "basesDeDonnees", href: "/construire/bases-de-donnees" },
          { labelKey: "tests", href: "/construire/tests" },
          { labelKey: "ciCd", href: "/construire/ci-cd" },
          { labelKey: "infrastructureAsCode", href: "/construire/infrastructure-as-code" },
          { labelKey: "containers", href: "/construire/containers" },
        ],
      },
      {
        titleKey: "documentationCategoryCycle",
        links: [
          { labelKey: "deployer", href: themePaths.deployer, primary: true },
          { labelKey: "exploiter", href: themePaths.exploiter, primary: true },
          { labelKey: "faireEvoluer", href: themePaths.faireEvoluer, primary: true },
          { labelKey: "infrastructure", href: "/deployer/infrastructure" },
          { labelKey: "cloud", href: "/deployer/cloud" },
          { labelKey: "monitoring", href: "/exploiter/monitoring" },
          { labelKey: "observabilite", href: "/exploiter/observabilite" },
          { labelKey: "incidents", href: "/exploiter/incidents" },
          { labelKey: "slaSlo", href: "/exploiter/sla-slo" },
          { labelKey: "versions", href: "/faire-evoluer/versions" },
        ],
      },
    ],
  },
  {
    kind: "megaMenu",
    labelKey: "ressources",
    // 4 sections, each with at least five links.
    categories: [
      {
        titleKey: "design",
        links: [
          { labelKey: "ads", href: "/construire/ads" },
          { labelKey: "frontend", href: "/construire/frontend" },
          { labelKey: "accessibilite", href: "/definir/accessibilite" },
          { labelKey: "referencesConception", href: "/concevoir/references-de-conception" },
          { labelKey: "parcoursUtilisateur", href: "/concevoir/parcours-utilisateur" },
        ],
      },
      {
        titleKey: "developpement",
        links: [
          { labelKey: "backend", href: "/construire/backend" },
          { labelKey: "basesDeDonnees", href: "/construire/bases-de-donnees" },
          { labelKey: "packages", href: "/construire/packages" },
          { labelKey: "sdk", href: "/construire/sdk" },
          { labelKey: "cli", href: "/construire/cli" },
          { labelKey: "templates", href: "/construire/templates" },
        ],
      },
      {
        titleKey: "infrastructure",
        links: [
          { labelKey: "cloud", href: "/deployer/cloud" },
          { labelKey: "reseaux", href: "/deployer/reseaux" },
          { labelKey: "dns", href: "/deployer/dns" },
          { labelKey: "environnements", href: "/deployer/environnements" },
          { labelKey: "containers", href: "/deployer/containers" },
          { labelKey: "infrastructureAsCode", href: "/construire/infrastructure-as-code" },
        ],
      },
      {
        titleKey: "apisEtDonnees",
        links: [
          { labelKey: "apis", href: "/construire/apis" },
          { labelKey: "donnees", href: "/definir/donnees" },
          { labelKey: "openData", href: openDataUrl, external: true },
          { labelKey: "interoperabilite", href: "/definir/interoperabilite" },
          { labelKey: "souveraineteNumerique", href: "/definir/souverainete-numerique" },
        ],
      },
    ],
  },
  {
    kind: "megaMenu",
    labelKey: "aPropos",
    // 4 sections, each with at least five links.
    categories: [
      {
        titleKey: "aProposStandards",
        links: [
          { labelKey: "standards", href: portalPaths.standards },
          { labelKey: "architectureReference", href: "/definir/architecture-de-reference" },
          { labelKey: "gouvernanceTechnique", href: "/definir/gouvernance-technique" },
          { labelKey: "interoperabilite", href: "/definir/interoperabilite" },
          { labelKey: "securite", href: "/definir/securite" },
          { labelKey: "accessibilite", href: "/definir/accessibilite" },
        ],
      },
      {
        titleKey: "aProposOpenSource",
        links: [
          { labelKey: "openSource", href: "/definir/open-source" },
          { labelKey: "github", href: codeRepositoryUrl, external: true },
          { labelKey: "releases", href: codeRepositoryReleasesUrl, external: true },
          { labelKey: "openData", href: openDataUrl, external: true },
          { labelKey: "projets", href: portalPaths.projets },
        ],
      },
      {
        titleKey: "aProposEvolutions",
        links: [
          { labelKey: "rfc", href: "/faire-evoluer/rfc" },
          { labelKey: "roadmap", href: "/faire-evoluer/roadmap" },
          { labelKey: "versions", href: "/faire-evoluer/versions" },
          { labelKey: "migrations", href: "/faire-evoluer/migrations" },
          { labelKey: "depreciations", href: "/faire-evoluer/depreciations" },
          { labelKey: "retoursExperience", href: "/faire-evoluer/retours-experience" },
        ],
      },
      {
        titleKey: "aProposPlus",
        links: [
          { labelKey: "documentation", href: portalPaths.documentation },
          { labelKey: "ressources", href: portalPaths.ressources },
          { labelKey: "communaute", href: portalPaths.communaute },
          { labelKey: "contributions", href: "/faire-evoluer/contributions" },
          { labelKey: "innovation", href: "/faire-evoluer/innovation" },
          { labelKey: "recherche", href: "/faire-evoluer/recherche" },
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
