"use client";

import { Footer } from "@codegouvaor/react-ads/Footer";
import type { FooterProps } from "@codegouvaor/react-ads/Footer";
import { useTranslations } from "next-intl";
import { footerNavigation, legalPaths, pageAnchors } from "@/lib/site-structure";
import { LocaleSwitcher } from "../locale-switcher";

const HOME_PATH = "/";

/** Official portal domains of the Republic of Astoria, shown in the footer. */
const OFFICIAL_DOMAINS: string[] = [
  "code.gouv.aor",
  "service-public.gouv.aor",
  "data.gouv.aor",
];

/**
 * Government Footer of the Astoria portal, modelled after info.gouv.fr:
 *  - brand block + “managed by” line,
 *  - five link columns (Actualités, Grands dossiers, Prévenir les risques,
 *    Outils, L'État et moi) fed by the centralized `site-structure`
 *    configuration,
 *  - official portal domains and the legal bottom bar.
 *
 * ADS provides the markup (columns, accessibility line, bottom bar) and the
 * responsive behaviour. This component only decides *what* is shown — from
 * the centralized `site-structure` configuration and the message catalogs.
 */
export function GovernmentFooter() {
  const t = useTranslations();
  const tNavPanel = useTranslations("nav.panel");
  const tBrand = useTranslations("brand");

  const linkList = footerNavigation.map((column) => ({
    categoryName: t(`footer.columns.${column.columnKey}.title`),
    links: column.links.map(({ labelKey, href }) => ({
      text: tNavPanel(labelKey),
      linkProps: { href },
    })),
  })) as FooterProps.LinkList.List;

  return (
    <Footer
      id={pageAnchors.footer}
      className="gov-footer"
      accessibility="partially compliant"
      identity={{
        imgUrl: "/astoria-gouv.png",
        alt: tBrand("republicName"),
        // The lockup artwork already carries the full wordmark, so no
        // institution line is displayed under the image. ADS requires the
        // field, hence the empty string.
        institution: "",
      }}
      homeLinkProps={{
        href: HOME_PATH,
        title: t("header.homeTitle"),
      }}
      contentDescription={t("footer.contentDescription")}
      domains={OFFICIAL_DOMAINS}
      websiteMapLinkProps={{ href: legalPaths.sitemap }}
      accessibilityLinkProps={{ href: legalPaths.accessibility }}
      termsLinkProps={{ href: legalPaths.terms }}
      bottomItems={[
        {
          text: t("footer.bottom.privacy"),
          linkProps: { href: legalPaths.privacy },
        },
        {
          text: t("footer.bottom.cookies"),
          linkProps: { href: legalPaths.cookies },
        },
        {
          text: t("footer.bottom.publications"),
          linkProps: { href: "/publications-officielles" },
        },
      ]}
      license={t.rich("footer.license", {
        link: (chunks) => (
          <a href="https://code.gouv.aor/" target="_blank" rel="noopener noreferrer">
            {chunks}
          </a>
        ),
      })}
      linkList={linkList}
    />
  );
}