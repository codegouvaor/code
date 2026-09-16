import type { ArticleContent } from "@/lib/theme-localize";
import { legalRelatedDestinations } from "@/lib/content/legal/legal-related";

/**
 * Content configuration of the legal landing page (`/legal`).
 *
 * Structure only: every display string is resolved from the message catalogs
 * (`pages.legal.landing.*` and `pages.legal.related.*`) by
 * `lib/theme-localize.ts`. The page opens the four legal documents of the
 * portal; the documents themselves are rendered by their dedicated pages.
 */
export const legalLandingContent: ArticleContent = {
  hero: {
    kickerKey: "hero.kicker",
    titleKey: "hero.title",
    leadKey: "hero.lead",
    noticeKey: "hero.notice",
  },
  sections: [],
  related: legalRelatedDestinations(),
} as const;