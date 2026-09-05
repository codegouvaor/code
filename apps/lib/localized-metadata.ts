import { getPathname } from "@/i18n/navigation";
import { routing, type Locale } from "@/i18n/routing";

/**
 * Builds the `alternates` metadata block (canonical + hreflang) for a page of
 * the localized portal. Hrefs are locale-agnostic pathnames; URLs are resolved
 * per locale through the next-intl router so the active language and the
 * language-prefix strategy stay in a single place.
 */
export function localizedAlternates(locale: Locale, href: string) {
  const languages: Record<string, string> = {};

  for (const targetLocale of routing.locales) {
    languages[targetLocale] = getPathname({ href, locale: targetLocale });
  }

  return {
    alternates: {
      canonical: languages[locale],
      languages,
    },
  };
}

/** Normalizes the `[locale]` param, defaulting to the configured default. */
export function resolveLocaleParam(localeParam: string): Locale {
  return routing.locales.includes(localeParam as Locale)
    ? (localeParam as Locale)
    : routing.defaultLocale;
}
