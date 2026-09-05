import type { Metadata } from "next";
import type { ReactNode } from "react";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { Accordion } from "@codegouvaor/react-ads/Accordion";
import { localizedAlternates, resolveLocaleParam } from "@/lib/localized-metadata";
import { Link } from "@/i18n/navigation";
import { lifecycleThemes } from "@/lib/site-structure";
import {
  homeContributions,
  homeDoors,
  homeFollowLinks,
  homeFoundations,
  homeResources,
} from "@/lib/home-content";

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
 * Homepage of the Pôle open-source et numérique commun — CODE.GOUV.AOR.
 *
 * Editorial content model of the reference open-source portal (code.gouv.fr),
 * adapted to the CODE ecosystem, organised in one mission hero followed by
 * seven sections:
 *
 *   01 — Héro mission (Construire le numérique public + illustration)
 *   02 — « Des standards pour et par les administrations » (3 portes)
 *   03 — Bande éditoriale « Le pôle open-source et numérique commun »
 *   04 — « Explorer la plateforme » (les 6 thèmes du cycle de vie)
 *   05 — Fondations & références
 *   06 — Ressources de l'écosystème
 *   07 — « Vos contributions sont les bienvenues ! » (accordéons)
 *   08 — Bande « Restez informé / réseaux »
 *
 * Text lives in the message catalogs (`home.*`); this component only decides
 * the structure, the data (`home-content`) and the destinations.
 */
export default async function HomePage({ params }: PageProps) {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);
  setRequestLocale(locale);

  const t = await getTranslations({ locale, namespace: "home" });
  const tPrimaryNav = await getTranslations({ locale, namespace: "nav.primary" });
  const tNavPanel = await getTranslations({ locale, namespace: "nav.panel" });
  const tCommon = await getTranslations({ locale, namespace: "common" });

  /** Accent words of a heading — same weight, coloured like the reference. */
  const accent = (chunks: ReactNode) => <span className="gov-code-accent">{chunks}</span>;
  /** Action words of the mission lead — bold, like the reference chapeau. */
  const strong = (chunks: ReactNode) => <strong>{chunks}</strong>;

  return (
    <>
      {/* 01 — Héro mission: text (md-6) + illustration (md-4), like the
          reference portal. Search lives in the header. */}
      <section
        className="gov-section gov-section--tinted gov-code-hero"
        aria-labelledby="home-hero-title"
      >
        <div className="gov-section__container">
          <div className="fr-grid-row fr-grid-row--center">
            <div className="fr-col-10 fr-col-md-6">
              <h1 id="home-hero-title" className="gov-code-hero__title">
                {t("hero.title")}
              </h1>
              <p className="gov-lead gov-code-hero__lead">
                {t.rich("hero.lead", { strong })}
              </p>
            </div>
            <div className="fr-hidden fr-unhidden-md fr-col-md-4 gov-code-hero__visual">
              {/* Decorative illustration of the mission band (same artwork
                  as the reference portal), hidden below 48em. */}
              <img src="/Programmer.png" alt="" />
            </div>
          </div>
        </div>
      </section>

      {/* 02 — Pour / Par / Avec les agents : the three doorways */}
      <section className="gov-section" aria-labelledby="home-doors-title">
        <div className="gov-section__container">
          <h2 id="home-doors-title" className="gov-section__title">
            {t.rich("doors.title", { accent })}
          </h2>
          <ul className="fr-grid-row fr-grid-row--gutters gov-code-gridlist">
            {homeDoors.map((door) => (
              <li key={door.key} className="fr-col-12 fr-col-md-4">
                <Link className="gov-code-card" href={door.href}>
                  <h3>{t(`doors.items.${door.key}.title`)}</h3>
                  <p>{t(`doors.items.${door.key}.desc`)}</p>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* 03 — Bande éditoriale : Le pôle open-source et numérique commun */}
      <section
        className="gov-section gov-section--subtle"
        aria-labelledby="home-band-title"
      >
        <div className="gov-section__container">
          <div className="gov-code-narrow">
            <h2 id="home-band-title" className="gov-section__title">
              {t("band.title")}
            </h2>
            <p className="gov-lead gov-code-band__text">{t("band.text")}</p>
          </div>
        </div>
      </section>

      {/* 04 — Explorer la plateforme : the six lifecycle themes */}
      <section className="gov-section" aria-labelledby="home-explore-title">
        <div className="gov-section__container">
          <h2 id="home-explore-title" className="gov-section__title">
            {t("explore.title")}
          </h2>
          <p className="gov-lead">{t("explore.lead")}</p>
          <ul className="fr-grid-row fr-grid-row--gutters gov-code-gridlist">
            {lifecycleThemes.map((theme) => (
              <li key={theme.labelKey} className="fr-col-12 fr-col-md-6 fr-col-lg-4">
                <Link className="gov-code-lifecycle-link" href={theme.href}>
                  <span>{tPrimaryNav(theme.labelKey)}</span>
                  <span className="fr-icon-arrow-right-line" aria-hidden="true" />
                </Link>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* 05 — Fondations : the cross-cutting references of CODE */}
      <section
        className="gov-section gov-section--subtle"
        aria-labelledby="home-foundations-title"
      >
        <div className="gov-section__container">
          <h2 id="home-foundations-title" className="gov-section__title">
            {t.rich("foundations.title", { accent })}
          </h2>
          <p className="gov-lead">{t("foundations.lead")}</p>
          <ul className="fr-grid-row fr-grid-row--gutters gov-code-gridlist">
            {homeFoundations.map((foundation) => (
              <li key={foundation.key} className="fr-col-12 fr-col-sm-6 fr-col-lg-4">
                <Link className="gov-code-card" href={foundation.href}>
                  <h3>{tNavPanel(foundation.key)}</h3>
                  <p>{t(`foundations.items.${foundation.key}`)}</p>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* 06 — Ressources de l'écosystème */}
      <section className="gov-section" aria-labelledby="home-resources-title">
        <div className="gov-section__container">
          <h2 id="home-resources-title" className="gov-section__title">
            {t("resources.title")}
          </h2>
          <p className="gov-lead">{t("resources.lead")}</p>
          <ul className="fr-grid-row fr-grid-row--gutters gov-code-gridlist">
            {homeResources.map((resource) => {
              const common = (
                <>
                  <h3>{tNavPanel(resource.labelKey)}</h3>
                  <p>{t(`resources.items.${resource.key}`)}</p>
                  <span className="gov-code-card__action">
                    {t("resources.discover")}
                    <span
                      className={
                        resource.external
                          ? "fr-icon-external-link-line"
                          : "fr-icon-arrow-right-line"
                      }
                      aria-hidden="true"
                    />
                  </span>
                </>
              );

              return (
                <li key={resource.key} className="fr-col-12 fr-col-sm-6 fr-col-lg-4">
                  {resource.external ? (
                    <Link
                      className="gov-code-card"
                      href={resource.href}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {common}
                    </Link>
                  ) : (
                    <Link className="gov-code-card" href={resource.href}>
                      {common}
                    </Link>
                  )}
                </li>
              );
            })}
          </ul>
        </div>
      </section>

      {/* 07 — Vos contributions sont les bienvenues ! */}
      <section
        className="gov-section gov-section--subtle"
        aria-labelledby="home-contributions-title"
      >
        <div className="gov-section__container">
          <div className="gov-code-narrow">
            <h2 id="home-contributions-title" className="gov-section__title">
              {t("contributions.title")}
            </h2>
            <div className="fr-accordions-group gov-code-accordions">
              {homeContributions.map((item) => (
                <Accordion
                  key={item.key}
                  titleAs="h3"
                  label={t(`contributions.items.${item.key}.question`)}
                >
                  <p className="gov-code-accordions__answer">
                    {t.rich(`contributions.items.${item.key}.answer`, {
                      link: (chunks) => <Link href={item.href}>{chunks}</Link>,
                    })}
                  </p>
                </Accordion>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* 08 — Bande follow : restez informé + réseaux */}
      <section
        className="gov-section gov-section--tinted gov-code-follow"
        aria-labelledby="home-follow-title"
      >
        <div className="gov-section__container">
          <div className="fr-grid-row fr-grid-row--gutters fr-grid-row--middle">
            <div className="fr-col-12 fr-col-md-7">
              <h2 id="home-follow-title" className="gov-code-follow__title">
                {t("follow.newsletterTitle")}
              </h2>
              <p className="gov-code-follow__text">{t("follow.newsletterText")}</p>
              {/* DSFR appends the external-link icon automatically on
                  target="_blank" links — no manual icon, no duplicate. */}
              <Link
                className="fr-btn"
                href={homeFollowLinks.releasesUrl}
                target="_blank"
                rel="noopener noreferrer"
                title={tCommon("openNewWindow")}
              >
                {t("follow.newsletterCta")}
              </Link>
            </div>
            <div className="fr-col-12 fr-col-md-5">
              <p className="gov-code-follow__social-title">{t("follow.socialTitle")}</p>
              <ul className="gov-code-follow__social">
                <li>
                  <Link
                    className="gov-code-follow__social-link"
                    href={homeFollowLinks.githubUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    title={tCommon("openNewWindow")}
                  >
                    {tNavPanel("github")}
                  </Link>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
