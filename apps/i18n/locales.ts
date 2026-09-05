import { routing, type Locale } from "./routing";

/**
 * Registry of the human-readable language labels shown in the header language
 * switcher. A language is displayed under its own endonym (« Français »,
 * « English », …), following internationalisation conventions.
 *
 * This is the single place where display metadata lives: adding a language
 * requires only adding an entry here, a message file in `apps/messages/` and
 * the code in `./routing.ts`.
 */
export const localeDisplayNames: Record<Locale, string> = {
  fr: "Français",
  en: "English",
};

/** Locales for which a message file exists (see `apps/messages/`). */
export const supportedLocales = routing.locales;
