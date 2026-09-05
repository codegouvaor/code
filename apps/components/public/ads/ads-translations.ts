import { addHeaderTranslations } from "@codegouvaor/react-ads/Header";
import { addFooterTranslations } from "@codegouvaor/react-ads/Footer";
import { addMainNavigationTranslations } from "@codegouvaor/react-ads/MainNavigation";
import { addLanguageSelectTranslations } from "@codegouvaor/react-ads/LanguageSelect";
import { addSearchBarTranslations } from "@codegouvaor/react-ads/SearchBar";

/**
 * ASTORIA DESIGN SYSTEM — internal component strings.
 *
 * ADS components ship their accessibility/aria strings in French by default.
 * Every non-French locale the portal supports must be registered here so the
 * components can pick the right strings at runtime. This module is imported
 * both by the server layout and by the client ADS provider so the registry is
 * populated in both bundles.
 *
 * Keep the strings of the French locale out of this file: they are the
 * built-in defaults of @codegouvaor/react-ads.
 */
export function registerAdsTranslations(): void {
  // Header (mobile menu / search modal buttons)
  addHeaderTranslations({
    lang: "en",
    messages: {
      menu: "Menu",
      close: "Close",
    },
  });

  // Main navigation (aria label of the nav landmark)
  addMainNavigationTranslations({
    lang: "en",
    messages: {
      "main menu": "Main menu",
    },
  });

  // Footer (accessibility line, legal links and helpers)
  addFooterTranslations({
    lang: "en",
    messages: {
      "hide message": "Hide this message",
      "website map": "Sitemap",
      accessibility: "Accessibility",
      "non compliant": "non compliant",
      "partially compliant": "partially compliant",
      "fully compliant": "fully compliant",
      terms: "Terms and conditions",
      "cookies management": "Cookie management",
      "our partners": "Our partners",
      "open new window": "Opens in a new window",
    },
  });

  // Language select (button accessible title)
  addLanguageSelectTranslations({
    lang: "en",
    messages: {
      "select language": "Select language",
    },
  });

  // Search bar (label / placeholder)
  addSearchBarTranslations({
    lang: "en",
    messages: {
      label: "Search",
    },
  });
}
