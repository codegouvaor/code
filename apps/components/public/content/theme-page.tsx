import type { CSSProperties, ReactNode } from "react";
import type { FrIconClassName } from "@codegouvaor/react-ads/fr";
import { Link } from "@/i18n/navigation";
import { CtaButtonsGroup, LinkTile, NoticeCallout, type CtaButton } from "./ads-fragments";

/**
 * Shared presentational layer of the BCA theme pages (Monnaie and, later,
 * the other editorial subjects of the portal). Server component: it only
 * passes serializable props (strings, hrefs) to the ADS client boundaries.
 *
 * Every page of a theme is composed with `ThemeArticle`, which renders a
 * *localized* content object produced by `lib/monnaie-localize.ts`. The
 * message catalogs (`apps/messages/{fr,en}.json`) hold every display string;
 * the shared blocks below only decide how to arrange those strings, the same
 * way the home page composes its sections.
 */

/* ---- Shared style constants (ADS design tokens, no local stylesheet) ------ */

export const heroContainerStyle: CSSProperties = {
  maxWidth: "52rem",
  marginInline: "auto",
  textAlign: "center",
};

export const teaserCardStyle: CSSProperties = {
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

export const teaserTagStyle: CSSProperties = {
  fontSize: "0.75rem",
  fontWeight: 700,
  letterSpacing: "0.06em",
  textTransform: "uppercase",
  color: "var(--ads-color-primary)",
};

export const teaserTitleStyle: CSSProperties = {
  display: "block",
  fontSize: "1.0625rem",
  lineHeight: 1.35,
  fontWeight: 700,
};

export const teaserDescStyle: CSSProperties = {
  display: "block",
  fontSize: "0.875rem",
  lineHeight: 1.55,
  color: "var(--ads-color-text-muted)",
};

export const figureValueStyle: CSSProperties = {
  display: "block",
  fontSize: "clamp(1.5rem, 3vw, 2rem)",
  lineHeight: 1.2,
  fontWeight: 700,
};

export const figureLabelStyle: CSSProperties = {
  display: "block",
  fontSize: "0.9375rem",
  fontWeight: 600,
};

export const cardGridStyle: CSSProperties = {
  listStyle: "none",
  margin: "0",
  padding: "0",
};

export const iconBlockStyle: CSSProperties = {
  fontSize: "1.375rem",
  lineHeight: 1,
  color: "var(--ads-color-primary)",
};

export const linkListStyle: CSSProperties = {
  listStyle: "none",
  margin: "0",
  padding: "0",
  display: "grid",
  gap: "0",
  maxWidth: "72rem",
};

const denominationCardStyle: CSSProperties = {
  ...teaserCardStyle,
  textAlign: "center",
  alignItems: "center",
};

const denominationValueStyle: CSSProperties = {
  display: "block",
  fontSize: "clamp(2rem, 5vw, 2.75rem)",
  lineHeight: 1.1,
  fontWeight: 700,
};

const denominationUnitStyle: CSSProperties = {
  display: "block",
  fontSize: "0.875rem",
  fontWeight: 600,
  color: "var(--ads-color-text-muted)",
};

const stepBadgeStyle: CSSProperties = {
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  width: "2rem",
  height: "2rem",
  flexShrink: 0,
  borderRadius: "50%",
  background: "var(--ads-color-primary)",
  color: "var(--ads-color-background)",
  fontSize: "0.9375rem",
  fontWeight: 700,
};

const flowItemStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.5rem",
};

const flowListStyle: CSSProperties = {
  listStyle: "none",
  margin: 0,
  padding: 0,
  display: "flex",
  flexWrap: "wrap",
  alignItems: "center",
  gap: "0.5rem",
};

const flowNodeStyle: CSSProperties = {
  padding: "0.625rem 1rem",
  background: "var(--ads-color-background)",
  border: "1px solid var(--ads-color-border)",
  borderTop: "3px solid var(--ads-color-primary)",
  fontSize: "0.9375rem",
  fontWeight: 600,
  color: "var(--ads-color-text)",
};

const factsListStyle: CSSProperties = {
  margin: 0,
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(15rem, 1fr))",
  gap: "1px",
  background: "var(--ads-color-border)",
  border: "1px solid var(--ads-color-border)",
};

const factCellStyle: CSSProperties = {
  padding: "1rem 1.25rem",
  background: "var(--ads-color-background)",
};

const factLabelStyle: CSSProperties = {
  margin: "0 0 0.25rem",
  fontSize: "0.8125rem",
  fontWeight: 600,
  color: "var(--ads-color-text-muted)",
};

const factValueStyle: CSSProperties = {
  margin: 0,
  fontSize: "1.0625rem",
  lineHeight: 1.4,
  fontWeight: 600,
  color: "var(--ads-color-text)",
};

const heroStatStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "baseline",
  gap: "0.75rem",
  marginTop: "1.5rem",
  padding: "0.875rem 1.25rem",
  background: "var(--ads-color-background)",
  border: "1px solid var(--ads-color-border)",
  borderTop: "3px solid var(--ads-color-primary)",
  textAlign: "left",
};

/* ---- Blocks --------------------------------------------------------------- */

/**
 * Generic process/flow representation: a chain of labelled nodes separated by
 * arrows. Used for the monetary circuits (“BCA → banques → usagers → BCA”) —
 * a schema, never a fake chart. Wraps gracefully on small screens.
 */
export function FlowDiagram({
  items,
  caption,
}: {
  items: ReadonlyArray<{ key: string; label: string }>;
  caption?: string;
}) {
  return (
    <figure style={{ margin: 0 }}>
      <ol role="list" style={flowListStyle}>
        {items.map((item, index) => (
          <li key={item.key} style={flowItemStyle}>
            {index > 0 ? (
              <span
                className="fr-icon-arrow-right-line"
                aria-hidden="true"
                style={{ color: "var(--ads-color-primary)" }}
              />
            ) : null}
            <span style={flowNodeStyle}>{item.label}</span>
          </li>
        ))}
      </ol>
      {caption ? (
        <figcaption
          className="fr-text--sm"
          style={{ marginTop: "0.75rem", color: "var(--ads-color-text-muted)" }}
        >
          {caption}
        </figcaption>
      ) : null}
    </figure>
  );
}

/** Definition grid for institutional reference facts (name, unit, status…). */
export function FactList({
  items,
}: {
  items: ReadonlyArray<{ key: string; label: string; value: string }>;
}) {
  return (
    <dl style={factsListStyle}>
      {items.map((item) => (
        <div key={item.key} style={factCellStyle}>
          <dt style={factLabelStyle}>{item.label}</dt>
          <dd style={factValueStyle}>{item.value}</dd>
        </div>
      ))}
    </dl>
  );
}

export function ThemeHero({
  kicker,
  title,
  lead,
  actions,
  stat,
}: {
  kicker: string;
  title: string;
  lead: string;
  actions?: [CtaButton, ...CtaButton[]];
  stat?: { value: string; label: string };
}) {
  return (
    <section className="gov-section" aria-labelledby="theme-hero-title">
      <div className="gov-section__container" style={heroContainerStyle}>
        <p className="gov-kicker">{kicker}</p>
        <h1 id="theme-hero-title">{title}</h1>
        <p className="gov-lead">{lead}</p>
        {stat ? (
          <p style={heroStatStyle}>
            <span style={{ fontSize: "clamp(1.5rem, 3vw, 2rem)", lineHeight: 1.1, fontWeight: 700 }}>
              {stat.value}
            </span>
            <span style={{ fontSize: "0.9375rem", fontWeight: 600, color: "var(--ads-color-text-muted)" }}>
              {stat.label}
            </span>
          </p>
        ) : null}
        {actions ? <CtaButtonsGroup alignment="center" buttons={actions} /> : null}
      </div>
    </section>
  );
}

export function ThemeSection({
  id,
  kicker,
  title,
  lead,
  subtle,
  action,
  children,
}: {
  id: string;
  kicker: string;
  title: string;
  lead?: string;
  subtle?: boolean;
  action?: ReactNode;
  children?: ReactNode;
}) {
  return (
    <section
      className={subtle ? "gov-section gov-section--subtle" : "gov-section"}
      aria-labelledby={id}
    >
      <div className="gov-section__container">
        <div className="gov-section__header">
          <div>
            <p className="gov-kicker">{kicker}</p>
            <h2 id={id} className="gov-section__title">
              {title}
            </h2>
            {lead ? <p className="gov-lead">{lead}</p> : null}
          </div>
          {action}
        </div>
        {children}
      </div>
    </section>
  );
}

export type StatItem = {
  key: string;
  value: string;
  label: string;
  href?: string;
  iconId?: string;
};

/** Figure cards — the “en chiffres” pattern of the home page. */
export function StatGrid({
  items,
  perRow = 4,
}: {
  items: ReadonlyArray<StatItem>;
  perRow?: 3 | 4;
}) {
  const columnClass =
    perRow === 3 ? "fr-col-12 fr-col-md-6 fr-col-lg-4" : "fr-col-12 fr-col-md-6 fr-col-lg-3";
  return (
    <ul className="fr-grid-row fr-grid-row--gutters" role="list" style={cardGridStyle}>
      {items.map((item) => {
        const content = (
          <>
            {item.iconId ? (
              <span className={item.iconId} aria-hidden="true" style={iconBlockStyle} />
            ) : null}
            <span style={figureValueStyle}>{item.value}</span>
            <span style={figureLabelStyle}>{item.label}</span>
            <span
              className="fr-icon-arrow-right-line"
              aria-hidden="true"
              style={{ marginTop: "auto", alignSelf: "flex-end", fontSize: "1rem", color: "var(--ads-color-primary)" }}
            />
          </>
        );
        return (
          <li key={item.key} className={columnClass}>
            {item.href ? (
              <Link href={item.href} style={teaserCardStyle}>
                {content}
              </Link>
            ) : (
              <div style={teaserCardStyle}>{content}</div>
            )}
          </li>
        );
      })}
    </ul>
  );
}

/** Responsive data table of the analytical pages (overflow scroll on mobile). */
export function DataTable({
  caption,
  headers,
  rows,
  note,
}: {
  caption: string;
  headers: string[];
  rows: ReadonlyArray<ReadonlyArray<string | number>>;
  note?: string;
}) {
  return (
    <figure style={{ margin: 0 }}>
      <div style={{ overflowX: "auto" }}>
        <table
          style={{
            width: "100%",
            borderCollapse: "collapse",
            fontSize: "0.9375rem",
            lineHeight: 1.5,
          }}
        >
          <caption
            style={{
              textAlign: "left",
              padding: "0 0 0.75rem",
              fontSize: "0.9375rem",
              fontWeight: 600,
              color: "var(--ads-color-text)",
            }}
          >
            {caption}
          </caption>
          <thead>
            <tr>
              {headers.map((header) => (
                <th
                  key={header}
                  scope="col"
                  style={{
                    textAlign: "left",
                    padding: "0.625rem 0.75rem",
                    borderBottom: "2px solid var(--ads-color-border)",
                    fontWeight: 700,
                  }}
                >
                  {header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, index) => (
              <tr key={index}>
                {row.map((cell, cellIndex) => (
                  <td
                    key={cellIndex}
                    style={{
                      padding: "0.625rem 0.75rem",
                      borderBottom: "1px solid var(--ads-color-border)",
                      background: cellIndex === 0 ? "var(--ads-color-surface-muted)" : undefined,
                      fontWeight: cellIndex === 0 ? 600 : undefined,
                    }}
                  >
                    {cell}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {note ? (
        <figcaption className="fr-text--sm" style={{ marginTop: "0.5rem", color: "var(--ads-color-text-muted)" }}>
          {note}
        </figcaption>
      ) : null}
    </figure>
  );
}

/** Cross-links between the pages of a theme — the footer of every theme page. */
export function RelatedPages({
  title,
  kicker = "Dans le même thème",
  pages,
}: {
  title: string;
  kicker?: string;
  pages: ReadonlyArray<{ key: string; label: string; desc: string; href: string }>;
}) {
  return (
    <ThemeSection id="related-pages" kicker={kicker} title={title} subtle>
      <ul role="list" style={linkListStyle}>
        {pages.map((page) => (
          <li key={page.key}>
            <Link
              href={page.href}
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
              <span>
                <span style={{ display: "block" }}>{page.label}</span>
                <span
                  className="fr-text--sm"
                  style={{ display: "block", fontWeight: 400, color: "var(--ads-color-text-muted)" }}
                >
                  {page.desc}
                </span>
              </span>
              <span className="fr-icon-arrow-right-line" aria-hidden="true" />
            </Link>
          </li>
        ))}
      </ul>
    </ThemeSection>
  );
}

/* ---- Article renderer of the theme pages ---------------------------------- */

/** One localized content block of a theme page. */
export type LocalizedSection = {
  id: string;
  kicker: string;
  title: string;
  lead?: string;
  subtle?: boolean;
  /** Editorial paragraphs, rendered inside `.gov-prose`. */
  paragraphs?: ReadonlyArray<string>;
  /** Simple list, rendered inside `.gov-prose`. */
  bullets?: ReadonlyArray<string>;
  /** Institutional reference facts (label + value), rendered as a definition grid. */
  facts?: ReadonlyArray<{ key: string; label: string; value: string }>;
  /** Process / flow schema (chain of labelled nodes). */
  flow?: ReadonlyArray<{ key: string; label: string }>;
  /** Feature cards (icon + title + text). */
  cards?: ReadonlyArray<{ key: string; title: string; text: string; iconId?: string }>;
  /** ADS tiles (title + desc + icon + link). */
  tiles?: ReadonlyArray<{ key: string; title: string; desc: string; href: string; iconId: string }>;
  /** Figure cards (“en chiffres”). */
  statGrid?: ReadonlyArray<StatItem>;
  /** Responsive data table. */
  table?: {
    caption: string;
    headers: string[];
    rows: ReadonlyArray<ReadonlyArray<string | number>>;
    note?: string;
  };
  /** Denomination cards (banknotes / coins). */
  denominations?: ReadonlyArray<{ key: string; value: string; unit: string; tag?: string; motif?: string }>;
  /** Numbered steps. */
  steps?: ReadonlyArray<{ key: string; title: string; text: string }>;
  /** Link rows. */
  links?: ReadonlyArray<{ key: string; label: string; href: string }>;
  /** Callout shown at the end of the section. */
  notice?: string;
  /** Inline link shown at the end of the section. */
  cta?: { label: string; href: string };
};

export type LocalizedArticle = {
  hero: {
    kicker: string;
    title: string;
    lead: string;
    cta?: { label: string; href: string };
    notice?: string;
    stat?: { value: string; label: string };
  };
  sections: ReadonlyArray<LocalizedSection>;
  /** Heading of the cross-links block, resolved from `pages.<theme>.related.title`. */
  relatedTitle: string;
  /** Kicker of the cross-links block, resolved from `pages.<theme>.related.kicker`. */
  relatedKicker: string;
  related: ReadonlyArray<{ key: string; label: string; desc: string; href: string }>;
};

/**
 * Full page composed from a *localized* content object (produced by
 * `lib/monnaie-localize.ts` from the message catalogs): hero, ordered
 * sections and cross-links.
 */
export function ThemeArticle({
  content,
  currentHref,
}: {
  content: LocalizedArticle;
  currentHref: string;
}) {
  return (
    <>
      <ThemeHero
        kicker={content.hero.kicker}
        title={content.hero.title}
        lead={content.hero.lead}
        stat={content.hero.stat}
        actions={
          content.hero.cta
            ? [
                {
                  children: content.hero.cta.label,
                  href: content.hero.cta.href,
                  priority: "secondary",
                  iconId: "fr-icon-arrow-right-line",
                },
              ]
            : undefined
        }
      />

      {content.hero.notice ? (
        <div
          className="gov-section__container"
          style={{ maxWidth: "52rem", marginTop: "-1.5rem", marginInline: "auto" }}
        >
          <NoticeCallout iconId="fr-icon-information-line">{content.hero.notice}</NoticeCallout>
        </div>
      ) : null}

      {content.sections.map((section) => (
        <ThemeSection
          key={section.id}
          id={section.id}
          kicker={section.kicker}
          title={section.title}
          lead={section.lead}
          subtle={section.subtle}
        >
          {section.paragraphs ? (
            <div className="gov-prose">
              {section.paragraphs.map((paragraph) => (
                <p key={paragraph}>{paragraph}</p>
              ))}
            </div>
          ) : null}

          {section.bullets ? (
            <div className="gov-prose">
              <ul>
                {section.bullets.map((bullet) => (
                  <li key={bullet}>{bullet}</li>
                ))}
              </ul>
            </div>
          ) : null}

          {section.facts ? (
            <div style={{ marginTop: "1.5rem" }}>
              <FactList items={section.facts} />
            </div>
          ) : null}

          {section.flow ? (
            <div style={{ marginTop: "1.5rem" }}>
              <FlowDiagram items={section.flow} />
            </div>
          ) : null}

          {section.cards ? (
            <ul className="fr-grid-row fr-grid-row--gutters" role="list" style={cardGridStyle}>
              {section.cards.map((card) => (
                <li key={card.key} className="fr-col-12 fr-col-md-6 fr-col-lg-3">
                  <div style={teaserCardStyle}>
                    {card.iconId ? (
                      <span className={card.iconId} aria-hidden="true" style={iconBlockStyle} />
                    ) : null}
                    <span style={teaserTitleStyle}>{card.title}</span>
                    <span style={teaserDescStyle}>{card.text}</span>
                  </div>
                </li>
              ))}
            </ul>
          ) : null}

          {section.tiles ? (
            <div className="fr-grid-row fr-grid-row--gutters">
              {section.tiles.map((tile) => (
                <div key={tile.key} className="fr-col-12 fr-col-md-6 fr-col-lg-3">
                  <LinkTile
                    title={tile.title}
                    desc={tile.desc}
                    href={tile.href}
                    iconId={tile.iconId as FrIconClassName}
                  />
                </div>
              ))}
            </div>
          ) : null}

          {section.statGrid ? (
            <StatGrid items={section.statGrid} perRow={section.statGrid.length % 3 === 0 ? 3 : 4} />
          ) : null}

          {section.table ? (
            <DataTable
              caption={section.table.caption}
              headers={[...section.table.headers]}
              rows={section.table.rows}
              note={section.table.note}
            />
          ) : null}

          {section.denominations ? (
            <>
              <ul className="fr-grid-row fr-grid-row--gutters" role="list" style={cardGridStyle}>
                {section.denominations.map((denomination) => (
                  <li key={denomination.key} className="fr-col-12 fr-col-sm-6 fr-col-lg-4">
                    <div style={denominationCardStyle}>
                      <span style={denominationValueStyle}>{denomination.value}</span>
                      <span style={denominationUnitStyle}>{denomination.unit}</span>
                      {denomination.tag ? <span style={teaserTagStyle}>{denomination.tag}</span> : null}
                      {denomination.motif ? <span style={teaserTitleStyle}>{denomination.motif}</span> : null}
                    </div>
                  </li>
                ))}
              </ul>
            </>
          ) : null}

          {section.steps ? (
            <ol
              role="list"
              style={{
                listStyle: "none",
                margin: "0",
                padding: "0",
                display: "grid",
                gap: "1rem",
                maxWidth: "72rem",
              }}
            >
              {section.steps.map((step, index) => (
                <li
                  key={step.key}
                  style={{
                    display: "flex",
                    gap: "1rem",
                    alignItems: "flex-start",
                    padding: "1rem 1.25rem",
                    background: "var(--ads-color-background)",
                    border: "1px solid var(--ads-color-border)",
                  }}
                >
                  <span style={stepBadgeStyle} aria-hidden="true">
                    {index + 1}
                  </span>
                  <span>
                    <span style={{ display: "block", fontWeight: 700 }}>{step.title}</span>
                    <span
                      style={{
                        display: "block",
                        fontSize: "0.9375rem",
                        lineHeight: 1.55,
                        color: "var(--ads-color-text-muted)",
                      }}
                    >
                      {step.text}
                    </span>
                  </span>
                </li>
              ))}
            </ol>
          ) : null}

          {section.links ? (
            <ul role="list" style={linkListStyle}>
              {section.links.map((link) => (
                <li key={link.key}>
                  <Link
                    href={link.href}
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
                    {link.label}
                    <span className="fr-icon-arrow-right-line" aria-hidden="true" />
                  </Link>
                </li>
              ))}
            </ul>
          ) : null}

          {section.cta ? (
            <p>
              <Link href={section.cta.href} style={{ fontWeight: 600, textUnderlineOffset: "0.2em" }}>
                {section.cta.label}
                <span className="fr-icon-arrow-right-line" aria-hidden="true" />
              </Link>
            </p>
          ) : null}

          {section.notice ? (
            <div style={{ marginTop: "1.5rem" }}>
              <NoticeCallout iconId="fr-icon-information-line">{section.notice}</NoticeCallout>
            </div>
          ) : null}
        </ThemeSection>
      ))}

      <RelatedPages
        title={content.relatedTitle}
        kicker={content.relatedKicker}
        pages={content.related.filter((page) => page.href !== currentHref)}
      />
    </>
  );
}