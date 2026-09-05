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
import { pageAnchors, primaryNavigation, searchPath } from "@/lib/site-structure";
import { useAuth } from "@/context/AuthContext";
import { UserAccountMenu } from "@/components/public/header/user-account-menu";
import { siteAccountConfig } from "@/lib/site-config";

/** Whether the current pathname corresponds to a navigation href. */
const isNavItemActive = (href: string, pathname: string): boolean =>
  href === "/" ? pathname === "/" : pathname === href || pathname.startsWith(`${href}/`);

/**
 * Collect every href reachable from a mega-menu item (categories, featured
 * link, leader link) so the parent tab can be marked active when the user
 * lands on any child page — even if the child href lives outside the
 * parent's own path tree (e.g. /decryptages under Actualites -> /news).
 */
function collectChildHrefs(item: (typeof primaryNavigation)[number]): string[] {
  if (item.type === "link") return [];
  const hrefs: string[] = [item.leader.link.href];
  if (item.featuredLink) hrefs.push(item.featuredLink.href);
  for (const cat of item.categories ?? []) {
    if (cat.mainLink) hrefs.push(cat.mainLink.href);
    for (const link of cat.links) hrefs.push(link.href);
  }
  return hrefs;
}

/**
 * Government Header of the Astoria portal.
 *
 * Main navigation — the permanent architecture of the portal, six entries,
 * each answering one user intention:
 *
 *   Le Gouvernement      -> Qui gouverne ?
 *   L'action publique    -> Que fait la Republique ?
 *   Services publics     -> Puis-je faire avec l'Etat ?
 *   Actualites           -> Que se passe-t-il actuellement ?
 *   La Republique        -> Comment fonctionne l'Etat ?
 *   Informations utiles  -> O trouver une information pratique ?
 *
 * When the user is authenticated, the "Se connecter" link in the
 * quick-access toolbar is hidden and a custom account menu
 * (`UserAccountMenu`) is rendered instead. The menu content is driven
 * by `siteAccountConfig` so each site can present a different account
 * interface without touching this component.
 */
export function GovernmentHeader() {
  const t = useTranslations();
  const tPrimaryNav = useTranslations("nav.primary");
  const tNavPanel = useTranslations("nav.panel");
  const tBrand = useTranslations("brand");
  const pathname = usePathname();
  const router = useRouter();

  const navigationItems: MainNavigationProps.Item[] = primaryNavigation.map((item) => {
    const childHrefs = collectChildHrefs(item);
    const isActive =
      isNavItemActive(item.href, pathname) ||
      childHrefs.some((href) => isNavItemActive(href, pathname));

    const common = {
      isActive,
      text: tPrimaryNav(item.labelKey),
    };

    if (item.type === "link") {
      return { ...common, linkProps: { href: item.href } };
    }

    const categories: MegaMenuProps.Category[] = [
      ...(item.featuredLink
        ? [
            {
              categoryMainLink: {
                text: tNavPanel(item.featuredLink.titleKey),
                linkProps: { href: item.featuredLink.href },
              },
              links: [
                {
                  text: t("home.news.featured.title"),
                  linkProps: { href: item.featuredLink.href },
                  isActive: isNavItemActive(item.featuredLink.href, pathname),
                },
              ],
            } satisfies MegaMenuProps.Category,
          ]
        : []),
      ...(item.categories ?? []).map((category): MegaMenuProps.Category =>
        category.mainLink
          ? {
              categoryMainLink: {
                text: tNavPanel(category.mainLink.labelKey),
                linkProps: { href: category.mainLink.href },
              },
              links: category.links.map((link) => ({
                text: tNavPanel(link.labelKey),
                linkProps: { href: link.href },
                isActive: isNavItemActive(link.href, pathname),
              })),
            }
          : {
              categoryMainText: tNavPanel(category.titleKey),
              links: category.links.map((link) => ({
                text: tNavPanel(link.labelKey),
                linkProps: { href: link.href },
                isActive: isNavItemActive(link.href, pathname),
              })),
            }
      ),
    ];

    return {
      ...common,
      megaMenu: {
        leader: {
          title: tNavPanel(item.leader.titleKey),
          paragraph: tNavPanel(item.leader.paragraphKey),
          link: {
            text: tNavPanel(item.leader.link.labelKey),
            linkProps: { href: item.leader.link.href },
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
