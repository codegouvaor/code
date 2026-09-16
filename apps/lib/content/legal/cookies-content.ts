import type { ArticleContent } from "@/lib/theme-localize";
import { legalRelatedDestinations } from "@/lib/content/legal/legal-related";

/**
 * Content configuration of the cookie policy page (`/legal/cookies`).
 *
 * Structure only: every display string is resolved from the message catalogs
 * (`pages.legal.cookies.*` and `pages.legal.related.*`) by
 * `lib/theme-localize.ts`. The “cookies déposés” section renders the
 * declaration table of the cookie families used by the portal.
 */
export const cookiesContent: ArticleContent = {
  hero: {
    kickerKey: "hero.kicker",
    titleKey: "hero.title",
    leadKey: "hero.lead",
    noticeKey: "hero.notice",
  },
  sections: [
    {
      key: "definition",
      id: "definition",
      lead: true,
      paragraphCount: 2,
    },
    {
      key: "cookiesDeposes",
      id: "cookies-deposes",
      subtle: true,
      lead: true,
      table: {
        headerCount: 3,
        rowCells: [3, 3, 3, 3],
        note: true,
      },
    },
    {
      key: "gestion",
      id: "gestion-des-cookies",
      lead: true,
      bulletCount: 3,
      notice: true,
    },
  ],
  related: legalRelatedDestinations(),
} as const;