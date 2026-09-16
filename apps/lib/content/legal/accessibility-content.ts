import type { ArticleContent } from "@/lib/theme-localize";
import { legalRelatedDestinations } from "@/lib/content/legal/legal-related";

/**
 * Content configuration of the accessibility declaration page
 * (`/legal/accessibility`).
 *
 * Structure only: every display string is resolved from the message catalogs
 * (`pages.legal.accessibility.*` and `pages.legal.related.*`) by
 * `lib/theme-localize.ts`.
 */
export const accessibilityContent: ArticleContent = {
  hero: {
    kickerKey: "hero.kicker",
    titleKey: "hero.title",
    leadKey: "hero.lead",
    noticeKey: "hero.notice",
  },
  sections: [
    {
      key: "declaration",
      id: "declaration-engagement",
      lead: true,
      paragraphCount: 2,
    },
    {
      key: "conformite",
      id: "etat-de-conformite",
      subtle: true,
      lead: true,
      bulletCount: 3,
    },
    {
      key: "nonAccessibles",
      id: "contenus-non-accessibles",
      lead: true,
      bulletCount: 3,
    },
    {
      key: "etablissement",
      id: "etablissement-de-cette-declaration",
      subtle: true,
      lead: true,
      facts: [{ key: "tests" }, { key: "referentiel" }, { key: "outils" }],
    },
    {
      key: "retour",
      id: "retour-d-information-et-contact",
      lead: true,
      paragraphCount: 1,
    },
    {
      key: "voiesRecours",
      id: "voies-de-recours",
      subtle: true,
      lead: true,
      bulletCount: 3,
    },
  ],
  related: legalRelatedDestinations(),
} as const;