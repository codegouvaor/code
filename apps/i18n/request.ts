import { getRequestConfig } from "next-intl/server";
import { routing } from "./routing";
import fr from "../messages/fr.json";
import en from "../messages/en.json";

/**
 * Message catalogs available on this server, keyed by locale.
 * A partial catalog is fine: every missing key falls back to the default
 * locale (`fr`), so a new language can be shipped incrementally without
 * touching the components.
 */
const catalogs = {
  fr,
  en,
} as const;

type Catalog = (typeof catalogs)[keyof typeof catalogs];

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function deepMerge<T extends Record<string, unknown>>(base: T, override: Record<string, unknown>): T {
  const result: Record<string, unknown> = { ...base };

  for (const [key, value] of Object.entries(override)) {
    const current = result[key];

    if (isPlainObject(current) && isPlainObject(value)) {
      result[key] = deepMerge(current, value);
      continue;
    }

    result[key] = value;
  }

  return result as T;
}

export default getRequestConfig(async ({ requestLocale }) => {
  const locale = await requestLocale;
  const resolvedLocale =
    locale && Object.prototype.hasOwnProperty.call(catalogs, locale)
      ? (locale as keyof typeof catalogs)
      : routing.defaultLocale;

  const base = catalogs[routing.defaultLocale];
  const messages =
    resolvedLocale === routing.defaultLocale
      ? base
      : deepMerge<Catalog>(base, catalogs[resolvedLocale] as unknown as Record<string, unknown>);

  return {
    locale: resolvedLocale,
    messages,
  };
});
