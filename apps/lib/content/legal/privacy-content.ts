import type { ArticleContent } from "@/lib/theme-localize";
import { legalPaths } from "@/lib/site-structure";
import { legalRelatedDestinations } from "@/lib/content/legal/legal-related";

/**
 * Content configuration of the privacy policy page (`/legal/privacy`).
 *
 * Structure only: every display string is resolved from the message catalogs
 * (`pages.legal.privacy.*` and `pages.legal.related.*`) by
 * `lib/theme-localize.ts`.
 */
export const privacyContent: ArticleContent = {
  hero: {
    kickerKey: "hero.kicker",
    titleKey: "hero.title",
    leadKey: "hero.lead",
    noticeKey: "hero.notice",
  },
  sections: [
    {
      key: "responsable",
      id: "responsable-de-traitement",
      lead: true,
      paragraphCount: 1,
      facts: [
        { key: "responsable" },
        { key: "adresse" },
        { key: "dpo" },
      ],
    },
    {
      key: "donnees",
      id: "donnees-collectees",
      subtle: true,
      lead: true,
      bulletCount: 4,
    },
    {
      key: "finalites",
      id: "finalites-et-bases-legales",
      lead: true,
      paragraphCount: 2,
    },
    {
      key: "conservation",
      id: "duree-de-conservation",
      subtle: true,
      lead: true,
      facts: [{ key: "duree" }, { key: "critere" }],
    },
    {
      key: "droits",
      id: "vos-droits",
      lead: true,
      bulletCount: 5,
    },
    {
      key: "cookies",
      id: "cookies-et-mesure-d-audience",
      subtle: true,
      lead: true,
      paragraphCount: 1,
      notice: true,
      cta: { href: legalPaths.cookies },
    },
  ],
  related: legalRelatedDestinations(),
} as const;