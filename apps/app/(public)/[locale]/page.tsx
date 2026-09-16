import type { Metadata } from "next";
import type { ReactNode } from "react";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { ADSSection, ADSContainer } from "@codegouvaor/react-ads/ads";
import { ButtonsGroup } from "@codegouvaor/react-ads/ButtonsGroup";
import { Card } from "@codegouvaor/react-ads/Card";
import { Highlight } from "@codegouvaor/react-ads/Highlight";
import { Notice } from "@codegouvaor/react-ads/Notice";
import { TagsGroup } from "@codegouvaor/react-ads/TagsGroup";
import { Tile } from "@codegouvaor/react-ads/Tile";
import type { TagProps } from "@codegouvaor/react-ads/Tag";
import type { ButtonProps } from "@codegouvaor/react-ads/Button";
import { localizedAlternates, resolveLocaleParam } from "@/lib/localized-metadata";
import {
  homeActivityHref,
  homeActivityExternal,
  homeActivityItems,
  homeContributeCtaHref,
  homeContributeItems,
  homeDocumentationCtaHref,
  homeDocumentationItems,
  homeEcosystemPlatforms,
  homeExploreDoors,
  homeFinalCtaLinks,
  homeOrganizationItems,
  homeOrganizationsCtaHref,
  homeProjectsCtaHref,
  homeProjectsModel,
  homeResourcesCtaHref,
  homeResourcesItems,
} from "@/lib/home-content";
import { portalPaths } from "@/lib/site-structure";

const HOME_PATH = "/";

type PageProps = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);

  const tHome = await getTranslations({ locale, namespace: "home" });
  const tMeta = await getTranslations({ locale, namespace: "meta" });

  return {
    title: { absolute: tHome("meta.title") },
    description: tMeta("description"),
    ...localizedAlternates(locale, HOME_PATH),
  };
}

/**
 * Homepage of CODE — the public developer platform of Astoria's
 * administration.
 *
 * CODE is not a Git forge: it is a common entry point to the projects,
 * organisations, documentation, standards, APIs, SDKs and technical resources
 * of the public sector, while remaining interoperable with the development
 * platforms it references (GitHub, GitLab, Giteria…). The conceptual model is
 * `Code Project → Repository Binding → GitHub / GitLab / Giteria`: a project
 * is independent from the forge that hosts its code.
 *
 * The page is organised in ten sections:
 *
 *   01 — Hero           « Le numérique public, en code. »
 *   02 — Explorer Code  the six main doorways of the platform
 *   03 — Les projets    the public software projects (interoperable model)
 *   04 — Écosystème     the integrated forges, an open ecosystem
 *   05 — Développer     the technical resources to build public services
 *   06 — Documentation  the developer documentation platform
 *   07 — Organisations  the organisations, teams and communities
 *   08 — Contribuer     the ways to contribute
 *   09 — Activité       the recent life of the ecosystem
 *   10 — CTA final      « Construire le numérique public. »
 *
 * The content is driven by the `home-content` data module (keys + hrefs) and
 * the message catalogs (`home.*`); this component only decides the structure.
 * No project, organisation, release or statistic is invented: sections whose
 * content is not yet available use a static institutional presentation.
 */
export default async function HomePage({ params }: PageProps) {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);
  setRequestLocale(locale);

  const t = await getTranslations({ locale, namespace: "home" });
  const tCommon = await getTranslations({ locale, namespace: "common" });

  /** Anchor button CTA used by several sections. */
  const cta = (
    buttons: ReadonlyArray<Pick<ButtonProps, "children" | "linkProps" | "priority" | "iconId" | "iconPosition">>
  ): ReactNode => (
    <div className="gov-code-section-cta">
      <ButtonsGroup inlineLayoutWhen="sm and up" buttons={[...buttons] as [ButtonProps, ...ButtonProps[]]} />
    </div>
  );

  return (
    <>
      {/* 01 — Hero: the front door of the platform. Two-column layout (text +
          developer illustration), like the reference portal; the global search
          lives in the header, so the hero stays a sober institutional
          statement with the two main entry points. */}
      <ADSSection tone="subtle">
        <ADSContainer size="wide">
          <div className="fr-grid-row fr-grid-row--gutters fr-grid-row--middle">
            <div className="fr-col-12 fr-col-md-7">
              <p className="gov-kicker">{t("hero.kicker")}</p>
              <h1 id="home-hero-title">{t("hero.title")}</h1>
              <p className="gov-lead">{t("hero.lead")}</p>
              <div style={{ marginTop: "1.5rem" }}>
                <ButtonsGroup
                  alignment="left"
                  inlineLayoutWhen="sm and up"
                  buttons={[
                    {
                      children: t("hero.primaryCta"),
                      linkProps: { href: homeProjectsCtaHref },
                      priority: "primary",
                      iconId: "fr-icon-arrow-right-line",
                      iconPosition: "right",
                    },
                    {
                      children: t("hero.secondaryCta"),
                      linkProps: { href: portalPaths.documentation },
                      priority: "secondary",
                    },
                  ]}
                />
              </div>
            </div>
            <div className="fr-col-12 fr-col-md-5">
              <img src="/Programmer.png" alt="" className="fr-responsive-img" />
            </div>
          </div>
        </ADSContainer>
      </ADSSection>

      {/* 02 — Explorer Code: the six doorways of the platform. */}
      <section className="gov-section" aria-labelledby="home-explore-title">
        <div className="gov-section__container">
          <p className="gov-kicker">{t("explore.kicker")}</p>
          <h2 id="home-explore-title" className="gov-section__title">
            {t("explore.title")}
          </h2>
          <p className="gov-lead">{t("explore.lead")}</p>
          <ul className="fr-grid-row fr-grid-row--gutters gov-code-gridlist">
            {homeExploreDoors.map((door) => (
              <li key={door.key} className="fr-col-12 fr-col-sm-6 fr-col-lg-4">
                <Tile
                  title={t(`explore.items.${door.key}.title`)}
                  desc={t(`explore.items.${door.key}.desc`)}
                  titleAs="h3"
                  linkProps={{ href: door.href }}
                  pictogram={
                    <span
                      className={`gov-code-tile-pictogram ${door.iconId}`}
                      aria-hidden="true"
                    />
                  }
                />
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* 03 — Les projets publics: at the centre of the platform. The model
          (Projet → Liaison de dépôt → Forge) is explained statically while
          no project is referenced yet; no project is invented. */}
      <section className="gov-section gov-section--subtle" aria-labelledby="home-projects-title">
        <div className="gov-section__container">
          <p className="gov-kicker">{t("projects.kicker")}</p>
          <h2 id="home-projects-title" className="gov-section__title">
            {t("projects.title")}
          </h2>
          <p className="gov-lead">{t("projects.lead")}</p>

          <h3 className="gov-section__subtitle">{t("projects.modelTitle")}</h3>
          <p className="gov-lead gov-code-model__text">{t("projects.modelText")}</p>
          <ul className="fr-grid-row fr-grid-row--gutters gov-code-gridlist">
            {homeProjectsModel.map((item) => (
              <li key={item.key} className="fr-col-12 fr-col-md-4">
                <Card
                  title={t(`projects.items.${item.key}.title`)}
                  desc={t(`projects.items.${item.key}.desc`)}
                  titleAs="h4"
                  start={
                    <span
                      className={`gov-code-card__icon ${item.iconId}`}
                      aria-hidden="true"
                    />
                  }
                  shadow
                />
              </li>
            ))}
          </ul>

          <div className="gov-code-empty">
            <Notice
              title={t("projects.emptyTitle")}
              description={t("projects.emptyText")}
            />
          </div>
          {cta([
            {
              children: t("projects.cta"),
              linkProps: { href: homeProjectsCtaHref },
              priority: "primary",
              iconId: "fr-icon-arrow-right-line",
              iconPosition: "right",
            },
          ])}
        </div>
      </section>

      {/* 04 — Un écosystème ouvert: CODE references the existing forges, it
          does not replace them. */}
      <section className="gov-section" aria-labelledby="home-ecosystem-title">
        <div className="gov-section__container gov-code-ecosystem">
          <p className="gov-kicker">{t("ecosystem.kicker")}</p>
          <h2 id="home-ecosystem-title" className="gov-section__title">
            {t("ecosystem.title")}
          </h2>
          <p className="gov-lead">{t("ecosystem.lead")}</p>
          <Highlight>{t("ecosystem.text")}</Highlight>
          <div className="gov-code-ecosystem__platforms">
            <h3 className="gov-section__subtitle">{t("ecosystem.platformsTitle")}</h3>
            <TagsGroup
              tags={homeEcosystemPlatforms.map(
                (platform): TagProps => {
                  if (platform.href && platform.external) {
                    return {
                      children: t(`ecosystem.platforms.${platform.key}`),
                      as: "a",
                      linkProps: {
                        href: platform.href,
                        target: "_blank",
                        rel: "noopener noreferrer",
                        title: tCommon("openNewWindow"),
                      },
                      iconId: "fr-icon-external-link-line",
                    };
                  }
                  return {
                    children: t(`ecosystem.platforms.${platform.key}`),
                    as: "span",
                  };
                }
              ) as [TagProps, ...TagProps[]]}
            />
          </div>
        </div>
      </section>

      {/* 05 — Développer les services publics: the technical building blocks. */}
      <section className="gov-section gov-section--subtle" aria-labelledby="home-resources-title">
        <div className="gov-section__container">
          <p className="gov-kicker">{t("resources.kicker")}</p>
          <h2 id="home-resources-title" className="gov-section__title">
            {t("resources.title")}
          </h2>
          <p className="gov-lead">{t("resources.lead")}</p>
          <ul className="fr-grid-row fr-grid-row--gutters gov-code-gridlist">
            {homeResourcesItems.map((item) => (
              <li key={item.key} className="fr-col-12 fr-col-sm-6 fr-col-lg-4">
                <Card
                  title={t(`resources.items.${item.key}.title`)}
                  desc={t(`resources.items.${item.key}.desc`)}
                  titleAs="h3"
                  enlargeLink
                  linkProps={{ href: item.href }}
                  start={
                    <span
                      className={`gov-code-card__icon ${item.iconId}`}
                      aria-hidden="true"
                    />
                  }
                  shadow
                />
              </li>
            ))}
          </ul>
          {cta([
            {
              children: t("resources.cta"),
              linkProps: { href: homeResourcesCtaHref },
              priority: "secondary",
              iconId: "fr-icon-arrow-right-line",
              iconPosition: "right",
            },
          ])}
        </div>
      </section>

      {/* 06 — La documentation: a developer documentation platform. */}
      <section className="gov-section" aria-labelledby="home-documentation-title">
        <div className="gov-section__container">
          <p className="gov-kicker">{t("documentation.kicker")}</p>
          <h2 id="home-documentation-title" className="gov-section__title">
            {t("documentation.title")}
          </h2>
          <p className="gov-lead">{t("documentation.lead")}</p>
          <ul className="fr-grid-row fr-grid-row--gutters gov-code-gridlist">
            {homeDocumentationItems.map((item) => (
              <li key={item.key} className="fr-col-12 fr-col-sm-6 fr-col-lg-4">
                <Card
                  title={t(`documentation.items.${item.key}.title`)}
                  desc={t(`documentation.items.${item.key}.desc`)}
                  titleAs="h3"
                  enlargeLink
                  linkProps={{ href: item.href }}
                  start={
                    <span
                      className={`gov-code-card__icon ${item.iconId}`}
                      aria-hidden="true"
                    />
                  }
                  shadow
                />
              </li>
            ))}
          </ul>
          {cta([
            {
              children: t("documentation.cta"),
              linkProps: { href: homeDocumentationCtaHref },
              priority: "secondary",
              iconId: "fr-icon-arrow-right-line",
              iconPosition: "right",
            },
          ])}
        </div>
      </section>

      {/* 07 — Organisations et communautés. */}
      <section className="gov-section gov-section--subtle" aria-labelledby="home-organizations-title">
        <div className="gov-section__container">
          <p className="gov-kicker">{t("organizations.kicker")}</p>
          <h2 id="home-organizations-title" className="gov-section__title">
            {t("organizations.title")}
          </h2>
          <p className="gov-lead">{t("organizations.lead")}</p>
          <ul className="fr-grid-row fr-grid-row--gutters gov-code-gridlist">
            {homeOrganizationItems.map((item) => (
              <li key={item.key} className="fr-col-12 fr-col-sm-6 fr-col-lg-4">
                <Card
                  title={t(`organizations.items.${item.key}.title`)}
                  desc={t(`organizations.items.${item.key}.desc`)}
                  titleAs="h3"
                  start={
                    <span
                      className={`gov-code-card__icon ${item.iconId}`}
                      aria-hidden="true"
                    />
                  }
                  shadow
                />
              </li>
            ))}
          </ul>
          {cta([
            {
              children: t("organizations.cta"),
              linkProps: { href: homeOrganizationsCtaHref },
              priority: "secondary",
              iconId: "fr-icon-arrow-right-line",
              iconPosition: "right",
            },
          ])}
        </div>
      </section>

      {/* 08 — Contribuer: the many ways to contribute, developer-oriented. */}
      <section className="gov-section" aria-labelledby="home-contribute-title">
        <div className="gov-section__container">
          <p className="gov-kicker">{t("contribute.kicker")}</p>
          <h2 id="home-contribute-title" className="gov-section__title">
            {t("contribute.title")}
          </h2>
          <p className="gov-lead">{t("contribute.lead")}</p>
          <ul className="fr-grid-row fr-grid-row--gutters gov-code-gridlist">
            {homeContributeItems.map((item) => (
              <li key={item.key} className="fr-col-12 fr-col-sm-6 fr-col-lg-3">
                <Tile
                  title={t(`contribute.items.${item.key}.title`)}
                  desc={t(`contribute.items.${item.key}.desc`)}
                  titleAs="h3"
                  pictogram={
                    <span
                      className={`gov-code-tile-pictogram ${item.iconId}`}
                      aria-hidden="true"
                    />
                  }
                  noBorder
                />
              </li>
            ))}
          </ul>
          {cta([
            {
              children: t("contribute.cta"),
              linkProps: { href: homeContributeCtaHref },
              priority: "primary",
              iconId: "fr-icon-arrow-right-line",
              iconPosition: "right",
            },
          ])}
        </div>
      </section>

      {/* 09 — Actualités et activité: the life of the ecosystem. No event is
          invented — the news channel of the platform is the public repository
          releases feed while no editorial flow exists yet. */}
      <section className="gov-section gov-section--subtle" aria-labelledby="home-activity-title">
        <div className="gov-section__container">
          <p className="gov-kicker">{t("activity.kicker")}</p>
          <h2 id="home-activity-title" className="gov-section__title">
            {t("activity.title")}
          </h2>
          <p className="gov-lead">{t("activity.lead")}</p>
          <ul className="fr-grid-row fr-grid-row--gutters gov-code-gridlist">
            {homeActivityItems.map((item) => (
              <li key={item.key} className="fr-col-12 fr-col-sm-6 fr-col-lg-4">
                <Card
                  title={t(`activity.items.${item.key}.title`)}
                  desc={t(`activity.items.${item.key}.desc`)}
                  titleAs="h3"
                  start={
                    <span
                      className={`gov-code-card__icon ${item.iconId}`}
                      aria-hidden="true"
                    />
                  }
                  shadow
                />
              </li>
            ))}
          </ul>
          {cta([
            {
              children: t("activity.cta"),
              linkProps: homeActivityExternal
                ? {
                    href: homeActivityHref,
                    target: "_blank",
                    rel: "noopener noreferrer",
                    title: tCommon("openNewWindow"),
                  }
                : { href: homeActivityHref },
              priority: "secondary",
              iconId: "fr-icon-external-link-line",
              iconPosition: "right",
            },
          ])}
        </div>
      </section>

      {/* 10 — CTA final: the natural transition to the institutional footer. */}
      <section
        className="gov-section gov-section--tinted gov-code-final"
        aria-labelledby="home-final-title"
      >
        <div className="gov-section__container gov-code-final__inner">
          <h2 id="home-final-title" className="gov-section__title">
            {t("finalCta.title")}
          </h2>
          <p className="gov-code-final__text">{t("finalCta.text")}</p>
          <ButtonsGroup
            alignment="center"
            inlineLayoutWhen="sm and up"
            buttons={[
              {
                children: t("finalCta.cta1"),
                linkProps: { href: homeFinalCtaLinks[0].href },
                priority: "primary",
                iconId: "fr-icon-arrow-right-line",
                iconPosition: "right",
              },
              {
                children: t("finalCta.cta2"),
                linkProps: { href: homeFinalCtaLinks[1].href },
                priority: "secondary",
              },
              {
                children: t("finalCta.cta3"),
                linkProps: { href: homeFinalCtaLinks[2].href },
                priority: "secondary",
              },
            ]}
          />
        </div>
      </section>
    </>
  );
}