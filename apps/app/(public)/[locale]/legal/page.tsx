import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { localizedAlternates, resolveLocaleParam } from "@/lib/localized-metadata";
import { legalLandingContent } from "@/lib/content/legal/legal-landing-content";
import { localizeArticle } from "@/lib/theme-localize";
import { ThemeArticle } from "@/components/public/content/theme-page";

const PAGE_PATH = "/legal";
const PAGE_NAMESPACE = "pages.legal.landing";

type PageProps = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);
  const t = await getTranslations({ locale, namespace: "pages.legal" });

  return {
    title: t("meta.title"),
    description: t("meta.description"),
    ...localizedAlternates(locale, PAGE_PATH),
  };
}

/**
 * Legal information landing page of the portal (`/legal`).
 *
 * Opens the four legal documents of the Banque centrale d'Astoria
 * (accessibility, privacy, terms of use, cookies) plus the portal sitemap.
 * Content is fully driven by the message catalogs (`pages.legal.landing.*` and
 * `pages.legal.related.*`) through `lib/theme-localize.ts`.
 */
export default async function LegalPage({ params }: PageProps) {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);
  setRequestLocale(locale);

  const t = await getTranslations({ locale, namespace: PAGE_NAMESPACE });
  const tRelated = await getTranslations({ locale, namespace: "pages.legal.related" });

  return <ThemeArticle content={localizeArticle(legalLandingContent, t, tRelated)} currentHref={PAGE_PATH} />;
}