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
import { getDomainUrl } from "@/lib/domains";
import {
  headerNavigation,
  pageAnchors,
  searchPath,
  type HeaderNavCategory,
  type HeaderNavEntry,
  type HeaderNavLink,
  type HeaderNavMegaEntry,
} from "@/lib/site-structure";
import { useAuth } from "@/context/AuthContext";
import { UserAccountMenu } from "@/components/public/header/user-account-menu";
import { siteAccountConfig } from "@/lib/site-config";

/** Whether the current pathname corresponds to a navigation href. */
const isNavItemActive = (href: string, pathname: string): boolean =>
  href === "/" ? pathname === "/" : pathname === href || pathname.startsWith(`${href}/`);

/** Every destination reachable from a mega-menu panel (heading links and
 *  rows) — used to mark the section current. */
function collectMegaMenuHrefs(section: HeaderNavMegaEntry): string[] {
  const hrefs: string[] = [];
  for (const category of section.categories) {
    if (category.href) hrefs.push(category.href);
    for (const link of category.links) hrefs.push(link.href);
  }
  return hrefs;
}

/** Hrefs used to decide whether an entry of the bar is current. */
function entryMatchHrefs(entry: HeaderNavEntry): string[] {
  if (entry.kind === "link") return [...(entry.matchHrefs ?? [entry.href])];
  return [...(entry.matchHrefs ?? collectMegaMenuHrefs(entry))];
}

/** External-link props (new window + accessible title). GitHub links carry the
 *  dedicated header title, everything else the generic one. */
const externalTitle = (link: HeaderNavLink, headerTitle: string, openWindowTitle: string) =>
  link.labelKey === "github" ? headerTitle : openWindowTitle;

/**
 * Header of CODE.GOUV.AOR — the public portal of reference for building
 * Astoria's public digital technology.
 *
 * The navigation is a government-portal information architecture answering
 * « Que puis-je trouver sur CODE.GOUV.AOR ? »:
 *
 *   Accueil       — simple link to the homepage (the portal entry point).
 *   Actualités    — simple link to the news channel of the platform
 *                   (external releases feed).
 *   Communauté    — mega-menu: who builds public digital technology.
 *   Documentation — mega-menu: « Comment construire le numérique public ? »,
 *                   with the six lifecycle themes as its columns. The six
 *                   themes are not six entries of the bar.
 *   Ressources    — mega-menu: the concrete building blocks.
 *   À propos      — mega-menu: institutional pages of the platform.
 *
 * The four mega-menus each expose four to six columns (defined in
 * `headerNavigation`). The data lives in `headerNavigation` (site-structure);
 * this component only decides how to render an entry. The quick-access
 * toolbar keeps the search box and the account entry (login, or the
 * `UserAccountMenu` when authenticated — driven by `siteAccountConfig`).
 * The source repository stays reachable from the navigation (Communauté,
 * À propos) and the footer.
 */
export function GovernmentHeader() {
  const t = useTranslations();
  const tPrimaryNav = useTranslations("nav.primary");
  const tNavPanel = useTranslations("nav.panel");
  const tBrand = useTranslations("brand");
  const pathname = usePathname();
  const router = useRouter();

  /** Top-level highlight: only one entry is current at a time (first match
   *  wins in bar order). Accueil matches `/`; Actualités is neutral (its
   *  panel is a feed of shortcuts); the other sections match their panel
   *  destinations. */
  const isEntryCurrent = React.useMemo(() => {
    const hrefLists = headerNavigation.map(entryMatchHrefs);
    let alreadyMatched = false;
    return hrefLists.map((hrefs) => {
      if (alreadyMatched) return false;
      const matches = hrefs.some((href) => isNavItemActive(href, pathname));
      if (matches) alreadyMatched = true;
      return matches;
    });
  }, [pathname]);

  /** One mega-menu column: heading (plain text or a destination) + rows. */
  const buildCategory = (category: HeaderNavCategory): MegaMenuProps.Category => {
    const title = category.titlePrimary
      ? tPrimaryNav(category.titleKey)
      : tNavPanel(category.titleKey);

    const links = category.links.map((link) => {
      const isExternal = link.external === true;
      return {
        text: link.primary ? tPrimaryNav(link.labelKey) : tNavPanel(link.labelKey),
        linkProps: {
          href: link.href,
          ...(isExternal && {
            target: "_blank",
            rel: "noopener noreferrer",
            title: externalTitle(link, t("header.githubTitle"), t("common.openNewWindow")),
          }),
        },
        isActive: isNavItemActive(link.href, pathname),
      };
    });

    if (category.href !== undefined) {
      return {
        categoryMainLink: {
          text: title,
          linkProps: {
            href: category.href,
            ...(category.external && {
              target: "_blank",
              rel: "noopener noreferrer",
              title: t("common.openNewWindow"),
            }),
          },
        },
        links,
      };
    }

    return { categoryMainText: title, links };
  };

  /** One navigation entry of the ADS header. */
  const buildEntry = (entry: HeaderNavEntry, index: number): MainNavigationProps.Item => {
    const label = tPrimaryNav(entry.labelKey);
    const isActive = isEntryCurrent[index];

    if (entry.kind === "link") {
      return {
        isActive,
        text: label,
        linkProps: {
          href: entry.href,
          ...(entry.external && {
            target: "_blank",
            rel: "noopener noreferrer",
            title: t("common.openNewWindow"),
          }),
        },
      };
    }

    return {
      isActive,
      text: label,
      megaMenu: {
        // No leader band (title + description) on purpose: a mega-menu only
        // exposes its sections as link columns, closed by the ADS “Fermer”
        // button that MegaMenu renders itself.
        categories: entry.categories.map(buildCategory),
      },
    };
  };

  const navigationItems: MainNavigationProps.Item[] = headerNavigation.map(buildEntry);

  const handleSearch = (text: string) => {
    const query = text.trim();
    router.push(query ? `${searchPath}?q=${encodeURIComponent(query)}` : searchPath);
  };

  // Auth state for conditional account UI
  const { isAuthenticated, isLoading: isAuthLoading } = useAuth();

  // Quick-access items — the login link, replaced by the account menu when
  // the user is authenticated, then the display settings.
  const quickAccessItems = React.useMemo(() => {
    const items: HeaderProps.QuickAccessItem[] = [];

    if (!isAuthenticated || isAuthLoading) {
      items.push({
        iconId: "fr-icon-account-circle-line",
        text: t("header.loginLink"),
        linkProps: { href: getDomainUrl("sso", "/login") },
      });
    }

    items.push(headerFooterDisplayItem);
    return items;
  }, [isAuthenticated, isAuthLoading, t, tNavPanel]);

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
        identity={{
          imgUrl: "/astoria-gouv.png",
          alt: tBrand("republicName"),
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
