import type { ArticleContent } from "@/lib/theme-localize";
import { legalRelatedDestinations } from "@/lib/content/legal/legal-related";

/**
 * Content configuration of the terms of use page (`/legal/terms`).
 *
 * Structure only: every display string is resolved from the message catalogs
 * (`pages.legal.terms.*` and `pages.legal.related.*`) by
 * `lib/theme-localize.ts`.
 */
export const termsContent: ArticleContent = {
  hero: {
    kickerKey: "hero.kicker",
    titleKey: "hero.title",
    leadKey: "hero.lead",
    noticeKey: "hero.notice",
  },
  sections: [
    {
      key: "objet",
      id: "objet",
      lead: true,
      paragraphCount: 2,
    },
    {
      key: "utilisation",
      id: "utilisation-du-portail",
      subtle: true,
      lead: true,
      bulletCount: 4,
    },
    {
      key: "contenus",
      id: "contenus-et-donnees",
      lead: true,
      paragraphCount: 2,
    },
    {
      key: "responsabilite",
      id: "responsabilite",
      subtle: true,
      lead: true,
      paragraphCount: 2,
    },
    {
      key: "liens",
      id: "liens",
      lead: true,
      paragraphCount: 1,
    },
    {
      key: "modification",
      id: "modification-des-conditions",
      subtle: true,
      lead: true,
      paragraphCount: 1,
      notice: true,
    },
  ],
  related: legalRelatedDestinations(),
} as const;