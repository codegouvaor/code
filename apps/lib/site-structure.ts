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
 * NOTE: most destinations below are provisional. The platform content model
 * is still being built; each theme hub (`/definir`, `/concevoir`, …) and its
 * child resources are the URLs this platform will publish.
 */
export const PORTAL_HOME = "/";

/**
 * The six lifecycle themes — the permanent architecture of the platform.
 * They form a loop: DÉFINIR → CONCEVOIR → CONSTRUIRE → DÉPLOYER →
 * EXPLOITER → FAIRE ÉVOLUER → ↺. Each theme is one door into the platform
 * and answers a single user intention:
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
 * One of the six lifecycle themes. It feeds the header mega-menu (one menu
 * per theme); top-level label resolves under `nav.primary`, every other
 * label under `nav.panel`. The footer deliberately does not repeat the
 * themes — the header owns the lifecycle navigation.
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

/**
 * Platform-level destinations shown in the footer bottom bar: they answer
 * “what can I do on CODE?” (documentation, catalogs, projects, community)
 * and give access to the source repository.
 *
 * Labels resolve under `nav.panel`. `github` points outside the platform and
 * is rendered as an external link by the footer.
 */
export const footerPlatformLinks: ReadonlyArray<PrimaryNavLink> = [
  { labelKey: "documentation", href: "/documentation" },
  { labelKey: "standards", href: "/standards" },
  { labelKey: "ressources", href: "/ressources" },
  { labelKey: "projets", href: "/projets" },
  { labelKey: "communaute", href: "/communaute" },
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
