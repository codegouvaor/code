/**
 * URL structure of the public portal.
 *
 * Hrefs are locale-agnostic pathnames: the next-intl Link (registered as the
 * ADS link renderer) prefixes the active locale automatically. Labels are
 * never stored here — they come from the message catalogs through the key
 * provided by each entry.
 */
export const PORTAL_HOME = "/";

export type PrimaryNavKey =
  | "gouvernement"
  | "actionPublique"
  | "servicesPublics"
  | "actualites"
  | "republique"
  | "informationsUtiles";

/** A link inside a mega-menu panel; its label is a `nav.panel` message key. */
export type PrimaryNavLink = {
  labelKey: string;
  href: string;
};

/**
 * One column of a mega-menu panel. The heading is either plain text
 * (`titleKey`) or, when the column heading is itself the main destination of
 * the group, a link (`mainLink`) — rendered by the ADS `MegaMenu` as
 * `categoryMainText` / `categoryMainLink`.
 */
export type MegaMenuCategory =
  | {
      /** Message key (`nav.panel`) of the category heading. */
      titleKey: string;
      mainLink?: never;
      links: ReadonlyArray<PrimaryNavLink>;
    }
  | {
      titleKey?: never;
      /** Heading rendered as a link (e.g. “Toutes les démarches”). */
      mainLink: PrimaryNavLink;
      links: ReadonlyArray<PrimaryNavLink>;
    };

/**
 * One top-level entry of the Government Header navigation.
 *
 * Navigation principle (info.gouv.fr-inspired, adapted to Astoria): every
 * section opens a mega-menu panel composed of
 *  - a leader band: the section name, a one-line description and the main
 *    section action (e.g. “→ Tout le Gouvernement”),
 *  - an optional featured zone (used by Actualités for “À la une”),
 *  - a small number of link columns — the mega-menu is not the sitemap of the
 *    portal: only the destinations that matter to the user journey.
 *
 * Top-level labels resolve under `nav.primary` (`labelKey`), panel content
 * under `nav.panel` (`titleKey`, `paragraphKey` and nested `labelKey`s).
 */
export type PrimaryNavItem =
  | {
      type: "link";
      labelKey: string;
      href: string;
    }
  | {
      type: "megaMenu";
      labelKey: string;
      href: string;
      /** Leader band shown on top of the panel. */
      leader: {
        titleKey: string;
        paragraphKey: string;
        link: PrimaryNavLink;
      };
      /**
       * Optional featured zone rendered as the first column of the panel
       * (e.g. “À la une” in Actualités), meant to be fed dynamically later.
       * The link label is NOT stored here: the header reads the shared
       * `home.news.featured` messages so the headline has a single source.
       */
      featuredLink?: {
        titleKey: string;
        href: string;
      };
      categories?: ReadonlyArray<MegaMenuCategory>;
    };

export type FooterColumn = {
  /** Message key (`footer.columns`) of the column heading. */
  columnKey: string;
  links: ReadonlyArray<PrimaryNavLink>;
};

export const sectionPaths = {
  composition: "/government/composition",
  decryptages: "/decryptages",
  lEtatEtMoi: "/l-etat-et-moi",
  preventionDesRisques: "/prevention-des-risques",
  suiviDesEngagements: "/suivi-des-engagements",
  liensUtiles: "/liens-utiles",
  /** Citizen participation: provisional section being published. */
  participation: "/participation",
  /** Priority policies (former “Politiques prioritaires” navigation entry). */
  politiquesPrioritaires: "/politiques-prioritaires",
  /** Public policies hub — main destination of “L'action publique”. */
  politiquesPubliques: "/politiques-publiques",
} as const;

export const legalPaths = {
  accessibility: "/legal/accessibility",
  privacy: "/legal/privacy",
  terms: "/legal/terms",
  cookies: "/legal/cookies",
  sitemap: "/sitemap",
} as const;

export const searchPath = "/search";

/** Press area of the portal (dedicated page being published). */
export const pressPath = "/presse";

/** DOM ids used as skip-link targets. */
export const pageAnchors = {
  content: "main-content",
  footer: "main-footer",
} as const;

/**
 * Link columns of the Government Footer, mirroring info.gouv.fr:
 * Actualités, Grands dossiers, Prévenir les risques, Outils, L'État et moi.
 * Labels resolve under `nav.panel`, column titles under `footer.columns`.
 */
export const footerNavigation: ReadonlyArray<FooterColumn> = [
  {
    columnKey: "actualites",
    links: [
      { labelKey: "allNews", href: "/news" },
      {
        labelKey: "actualitePremierMinistre",
        href: "/news/actualite-du-premier-ministre",
      },
      { labelKey: "budget", href: "/news/budget" },
      { labelKey: "sante", href: "/news/sante" },
      { labelKey: "ceQuiChange", href: "/news/ce-qui-change" },
      { labelKey: "vosQuestions", href: "/news/vos-questions-nos-reponses" },
      { labelKey: "videos", href: "/news/videos" },
      { labelKey: "podcasts", href: "/news/podcasts" },
      { labelKey: "lettresInfo", href: "/news/newsletters" },
      { labelKey: "articlesAudio", href: "/news/articles-audio" },
    ],
  },
  {
    columnKey: "grandsDossiers",
    links: [
      { labelKey: "parlonsSanteMentale", href: "/news/parlons-sante-mentale" },
      { labelKey: "toutesEtTousEgaux", href: "/news/toutes-et-tous-egaux" },
      { labelKey: "relanceLogement", href: "/news/relance-logement" },
      { labelKey: "republiqueGrandAngle", href: "/news/la-republique-en-grand-angle" },
      { labelKey: "astoria2030", href: "/news/astoria-2030" },
      { labelKey: "astoriaNationVerte", href: "/news/astoria-nation-verte" },
      { labelKey: "maisonsServicesPublics", href: "/news/maisons-services-publics" },
      { labelKey: "tousLesGrandsDossiers", href: "/news/grands-dossiers" },
    ],
  },
  {
    columnKey: "prevenirLesRisques",
    links: [
      {
        labelKey: "sePreparerUrgenceTitle",
        href: "/prevention-des-risques/se-preparer-a-une-situation-durgence",
      },
      {
        labelKey: "risquesNaturels",
        href: "/prevention-des-risques/risques-naturels-et-technologiques",
      },
      { labelKey: "risquesEpidemiques", href: "/prevention-des-risques/risques-epidemiques" },
      { labelKey: "menaceTerroriste", href: "/prevention-des-risques/menace-terroriste" },
      { labelKey: "menaceCyber", href: "/prevention-des-risques/menace-cyber" },
    ],
  },
  {
    columnKey: "outils",
    links: [
      { labelKey: "portailAccessibilite", href: legalPaths.accessibility },
      { labelKey: "marqueEtat", href: "/marque-de-letat" },
      { labelKey: "devenirPartenaireEtat", href: "/devenir-partenaire-de-letat" },
      { labelKey: "liensUtilesDemarches", href: sectionPaths.liensUtiles },
      // Kept reachable from the footer after leaving the main navigation
      // (info.gouv.fr keeps its bar to 6 items).
      { labelKey: "suiviEngagementsFooter", href: sectionPaths.suiviDesEngagements },
    ],
  },
  {
    columnKey: "lEtatEtMoi",
    links: [
      { labelKey: "homeLink", href: sectionPaths.lEtatEtMoi },
      { labelKey: "etatOrganisateur", href: "/l-etat-et-moi/etat-organisateur" },
      { labelKey: "etatProtecteur", href: "/l-etat-et-moi/etat-protecteur" },
      { labelKey: "etatFacilitateur", href: "/l-etat-et-moi/etat-facilitateur" },
      { labelKey: "etatMoteur", href: "/l-etat-et-moi/etat-moteur" },
      { labelKey: "etatPromoteur", href: "/l-etat-et-moi/etat-promoteur" },
      { labelKey: "etatInfluent", href: "/l-etat-et-moi/etat-influent" },
    ],
  },
];

/**
 * Main navigation of the Government Header — the permanent architecture of
 * the portal, in six entries. Each entry answers one user intention:
 *
 *   Le Gouvernement      → Qui gouverne et comment fonctionne le Gouvernement ?
 *   L'action publique    → Que fait la République ?
 *   Services publics     → Que puis-je faire avec l'État ?
 *   Actualités           → Que se passe-t-il actuellement ?
 *   La République        → Comment fonctionne l'État ?
 *   Informations utiles  → Où trouver une information pratique ou importante ?
 *
 * The former entries are repositioned, not deleted (their URLs stay alive):
 * À la une / Décryptages → Actualités, Politiques prioritaires →
 * L'action publique, L'État et moi → Services publics, Prévention des
 * risques / Liens utiles → Informations utiles.
 *
 * Hrefs follow the URL plan already used across the portal; a few point to
 * pages being published (e.g. /politiques-publiques/<slug>, /republique/…)
 * and will resolve as soon as those sections ship.
 */
export const primaryNavigation: ReadonlyArray<PrimaryNavItem> = [
  {
    type: "megaMenu",
    labelKey: "gouvernement",
    href: "/government",
    leader: {
      titleKey: "gouvernementTitle",
      paragraphKey: "gouvernementText",
      link: { labelKey: "gouvernementAllLink", href: "/government" },
    },
    categories: [
      {
        // LE GOUVERNEMENT
        mainLink: { labelKey: "gouvernementCategory", href: "/government" },
        links: [
          { labelKey: "gouvernementComposition", href: "/government/composition" },
          { labelKey: "gouvernementMembres", href: "/government/membres" },
          { labelKey: "premierMinistre", href: "/government/liamvonastoria" },
          { labelKey: "gouvernementBiographies", href: "/government/biographies" },
        ],
      },
      {
        // ORGANISATION
        mainLink: { labelKey: "gouvernementOrganisation", href: "/government/organisation" },
        links: [
          { labelKey: "conseilMinistres", href: "/government/conseil-des-ministres" },
          { labelKey: "ministeres", href: "/government/ministere" },
          { labelKey: "gouvernementSecretariats", href: "/government/organisation/secretariats" },
          {
            labelKey: "administrationsRattachees",
            href: "/government/organisation/administrations-rattachees",
          },
          {
            labelKey: "serviceInformationGouvernement",
            href: "/government/sig",
          },
        ],
      },
      {
        // ACTION DU GOUVERNEMENT
        mainLink: { labelKey: "politiquesPubliques", href: sectionPaths.politiquesPubliques },
        links: [
          { labelKey: "prioritesGouvernement", href: sectionPaths.politiquesPrioritaires },
          { labelKey: "grandsDossiers", href: "/news/grands-dossiers" },
          { labelKey: "reformes", href: "/reformes" },
          { labelKey: "ceQuiChange", href: "/news/ce-qui-change" },
          { labelKey: "resultatsActionPublique", href: sectionPaths.suiviDesEngagements },
        ],
      },
      {
        // AGENDA & TRAVAUX
        mainLink: { labelKey: "agenda", href: "/agenda" },
        links: [
          { labelKey: "decisions", href: "/decisions" },
          { labelKey: "discours", href: "/discours-et-rapports" },
          { labelKey: "communiques", href: "/communiques" },
          { labelKey: "publications", href: "/publications-officielles" },
        ],
      },
    ],
  },
  {
    type: "megaMenu",
    labelKey: "actionPublique",
    href: sectionPaths.politiquesPubliques,
    leader: {
      titleKey: "actionPubliqueTitle",
      paragraphKey: "actionPubliqueText",
      link: { labelKey: "actionPubliqueAllLink", href: sectionPaths.politiquesPubliques },
    },
    categories: [
      {
        // POLITIQUES PUBLIQUES — the permanent policy areas
        mainLink: { labelKey: "actionPubliquePolitiquesCategory", href: sectionPaths.politiquesPubliques },
        links: [
          { labelKey: "politiqueEconomie", href: "/politiques-publiques/economie" },
          { labelKey: "politiqueEducation", href: "/politiques-publiques/education" },
          { labelKey: "politiqueSante", href: "/politiques-publiques/sante" },
          { labelKey: "politiqueSecurite", href: "/politiques-publiques/securite" },
          { labelKey: "politiqueNumerique", href: "/politiques-publiques/numerique" },
          { labelKey: "politiqueEnvironnement", href: "/politiques-publiques/environnement" },
          { labelKey: "politiqueMobilite", href: "/politiques-publiques/mobilite" },
          { labelKey: "politiqueCulture", href: "/politiques-publiques/culture" },
        ],
      },
      {
        // PRIORITÉS
        mainLink: { labelKey: "prioritesGouvernement", href: sectionPaths.politiquesPrioritaires },
        links: [
          { labelKey: "grandsDossiers", href: "/news/grands-dossiers" },
          { labelKey: "reformes", href: "/government/reformes" },
        ],
      },
      {
        // COMPRENDRE L'ACTION PUBLIQUE
        titleKey: "actionPubliqueComprendreCategory",
        links: [
          { labelKey: "ceQuiChange", href: "/news/ce-qui-change" },
          { labelKey: "resultatsActionPublique", href: sectionPaths.suiviDesEngagements },
          { labelKey: "evaluationPolitiquesPubliques", href: "/politiques-publiques/evaluation" },
        ],
      },
      {
        // RESSOURCES
        titleKey: "actionPubliqueRessourcesCategory",
        links: [
          { labelKey: "publications", href: "/publications-officielles" },
          { labelKey: "rapports", href: "/discours-et-rapports" },
          { labelKey: "donneesOuvertes", href: "https://data.gouv.aor/" },
        ],
      },
    ],
  },
  {
    type: "megaMenu",
    labelKey: "servicesPublics",
    href: "/services",
    leader: {
      titleKey: "servicesTitle",
      paragraphKey: "servicesText",
      // The services hub is the entry point to search for a “démarche”.
      link: { labelKey: "servicesRechercheLink", href: "/services" },
    },
    categories: [
      {
        // DÉMARCHES
        mainLink: { labelKey: "toutesLesDemarches", href: "/services" },
        links: [
          { labelKey: "demarcheIdentite", href: "/services/demarches/identite" },
          { labelKey: "demarcheFiscalite", href: "/services/demarches/fiscalite" },
          { labelKey: "demarcheLogement", href: "/services/demarches/logement" },
          { labelKey: "demarcheTransport", href: "/services/demarches/transport" },
          { labelKey: "demarcheEducation", href: "/services/demarches/education" },
          { labelKey: "demarcheSante", href: "/services/demarches/sante" },
        ],
      },
      {
        // PAR SITUATION
        titleKey: "servicesSituationsCategory",
        links: [
          { labelKey: "situationCitoyen", href: "/services/citoyen" },
          { labelKey: "situationEtudiant", href: "/services/etudiant" },
          { labelKey: "situationProfessionnel", href: "/services/professionnel" },
          { labelKey: "situationEntreprise", href: "/services/entreprise" },
          { labelKey: "situationAssociation", href: "/services/association" },
          { labelKey: "situationEtranger", href: "/services/etranger" },
        ],
      },
      {
        // SERVICES
        titleKey: "servicesServicesCategory",
        links: [
          { labelKey: "administrations", href: "/organisation/administrations" },
          { labelKey: "servicesEnLigne", href: "/services/services-en-ligne" },
        ],
      },
    ],
  },
  {
    type: "megaMenu",
    labelKey: "actualites",
    href: "/news",
    leader: {
      titleKey: "actualitesTitle",
      paragraphKey: "actualitesText",
      link: { labelKey: "actualitesAllLink", href: "/news" },
    },
    // “À la une” — the featured news zone, first column of the panel.
    // Keep in sync with `featuredArticle` in `lib/home-content.ts`; feed it
    // from the editorial API when it becomes available.
    featuredLink: {
      titleKey: "actualitesUneCategory",
      // Keep in sync with `featuredArticle` (lib/home-content.ts).
      href: "/news/ce-qui-change",
    },
    categories: [
      {
        // ACTUALITÉS
        titleKey: "actualitesCategory",
        links: [
          { labelKey: "allNews", href: "/news" },
          { labelKey: "decryptages", href: sectionPaths.decryptages },
          { labelKey: "communiques", href: "/communiques" },
          { labelKey: "discours", href: "/discours-et-rapports" },
          { labelKey: "conferencesPresse", href: "/actualites/conferences-de-presse" },
          { labelKey: "videos", href: "/news/videos" },
        ],
      },
      {
        // THÉMATIQUES
        titleKey: "thematiquesNewsCategory",
        links: [
          { labelKey: "politiqueEconomie", href: "/news/budget" },
          { labelKey: "politiqueEducation", href: "/news/rentree-scolaire" },
          { labelKey: "politiqueSante", href: "/news/sante" },
          { labelKey: "politiqueSecurite", href: "/news/narcotrafic" },
          { labelKey: "politiqueNumerique", href: "/news/intelligence-artificielle" },
          { labelKey: "politiqueEnvironnement", href: "/news/astoria-nation-verte" },
          { labelKey: "thematiqueSociete", href: "/news/laicite" },
        ],
      },
    ],
  },
  {
    type: "megaMenu",
    labelKey: "republique",
    href: "/republique",
    leader: {
      titleKey: "republiqueTitle",
      paragraphKey: "republiqueText",
      link: { labelKey: "republiqueDiscoverLink", href: "/republique" },
    },
    categories: [
      {
        // INSTITUTIONS
        mainLink: { labelKey: "institutionsRepublic", href: "/republique" },
        links: [
          { labelKey: "constitution", href: "/republique/constitution" },
          { labelKey: "organisationEtat", href: "/republique/organisation" },
          { labelKey: "administrations", href: "/republique/administrations" },
          { labelKey: "autoritesPubliques", href: "/republique/autorites-publiques" },
          { labelKey: "territoiresCollectivites", href: "/republique/territoires" },
        ],
      },
      {
        // FONCTIONNEMENT
        titleKey: "republiqueFonctionnementCategory",
        links: [
          { labelKey: "fonctionnementInstitutions", href: "/republique/fonctionnement" },
          { labelKey: "relationsInstitutions", href: "/republique/relations-entre-institutions" },
          { labelKey: "servicePublicLink", href: "/republique/services" },
        ],
      },
      {
        // RÉPUBLIQUE OUVERTE
        titleKey: "republiqueOuverteCategory",
        links: [
          { labelKey: "transparence", href: "/republique/transparence" },
          { labelKey: "donneesOuvertes", href: "https://data.gouv.aor/" },
          { labelKey: "publications", href: "/republique/publications-officielles" },
          { labelKey: "openSource", href: "/republique/open-source" },
          { labelKey: "participationCitoyenne", href: sectionPaths.participation },
          { labelKey: "devenirVolontaire", href: "/republique/devenir-volontaire" },
        ],
      },
    ],
  },
  {
    type: "megaMenu",
    labelKey: "informationsUtiles",
    href: sectionPaths.liensUtiles,
    leader: {
      titleKey: "infosUtilesTitle",
      paragraphKey: "infosUtilesText",
      link: { labelKey: "infosUtilesAllLink", href: sectionPaths.liensUtiles },
    },
    categories: [
      {
        // PRÉVENTION & SÉCURITÉ
        titleKey: "infosPreventionCategory",
        links: [
          { labelKey: "risquesMajeurs", href: sectionPaths.preventionDesRisques },
          {
            labelKey: "sePreparerUrgenceTitle",
            href: "/prevention-des-risques/se-preparer-a-une-situation-durgence",
          },
          {
            labelKey: "risquesNaturels",
            href: "/prevention-des-risques/risques-naturels-et-technologiques",
          },
          { labelKey: "menaceTerroriste", href: "/prevention-des-risques/menace-terroriste" },
          { labelKey: "menaceCyber", href: "/prevention-des-risques/menace-cyber" },
        ],
      },
      {
        // INFORMATIONS PRATIQUES
        titleKey: "infosPratiquesCategory",
        links: [
          { labelKey: "contacts", href: "/contact" },
          { labelKey: "numerosUtiles", href: "/liens-utiles/numeros-utiles" },
          { labelKey: "accessibilite", href: legalPaths.accessibility },
          { labelKey: "questionsFrequentes", href: "/liens-utiles/faq" },
          // The official-sites directory stays reachable through its page.
          { labelKey: "liensUtiles", href: sectionPaths.liensUtiles },
        ],
      },
      {
        // LE PORTAIL
        titleKey: "infosPortailCategory",
        links: [
          { labelKey: "planDuSite", href: legalPaths.sitemap },
          { labelKey: "espacePresse", href: pressPath },
          { labelKey: "accessibilite", href: legalPaths.accessibility },
          { labelKey: "informationsLegales", href: legalPaths.terms },
          { labelKey: "confidentialite", href: legalPaths.privacy },
          { labelKey: "gestionCookies", href: legalPaths.cookies },
        ],
      },
    ],
  },
];
