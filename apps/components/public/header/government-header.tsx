"use client";

import * as React from "react";
import { Header } from "@codegouvaor/react-ads/Header";
import { SkipLinks } from "@codegouvaor/react-ads/SkipLinks";
import { headerFooterDisplayItem } from "@codegouvaor/react-ads/Display";
import type { HeaderProps } from "@codegouvaor/react-ads/Header";
import type { MainNavigationProps } from "@codegouvaor/react-ads/MainNavigation";
import type { MegaMenuProps } from "@codegouvaor/react-ads/MainNavigation/MegaMenu";
import { useTranslations } from "next-intl";
import { usePathname, useRouter } from "@/i18n/navigation";
import {
  pageAnchors,
  primaryNavigation,
  searchPath,
  type NavigationSection,
} from "@/lib/site-structure";
import { useAuth } from "@/context/AuthContext";
import { UserAccountMenu } from "@/components/public/header/user-account-menu";
import { siteAccountConfig } from "@/lib/site-config";

/** Whether the current pathname corresponds to a navigation href. */
const isNavItemActive = (href: string, pathname: string): boolean =>
  href === "/" ? pathname === "/" : pathname === href || pathname.startsWith(`${href}/`);

/**
 * Collect every href reachable from a theme (leader action, sections and their
 * entries) so the parent tab can be marked active when the user lands on any
 * child page — even if the child href lives outside the parent's own path
 * tree.
 */
function collectChildHrefs(section: NavigationSection): string[] {
  const hrefs: string[] = [section.leader.link.href];
  for (const item of section.primaryItems) {
    hrefs.push(item.href);
    for (const link of item.links) hrefs.push(link.href);
  }
  return hrefs;
}

/**
 * Header of CODE — the public developer platform of Astoria's administration.
 *
 * The navigation is the spine of the platform, organised on a 7 × 4 × 4
 * information architecture (7 grands thèmes → 4 sections → 4 entrées):
 *
 *   Explorer       → discover the ecosystem
 *   Projets        → the live project & repository platform
 *   Organisations  → the organisations, teams and communities
 *   Développement  → the tools and integrations to build
 *   Documentation  → the lifecycle of a public digital service
 *   Ressources     → the reusable building blocks
 *   Communauté     → contribute, propose and debate
 *
 * Each theme opens an institutional mega-menu: a leader band (the theme
 * description and its main action) and up to four section columns, each with
 * up to four entries. The global navigation is deliberately distinct from the
 * navigation *inside* a project (code, issues, pull requests…): on Code you
 * first navigate Code, then a project.
 *
 * The whole navigation is configuration-driven (`primaryNavigation` in
 * `@/lib/site-structure`): the seven themes and their panels are defined
 * there, and nothing else competes with them in the bar.
 *
 * The header behaviour (mega-menu opening on click, close on outside click and
 * `Escape`, keyboard support, mobile drawer) is provided by the ADS runtime
 * (`StartDsfrOnHydration`): the same `primaryNavigation` data drives the
 * desktop mega-menus and the hierarchical mobile drawer, so `site-structure.ts`
 * is the single source of truth for both.
 *
 * The search button opens the header search dialog, which targets the central
 * search of the platform (`/search`). When the user is authenticated the
 * account entry is replaced by a custom account menu (`UserAccountMenu`).
 */
export function GovernmentHeader() {
  const t = useTranslations();
  const tPrimaryNav = useTranslations("nav.primary");
  const tNavPanel = useTranslations("nav.panel");
  const tBrand = useTranslations("brand");
  const pathname = usePathname();
  const router = useRouter();

  /** External-link props (new window + accessible title). GitHub links carry
   *  the dedicated header title, everything else the generic one. */
  const externalProps = (labelKey: string) => ({
    target: "_blank" as const,
    rel: "noopener noreferrer",
    title: labelKey === "github" ? t("header.githubTitle") : t("common.openNewWindow"),
  });

  const navigationItems: MainNavigationProps.Item[] = primaryNavigation.map((section) => {
    const childHrefs = collectChildHrefs(section);
    const isActive =
      isNavItemActive(section.href, pathname) ||
      childHrefs.some((href) => isNavItemActive(href, pathname));

    // The four sections of the theme are the pillars of its panel: each is
    // headed by its title (plain text, not a destination) and followed by its
    // entries, so the destinations are immediately visible and reachable.
    // External entries open in a new window.
    const categories: MegaMenuProps.Category[] = section.primaryItems.map(
      (item): MegaMenuProps.Category => ({
        categoryMainText: tNavPanel(item.labelKey),
        links: item.links.map((link) => ({
          text: tNavPanel(link.labelKey),
          linkProps: {
            href: link.href,
            ...(link.external && externalProps(link.labelKey)),
          },
          isActive: isNavItemActive(link.href, pathname),
        })),
      })
    );

    return {
      isActive,
      text: tPrimaryNav(section.labelKey),
      // The seven themes must each stay on a single, aligned line: the DSFR
      // nav buttons shrink and wrap their label by default, so every theme
      // button keeps its label unwrapped.
      buttonProps: { style: { whiteSpace: "nowrap" } },
      megaMenu: {
        leader: {
          title: tNavPanel(section.leader.titleKey),
          paragraph: tNavPanel(section.leader.paragraphKey),
          link: {
            text: tNavPanel(section.leader.link.labelKey),
            linkProps: {
              href: section.leader.link.href,
              ...(section.leader.link.external && externalProps(section.leader.link.labelKey)),
            },
          },
        },
        categories,
      },
    };
  });

  const handleSearch = (text: string) => {
    const query = text.trim();
    router.push(query ? `${searchPath}?q=${encodeURIComponent(query)}` : searchPath);
  };

  // Auth state for conditional account UI
  const { isAuthenticated, isLoading: isAuthLoading } = useAuth();

  // Quick-access items — the account entry links to the login page when the
  // user is not authenticated, and is replaced by the account menu (with its
  // personal entries) once the user is authenticated.
  //
  // The login route lives at the root (`/login`, the auth group is not
  // locale-prefixed), so it is rendered as a plain <a> rather than through the
  // localized Link (which would prefix the locale to `/fr/login`).
  const quickAccessItems = React.useMemo(() => {
    const items: (HeaderProps.QuickAccessItem | React.ReactElement | null)[] = [];

    if (!isAuthenticated || isAuthLoading) {
      items.push(
        <a className="fr-btn fr-icon-account-circle-line" href="/login">
          {t("header.loginLink")}
        </a>
      );
    }

    items.push(headerFooterDisplayItem);
    return items;
  }, [isAuthenticated, isAuthLoading, t]);

  return (
    <>
      <SkipLinks
        links={[
          { label: t("common.skipToContent"), anchor: `#${pageAnchors.content}` },
          { label: t("common.skipToFooter"), anchor: `#${pageAnchors.footer}` },
        ]}
      />
      <Header
        className="gov-header"
        classes={{ brand: "fr-enlarge-link" }}
        identity={{
          imgUrl: "/astoria-gouv.png",
          alt: tBrand("republicName"),
          // The lockup artwork already carries the full wordmark, so no
          // institution line is displayed under the image. ADS requires the
          // field, hence the empty string.
          institution: "",
        }}
        homeLinkProps={{
          href: "/",
          title: t("header.homeTitle"),
        }}
        serviceTitle={t("header.serviceTitle")}
        serviceTagline={t("header.serviceTagline")}
        navigation={navigationItems}
        quickAccessItems={quickAccessItems}
        renderSearchInput={(params) => (
          <input {...params} placeholder={t("meta.searchPlaceholder")} />
        )}
        onSearchButtonClick={handleSearch}
      />
      {/* Account menu — rendered outside the ADS Header so it can use
          its own dropdown positioning and auth state without conflicting
          with the ADS quick-access toolbar. */}
      {siteAccountConfig.enabled && <UserAccountMenu />}
    </>
  );
}