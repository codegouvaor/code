import type { CSSProperties } from "react";
import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { localizedAlternates, resolveLocaleParam } from "@/lib/localized-metadata";
import { ministryHome } from "@/lib/home-content";
import { PortalSearchBar } from "@/components/public/search/portal-search-bar";
import {
  ArticleCard,
  CtaButtonsGroup,
  LinkTile,
  SearchSuggestionTag,
} from "@/components/public/content/ads-fragments";
import { MonEspace } from "@/components/public/home/mon-espace";

const HOME_PATH = "/";
const NEWS_PATH = "/news";
/** Données & Ressources → Statistiques & indicateurs. */
const STATS_PATH = "/donnees-et-ressources/indicateurs";
const DATA_PATH = "/donnees-et-ressources";

type PageProps = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);

  const tHome = await getTranslations({ locale, namespace: "home" });
  const tMeta = await getTranslations({ locale, namespace: "meta" });

  return {
    title: { absolute: tHome("metaTitle") },
    description: tMeta("description"),
    ...localizedAlternates(locale, HOME_PATH),
  };
}

/* Layout helpers below use the ADS design tokens through `var(--ads-*)` (the
 * single source of tokens — main.css) so light/dark switching and theming stay
 * owned by the Design System. Only the ministry-specific arrangement of these
 * blocks is expressed here, inline, without any local stylesheet. */

const heroContainerStyle: CSSProperties = {
  maxWidth: "52rem",
  marginInline: "auto",
  textAlign: "center",
};

const searchBlockStyle: CSSProperties = {
  maxWidth: "42rem",
  margin: "2.25rem auto 0",
  textAlign: "left",
};

const searchTitleStyle: CSSProperties = {
  margin: "0 0 0.75rem",
  fontSize: "1.25rem",
  lineHeight: 1.3,
  fontWeight: 700,
};

const popularLabelStyle: CSSProperties = {
  margin: "0 0 0.5rem",
  fontSize: "0.875rem",
  color: "var(--ads-color-text-muted)",
};

const popularListStyle: CSSProperties = {
  listStyle: "none",
  margin: "0",
  padding: "0",
  display: "flex",
  flexWrap: "wrap",
  gap: "0.5rem",
};

const cardListStyle: CSSProperties = {
  listStyle: "none",
  margin: "0",
  padding: "0",
  display: "grid",
  gap: "1.5rem",
};

const linkListStyle: CSSProperties = {
  listStyle: "none",
  margin: "0",
  padding: "0",
  display: "grid",
  gap: "0",
  maxWidth: "72rem",
};

/** Whole-card link block (alerts, indicators, budget figures…). */
const teaserCardStyle: CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: "0.75rem",
  height: "100%",
  padding: "1.25rem",
  background: "var(--ads-color-background)",
  border: "1px solid var(--ads-color-border)",
  borderTop: "3px solid var(--ads-color-primary)",
  textDecoration: "none",
  color: "var(--ads-color-text)",
};

const teaserTagStyle: CSSProperties = {
  fontSize: "0.75rem",
  fontWeight: 700,
  letterSpacing: "0.06em",
  textTransform: "uppercase",
  color: "var(--ads-color-primary)",
};

const teaserTitleStyle: CSSProperties = {
  display: "block",
  fontSize: "1.0625rem",
  lineHeight: 1.35,
  fontWeight: 700,
};

const teaserDescStyle: CSSProperties = {
  display: "block",
  fontSize: "0.875rem",
  lineHeight: 1.55,
  color: "var(--ads-color-text-muted)",
};

const teaserArrowStyle: CSSProperties = {
  marginTop: "auto",
  alignSelf: "flex-end",
  fontSize: "1rem",
  color: "var(--ads-color-primary)",
};

const figureValueStyle: CSSProperties = {
  display: "block",
  fontSize: "clamp(1.5rem, 3vw, 2rem)",
  lineHeight: 1.2,
  fontWeight: 700,
};

const figureLabelStyle: CSSProperties = {
  display: "block",
  fontSize: "0.9375rem",
  fontWeight: 600,
};

/**
 * Homepage of the Ministry of Economy and Finance — the functional front
 * door of the economy and finance platform.
 *
 * The header allows exploring the ministry (six domains, unchanged); this
 * page allows acting. It answers “what can I do now?”, section after
 * section:
 *
 *   01 Hero / Recherche            — who we are, and search first
 *   02 Que souhaitez-vous faire ?  — frequent actions, each a real parcours
 *   03 Mon espace                  — the personal space (auth via MyGouv)
 *   04 Informations importantes    — operational alerts, before general news
 *   05 L'économie d'Astoria        — key economic indicators
 *   06 Budget de l'État            — public finances, made understandable
 *   07 Actualités                  — ministry news, after usages
 *   08 Ressources                  — compact shortcut to documentary content
 *   09 Le ministère                — discreet institutional closing
 *
 * Every section is driven by the `ministryHome` configuration
 * (lib/home-content.ts) and the message catalogs, so the content can evolve
 * without rewriting the interface. The six domains deliberately do not
 * appear here as a second navigation: they belong to the header. Newsletter
 * and social accounts are relegated to the footer (stay-in-touch zone of
 * GovernmentFooter).
 */
export default async function HomePage({ params }: PageProps) {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);
  setRequestLocale(locale);

  const t = await getTranslations({ locale, namespace: "home" });
  const tNavPanel = await getTranslations({ locale, namespace: "nav.panel" });

  return (
    <>
      {/* 01 — Hero / Recherche: institutional statement and the main search,
          visually the most important element of the page. */}
      <section className="gov-section" aria-labelledby="home-hero-title">
        <div className="gov-section__container" style={heroContainerStyle}>
          <p className="gov-kicker">{t("hero.kicker")}</p>
          <h1 id="home-hero-title">{t("hero.title")}</h1>
          <p className="gov-lead">{t("hero.lead")}</p>
          <div style={searchBlockStyle}>
            <h2 id="home-search-title" style={searchTitleStyle}>
              {t("search.title")}
            </h2>
            <PortalSearchBar label={t("search.label")} placeholder={t("search.placeholder")} />
            <div style={{ marginTop: "1.25rem" }}>
              <p style={popularLabelStyle} id="popular-searches-label">
                {t("search.popularLabel")}
              </p>
              <ul style={popularListStyle} aria-labelledby="popular-searches-label">
                {ministryHome.popularSearches.map((search) => (
                  <li key={search.key}>
                    <SearchSuggestionTag
                      label={t(`search.popular.${search.key}`)}
                      href={search.href}
                    />
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* 02 — Que souhaitez-vous faire ?: the functional heart of the page,
          concrete frequent actions each mapped to a real destination. */}
      <section className="gov-section gov-section--subtle" aria-labelledby="actions-title">
        <div className="gov-section__container">
          <div className="gov-section__header">
            <div>
              <p className="gov-kicker">{t("actions.kicker")}</p>
              <h2 id="actions-title" className="gov-section__title">
                {t("actions.title")}
              </h2>
              <p className="gov-lead">{t("actions.lead")}</p>
            </div>
          </div>
          <div className="fr-grid-row fr-grid-row--gutters">
            {ministryHome.actions.map((item) => (
              <div key={item.key} className="fr-col-12 fr-col-md-6 fr-col-lg-4">
                <LinkTile
                  title={t(`actions.items.${item.key}.title`)}
                  desc={t(item.descKey)}
                  href={item.href}
                  iconId={item.iconId}
                />
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* 03 — Mon espace: the personal space of the ministry. MyGouv provides
          the identity/SSO; this section presents the space, not MyGouv. */}
      <section className="gov-section" aria-labelledby="espace-title">
        <div className="gov-section__container" style={{ maxWidth: "64rem" }}>
          <MonEspace />
        </div>
      </section>

      {/* 04 — Informations importantes: operational alerts (échéance,
          changement réglementaire, mesure, réforme…). They come before
          general news when an item has an operational importance. */}
      <section className="gov-section gov-section--subtle" aria-labelledby="alerts-title">
        <div className="gov-section__container">
          <div className="gov-section__header">
            <div>
              <p className="gov-kicker">{t("alerts.kicker")}</p>
              <h2 id="alerts-title" className="gov-section__title">
                {t("alerts.title")}
              </h2>
              <p className="gov-lead">{t("alerts.lead")}</p>
            </div>
          </div>
          <ul className="fr-grid-row fr-grid-row--gutters" role="list">
            {ministryHome.importantInfo.map((item) => (
              <li key={item.key} className="fr-col-12 fr-col-md-6 fr-col-lg-4">
                <a href={item.href} style={teaserCardStyle}>
                  <span
                    className={item.iconId}
                    aria-hidden="true"
                    style={{ fontSize: "1.375rem", lineHeight: 1, color: "var(--ads-color-primary)" }}
                  />
                  <span style={{ display: "flex", flexDirection: "column", gap: "0.375rem" }}>
                    <span style={teaserTagStyle}>{t(`alerts.items.${item.key}.tag`)}</span>
                    <span style={teaserTitleStyle}>{t(`alerts.items.${item.key}.title`)}</span>
                    <span style={teaserDescStyle}>{t(`alerts.items.${item.key}.desc`)}</span>
                  </span>
                  <span
                    className="fr-icon-arrow-right-line"
                    aria-hidden="true"
                    style={teaserArrowStyle}
                  />
                </a>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* 05 — L'économie d'Astoria en chiffres: synthetic view of the
          country's economy, driven by the indicators config (the seam where a
          future data source plugs in). */}
      <section className="gov-section" aria-labelledby="stats-title">
        <div className="gov-section__container">
          <div className="gov-section__header">
            <div>
              <p className="gov-kicker">{t("stats.kicker")}</p>
              <h2 id="stats-title" className="gov-section__title">
                {t("stats.title")}
              </h2>
              <p className="gov-lead">{t("stats.lead")}</p>
            </div>
            <CtaButtonsGroup
              buttons={[
                {
                  children: t("stats.exploreLink"),
                  href: STATS_PATH,
                  priority: "secondary",
                  iconId: "fr-icon-arrow-right-line",
                },
              ]}
            />
          </div>
          <ul className="fr-grid-row fr-grid-row--gutters" role="list">
            {ministryHome.indicators.map((indicator) => (
              <li key={indicator.key} className="fr-col-12 fr-col-md-6 fr-col-lg-4">
                <a href={indicator.href} style={teaserCardStyle}>
                  <span style={figureValueStyle}>{indicator.value}</span>
                  <span style={figureLabelStyle}>{t(`stats.items.${indicator.key}.label`)}</span>
                  <span
                    className="fr-icon-arrow-right-line"
                    aria-hidden="true"
                    style={teaserArrowStyle}
                  />
                </a>
              </li>
            ))}
          </ul>
          <p className="fr-text--sm" style={{ color: "var(--ads-color-text-muted)" }}>
            {t("stats.note")}
          </p>
        </div>
      </section>

      {/* 06 — Budget de l'État: the State budget made understandable — total,
          revenue, expenditure, balance and debt. Sober on purpose: no chart
          until the real budget data is available. */}
      <section className="gov-section gov-section--subtle" aria-labelledby="budget-title">
        <div className="gov-section__container">
          <div className="gov-section__header">
            <div>
              <p className="gov-kicker">{t("budget.kicker")}</p>
              <h2 id="budget-title" className="gov-section__title">
                {t("budget.title")}
              </h2>
              <p className="gov-lead">{t("budget.lead")}</p>
            </div>
          </div>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr",
              gap: "1.75rem",
              maxWidth: "72rem",
              padding: "1.75rem",
              background: "var(--ads-color-background)",
              border: "1px solid var(--ads-color-border)",
            }}
          >
            <div>
              <p style={{ margin: "0 0 0.375rem", fontSize: "0.875rem", fontWeight: 600, color: "var(--ads-color-text-muted)" }}>
                {t("budget.totalLabel")}
              </p>
              <p style={{ margin: "0 0 1.5rem", fontSize: "clamp(1.75rem, 3.5vw, 2.5rem)", lineHeight: 1.2, fontWeight: 700 }}>
                {ministryHome.budget.total.value}
              </p>
              <CtaButtonsGroup
                buttons={[
                  {
                    children: t("budget.cta"),
                    href: ministryHome.budget.total.href,
                    priority: "secondary",
                    iconId: "fr-icon-arrow-right-line",
                  },
                ]}
              />
            </div>
            <ul className="fr-grid-row fr-grid-row--gutters" role="list">
              {ministryHome.budget.figures.map((figure) => (
                <li key={figure.key} className="fr-col-12 fr-col-sm-6 fr-col-lg-3">
                  <a
                    href={figure.href}
                    style={{
                      display: "flex",
                      flexDirection: "column",
                      gap: "0.25rem",
                      height: "100%",
                      padding: "0.875rem 1rem",
                      background: "var(--ads-color-surface-muted)",
                      border: "1px solid var(--ads-color-border)",
                      textDecoration: "none",
                      color: "var(--ads-color-text)",
                    }}
                  >
                    <span style={{ fontSize: "0.8125rem", fontWeight: 600, color: "var(--ads-color-text-muted)" }}>
                      {t(`budget.figures.${figure.key}`)}
                    </span>
                    <span style={{ fontSize: "1.125rem", lineHeight: 1.3, fontWeight: 700 }}>
                      {figure.value}
                    </span>
                  </a>
                </li>
              ))}
            </ul>
          </div>
          <p className="fr-text--sm" style={{ color: "var(--ads-color-text-muted)" }}>
            {t("budget.note")}
          </p>
        </div>
      </section>

      {/* 07 — Actualités: one featured article, several secondary ones. News
          comes after usages, figures and the budget — it no longer dominates
          the page. */}
      <section className="gov-section" aria-labelledby="news-title">
        <div className="gov-section__container">
          <div className="gov-section__header">
            <div>
              <p className="gov-kicker">{t("news.kicker")}</p>
              <h2 id="news-title" className="gov-section__title">
                {t("news.title")}
              </h2>
              <p className="gov-lead">{t("news.lead")}</p>
            </div>
            <CtaButtonsGroup
              buttons={[
                {
                  children: t("news.allLink"),
                  href: NEWS_PATH,
                  priority: "secondary",
                  iconId: "fr-icon-arrow-right-line",
                },
              ]}
            />
          </div>
          <div className="fr-grid-row fr-grid-row--gutters">
            <div className="fr-col-12 fr-col-lg-7">
              <ArticleCard
                title={t(ministryHome.news.featured.titleKey)}
                desc={
                  ministryHome.news.featured.textKey
                    ? t(ministryHome.news.featured.textKey)
                    : undefined
                }
                tag={t(ministryHome.news.featured.tagKey)}
                date={t(ministryHome.news.featured.dateKey)}
                href={ministryHome.news.featured.href}
                size="large"
              />
            </div>
            <div className="fr-col-12 fr-col-lg-5">
              <ul style={cardListStyle}>
                {ministryHome.news.secondary.map((article) => (
                  <li key={article.href}>
                    <ArticleCard
                      title={t(article.titleKey)}
                      tag={t(article.tagKey)}
                      date={t(article.dateKey)}
                      href={article.href}
                      size="small"
                    />
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* 08 — Ressources: compact shortcut to the most consulted documentary
          content of “Données & Ressources”. Not a sixth domain and not a
          second navigation — labels reuse the header vocabulary. */}
      <section className="gov-section gov-section--subtle" aria-labelledby="resources-title">
        <div className="gov-section__container">
          <div className="gov-section__header">
            <div>
              <p className="gov-kicker">{t("resources.kicker")}</p>
              <h2 id="resources-title" className="gov-section__title">
                {t("resources.title")}
              </h2>
              <p className="gov-lead">{t("resources.lead")}</p>
            </div>
            <CtaButtonsGroup
              buttons={[
                {
                  children: t("resources.allLink"),
                  href: DATA_PATH,
                  priority: "secondary",
                  iconId: "fr-icon-arrow-right-line",
                },
              ]}
            />
          </div>
          <ul role="list" style={linkListStyle}>
            {ministryHome.resources.map((item) => (
              <li key={item.key}>
                <a
                  href={item.href}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "space-between",
                    gap: "1rem",
                    padding: "1rem 1.25rem",
                    fontWeight: 600,
                    textDecoration: "none",
                    color: "var(--ads-color-text)",
                    border: "1px solid var(--ads-color-border)",
                    borderTop: "none",
                    background: "var(--ads-color-background)",
                  }}
                >
                  {tNavPanel(item.key)}
                  <span className="fr-icon-arrow-right-line" aria-hidden="true" />
                </a>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* 09 — Le Ministère de l'Économie et des Finances: discreet
          institutional closing. The user first does, then understands, then
          discovers the institution. */}
      <section className="gov-section" aria-labelledby="ministry-title">
        <div className="gov-section__container" style={heroContainerStyle}>
          <p className="gov-kicker">{t("ministry.kicker")}</p>
          <h2 id="ministry-title" className="gov-section__title">
            {t("ministry.title")}
          </h2>
          <p
            style={{
              margin: "0 auto 1.5rem",
              maxWidth: "42rem",
              fontSize: "0.9375rem",
              lineHeight: 1.7,
              color: "var(--ads-color-text-muted)",
            }}
          >
            {t("ministry.lead")}
          </p>
          <ul
            role="list"
            style={{
              ...popularListStyle,
              justifyContent: "center",
              marginTop: "1.5rem",
              gap: "0.625rem 1.75rem",
            }}
          >
            {ministryHome.ministry.map((link) => (
              <li key={link.key}>
                <a href={link.href} style={{ fontWeight: 600, textUnderlineOffset: "0.2em" }}>
                  {t(`ministry.links.${link.key}.title`)}
                  <span className="fr-icon-arrow-right-line" aria-hidden="true" />
                </a>
              </li>
            ))}
          </ul>
        </div>
      </section>
    </>
  );
}
