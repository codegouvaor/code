import { defineRouting } from "next-intl/routing";

/**
 * Central routing configuration used by next-intl (navigation, middleware,
 * static params…) and by the portal components.
 *
 * To add a new supported language:
 *   1. add its code below,
 *   2. add a message file in `apps/messages/<code>.json` (a partial file is
 *      fine: untranslated keys gracefully fall back to the default locale),
 *   3. add its endonym in `./locales.ts`.
 *
 * The `localePrefix` is always on for web builds so that every public URL
 * carries its locale (`/fr/…`, `/en/…`) and switching language never loses
 * the current page. The only exception is the Capacitor native shell, which
 * cannot use a path prefix.
 */
const isCapacitor = process.env.CAPACITOR === "true";

export const routing = defineRouting({
  locales: ["fr", "en"],
  defaultLocale: "fr",
  localePrefix: isCapacitor ? "never" : "always",
});

export type Locale = (typeof routing.locales)[number];
