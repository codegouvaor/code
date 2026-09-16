import type { CSSProperties } from "react";
import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { localizedAlternates, resolveLocaleParam } from "@/lib/localized-metadata";
import { Link } from "@/i18n/navigation";
import { primaryNavigation } from "@/lib/site-structure";
import { sitemapTransversalLinks } from "@/lib/content/sitemap/sitemap-content";
import { ThemeHero, ThemeSection, linkListStyle } from "@/components/public/content/theme-page";

const PAGE_PATH = "/sitemap";
const PAGE_NAMESPACE = "pages.sitemap";

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

/* Layout helpers below use the ADS design tokens through `var(--ads-*)`, the
 * single source of tokens, the same way the home page composes its blocks. */

const sectionHeadingStyle: CSSProperties = {
  margin: "0 0 1.25rem",
  fontSize: "1.25rem",
  lineHeight: 1.3,
  fontWeight: 700,
};

const themeHeadingStyle: CSSProperties = {
  margin: "0 0 0.75rem",
  fontSize: "0.8125rem",
  fontWeight: 700,
  letterSpacing: "0.06em",
  textTransform: "uppercase",
  color: "var(--ads-color-primary)",
};

const linkRowStyle: CSSProperties = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "1rem",
  padding: "0.75rem 1rem",
  fontWeight: 600,
  fontSize: "0.9375rem",
  textDecoration: "none",
  color: "var(--ads-color-text)",
  border: "1px solid var(--ads-color-border)",
  borderTop: "none",
  background: "var(--ads-color-background)",
};

const themeGroupStyle: CSSProperties = {
  marginBottom: "2.5rem",
};

const themeGridStyle: CSSProperties = {
  listStyle: "none",
  margin: "0",
  padding: "0",
};

/**
 * Sitemap of the Banque centrale d'Astoria portal (`/sitemap`).
 *
 * The page mirrors the information architecture of the portal: the seven
 * entries of the header navigation and, below, the transversal pages of the
 * footer. It is fully driven by `primaryNavigation` (`lib/site-structure.ts`)
 * and the message catalogs (`pages.sitemap.*`, `nav.primary` and
 * `nav.panel.*`), so the sitemap, the header and the footer can never drift
 * apart.
 */
export default async function SitemapPage({ params }: PageProps) {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);
  setRequestLocale(locale);

  const t = await getTranslations({ locale, namespace: PAGE_NAMESPACE });
  const tPanel = await getTranslations({ locale, namespace: "nav.panel" });

  return (
    <>
      <ThemeHero kicker={t("hero.kicker")} title={t("hero.title")} lead={t("hero.lead")} />

      <ThemeSection
        id="sitemap-portail"
        kicker={t("sections.portail.kicker")}
        title={t("sections.portail.title")}
        lead={t("sections.portail.lead")}
      >
        {primaryNavigation.map((section) => (
          <section key={section.labelKey} style={themeGroupStyle}>
            <h3 style={sectionHeadingStyle}>{tPanel(`${section.labelKey}.title`)}</h3>
            <div className="fr-grid-row fr-grid-row--gutters">
              {section.primaryItems.map((theme) => (
                <div key={theme.labelKey} className="fr-col-12 fr-col-md-6 fr-col-lg-3">
                  <h4 style={themeHeadingStyle}>{tPanel(theme.labelKey)}</h4>
                  <ul role="list" style={themeGridStyle}>
                    {theme.links.map((link) => (
                      <li key={link.labelKey}>
                        <Link href={link.href} style={linkRowStyle}>
                          {tPanel(link.labelKey)}
                          <span className="fr-icon-arrow-right-line" aria-hidden="true" />
                        </Link>
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          </section>
        ))}
      </ThemeSection>

      <ThemeSection
        id="sitemap-transversal"
        kicker={t("sections.transversal.kicker")}
        title={t("sections.transversal.title")}
        lead={t("sections.transversal.lead")}
        subtle
      >
        <ul role="list" style={linkListStyle}>
          {sitemapTransversalLinks.map((link) => (
            <li key={link.key}>
              <Link href={link.href} style={linkRowStyle}>
                {t(`sections.transversal.links.${link.key}`)}
                <span className="fr-icon-arrow-right-line" aria-hidden="true" />
              </Link>
            </li>
          ))}
        </ul>
      </ThemeSection>
    </>
  );
}