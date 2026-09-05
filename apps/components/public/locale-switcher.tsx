"use client";

import { useCallback } from "react";
import { LanguageSelect } from "@codegouvaor/react-ads/LanguageSelect";
import { useLocale } from "next-intl";
import { usePathname, useRouter } from "@/i18n/navigation";
import { routing, type Locale } from "@/i18n/routing";
import { localeDisplayNames } from "@/i18n/locales";

/**
 * Language switcher shown in the Government Footer bottom bar.
 *
 * It reuses the ADS `LanguageSelect` component (visual + behaviour source of
 * truth) and plugs the next-intl router into it so that switching language:
 *  - keeps the current page,
 *  - keeps the current URL query parameters,
 *  - uses the i18n routing (the locale prefix is managed by next-intl).
 *
 * The list of languages comes from the centralized `routing` configuration,
 * never from hard-coded JSX.
 *
 * NOTE: this component deliberately avoids `useSearchParams`. That hook
 * suspends during server rendering and would require a Suspense boundary,
 * whose hydration can race with the ADS core (started on hydration by
 * `StartDsfrOnHydration`), producing hydration mismatches on the collapse
 * markup. The query string is read from `window.location` at click time
 * instead, keeping the tree synchronous.
 */
export function LocaleSwitcher() {
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();

  const setLang = useCallback(
    (nextLang: string) => {
      const nextLocale = routing.locales.includes(nextLang as Locale)
        ? (nextLang as Locale)
        : routing.defaultLocale;

      if (nextLocale === locale) {
        return;
      }

      const query = typeof window !== "undefined" ? window.location.search : "";
      router.replace(query ? `${pathname}${query}` : pathname, {
        locale: nextLocale,
      });
    },
    [locale, pathname, router]
  );

  return (
    <LanguageSelect
      supportedLangs={routing.locales}
      fullNameByLang={localeDisplayNames}
      lang={locale as Locale}
      setLang={setLang}
    />
  );
}