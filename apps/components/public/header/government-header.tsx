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
  codeRepositoryUrl,
  lifecycleThemes,
  pageAnchors,
  searchPath,
  type LifecycleTheme,
} from "@/lib/site-structure";
import { useAuth } from "@/context/AuthContext";
import { UserAccountMenu } from "@/components/public/header/user-account-menu";
import { siteAccountConfig } from "@/lib/site-config";

/** Whether the current pathname corresponds to a navigation href. */
const isNavItemActive = (href: string, pathname: string): boolean =>
  href === "/" ? pathname === "/" : pathname === href || pathname.startsWith(`${href}/`);

/**
 * Collect every href reachable from a theme mega-menu (theme hub, leader
 * link and category links) so the parent tab can be marked active when the
 * user lands on any child page.
 */
function collectChildHrefs(theme: LifecycleTheme): string[] {
  const hrefs: string[] = [theme.href];
  for (const cat of theme.categories) {
    for (const link of cat.links) hrefs.push(link.href);
  }
  return hrefs;
}

/**
 * Header of CODE.GOUV.AOR.
 *
 * The platform is organised around the lifecycle of the public digital
 * infrastructure. The permanent architecture of the header is therefore the
 * six lifecycle themes — Définir, Concevoir, Construire, Déployer,
 * Exploiter, Faire évoluer — each opening a mega-menu panel describing the
 * resources of that phase. Sub-resources never clutter the main bar.
 *
 * The quick-access toolbar keeps a search box, an external link to the
 * source repository (marked as external) and the account entry (login, or
 * the `UserAccountMenu` when authenticated — driven by `siteAccountConfig`).
 */
export function GovernmentHeader() {
  const t = useTranslations();
  const tPrimaryNav = useTranslations("nav.primary");
  const tNavPanel = useTranslations("nav.panel");
  const tBrand = useTranslations("brand");
  const pathname = usePathname();
  const router = useRouter();

  const navigationItems: MainNavigationProps.Item[] = lifecycleThemes.map((theme) => {
    const childHrefs = collectChildHrefs(theme);
    const isActive =
      isNavItemActive(theme.href, pathname) ||
      childHrefs.some((href) => isNavItemActive(href, pathname));

    const categories: MegaMenuProps.Category[] = theme.categories.map((category) => ({
      categoryMainText: tNavPanel(category.titleKey),
      links: category.links.map((link) => ({
        text: tNavPanel(link.labelKey),
        linkProps: { href: link.href },
        isActive: isNavItemActive(link.href, pathname),
      })),
    }));

    return {
      isActive,
      text: tPrimaryNav(theme.labelKey),
      megaMenu: {
        leader: {
          title: tPrimaryNav(theme.labelKey),
          paragraph: tNavPanel(theme.leader.paragraphKey),
          link: {
            text: tNavPanel(theme.leader.linkLabelKey),
            linkProps: { href: theme.href },
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

  // Quick-access items — the login link is replaced by the account menu
  // when the user is authenticated.
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
