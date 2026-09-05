"use client";

import { Footer } from "@codegouvaor/react-ads/Footer";
import type { FooterProps } from "@codegouvaor/react-ads/Footer";
import { useTranslations } from "next-intl";
import {
  footerLegalLinks,
  footerPlatformLinks,
  legalPaths,
  openDataUrl,
  pageAnchors,
} from "@/lib/site-structure";
import { LocaleSwitcher } from "../locale-switcher";

const HOME_PATH = "/";

/** Official portal domains of the Republic of Astoria, shown in the footer. */
const ECOSYSTEM_DOMAINS: string[] = ["info.gouv.aor", "design.gouv.aor", "data.gouv.aor"];

/**
 * Footer of CODE.GOUV.AOR — the shared institutional base of the platform.
 *
 * It is built entirely from the centralized `site-structure` configuration
 * and the message catalogs, so it can be reused as the official footer of
 * the whole CODE ecosystem without page-specific implementation:
 *
 *  - the ecosystem portals (official domains) next to the institutional
 *    identity and the platform mission,
 *  - the platform-level destinations (Documentation, Standards, Ressources,
 *    Projets, Communauté, GitHub) and the institutional bottom bar.
 *
 * The six lifecycle themes are intentionally left out of the footer: they
 * belong to the main navigation of the platform, and the footer stays a
 * sober, common base. ADS provides the markup and the responsive behaviour;
 * this component only decides *what* is shown.
 */
export function GovernmentFooter() {
  const t = useTranslations();
  const tNavPanel = useTranslations("nav.panel");
  const tBrand = useTranslations("brand");

  const platformItems: FooterProps.BottomItem[] = footerPlatformLinks.map(
    ({ labelKey, href }) => ({
      text: tNavPanel(labelKey),
      iconId: labelKey === "github" ? "fr-icon-external-link-line" : undefined,
      linkProps:
        labelKey === "github"
          ? {
              href,
              target: "_blank",
              rel: "noopener noreferrer",
              title: t("header.githubTitle"),
            }
          : { href },
    })
  );

  const legalItems: FooterProps.BottomItem[] = footerLegalLinks.map(
    ({ labelKey, href, external }) => ({
      text: t(`footer.bottom.${labelKey}`),
      iconId: external ? "fr-icon-external-link-line" : undefined,
      linkProps: external
        ? {
            href,
            target: "_blank",
            rel: "noopener noreferrer",
            title: t("common.openNewWindow"),
          }
        : { href },
    })
  );

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
      domains={ECOSYSTEM_DOMAINS}
      websiteMapLinkProps={{ href: legalPaths.sitemap }}
      accessibilityLinkProps={{ href: legalPaths.accessibility }}
      termsLinkProps={{ href: legalPaths.terms }}
      bottomItems={[
        ...platformItems,
        ...legalItems,
        // Language switcher rendered as a React node inside the bottom bar.
        <LocaleSwitcher key="locale-switcher" />,
      ]}
      license={t.rich("footer.license", {
        link: (chunks) => (
          <a href={openDataUrl} target="_blank" rel="noopener noreferrer">
            {chunks}
          </a>
        ),
      })}
    />
  );
}
