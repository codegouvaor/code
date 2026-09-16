"use client";

import { Footer } from "@codegouvaor/react-ads/Footer";
import { useTranslations } from "next-intl";
import { legalPaths, pageAnchors } from "@/lib/site-structure";

const HOME_PATH = "/";

/** Official portal domains of the Republic of Astoria, shown in the footer. */
const OFFICIAL_DOMAINS: string[] = ["info.gouv.aor", "code.gouv.aor", "data.gouv.aor", "infra.gouv.aor"];

/**
 * Government Footer of the Astoria portal — secondary navigation zone of the
 * site, distinct from the main header navigation:
 *  - brand block + “managed by” line,
 *  - official portal domains,
 *  - bottom bar: Contact, Plan du portail, Documents opposables and the
 *    legal links.
 *
 * ADS provides the markup (columns, accessibility line, bottom bar) and the
 * responsive behaviour. This component only decides *what* is shown — from
 * the centralized `site-structure` configuration and the message catalogs.
 */
export function GovernmentFooter() {
  const t = useTranslations();
  const tBrand = useTranslations("brand");

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
      accessibilityLinkProps={{ href: legalPaths.accessibility }}
      termsLinkProps={{ href: legalPaths.terms }}
      bottomItems={[
        {
          text: t("footer.bottom.contact"),
          linkProps: { href: "/contact" },
        },
        {
          text: t("footer.bottom.planDuPortail"),
          linkProps: { href: legalPaths.sitemap },
        },
        {
          text: t("footer.bottom.documentsOpposables"),
          linkProps: { href: "/documents-opposables" },
        },
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
          <a href="https://code.astoria-gouv.org/" target="_blank" rel="noopener noreferrer">
            {chunks}
          </a>
        ),
      })}
    />
  );
}
