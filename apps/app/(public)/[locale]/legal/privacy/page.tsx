import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { localizedAlternates, resolveLocaleParam } from "@/lib/localized-metadata";
import { privacyContent } from "@/lib/content/legal/privacy-content";
import { localizeArticle } from "@/lib/theme-localize";
import { ThemeArticle } from "@/components/public/content/theme-page";

const PAGE_PATH = "/legal/privacy";
const PAGE_NAMESPACE = "pages.legal.privacy";

type PageProps = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);
  const t = await getTranslations({ locale, namespace: PAGE_NAMESPACE });

  return {
    title: t("meta.title"),
    description: t("meta.description"),
    ...localizedAlternates(locale, PAGE_PATH),
  };
}

/**
 * Privacy policy of the Banque centrale d'Astoria portal.
 *
 * Content is fully driven by the message catalogs
 * (`pages.legal.privacy.*` and `pages.legal.related.*`) through
 * `lib/theme-localize.ts`.
 */
export default async function PrivacyPage({ params }: PageProps) {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);
  setRequestLocale(locale);

  const t = await getTranslations({ locale, namespace: PAGE_NAMESPACE });
  const tRelated = await getTranslations({ locale, namespace: "pages.legal.related" });

  return <ThemeArticle content={localizeArticle(privacyContent, t, tRelated)} currentHref={PAGE_PATH} />;
}