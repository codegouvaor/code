import type {
  LocalizedArticle,
  LocalizedSection,
  StatItem,
} from "@/components/public/content/theme-page";

/**
 * Localization layer shared by every theme of the BCA portal.
 *
 * A theme (Monnaie, Politique monétaire, and later Banques, Marchés
 * financiers, Paiements, Système financier) is a group of pages. Every display
 * string of those pages lives in the message catalogs
 * (`apps/messages/{fr,en}.json`) under `pages.<theme>.<page>` (per page) and
 * `pages.<theme>.related` (shared cross-links). The content files hold the
 * *structure* only: section ids, message keys, hrefs, icons and pure data
 * (values, years, figures).
 *
 * `localizeArticle` resolves a key-based content object into the localized
 * object rendered by `ThemeArticle`. It never holds text itself — it only
 * assembles the strings provided by the active locale's catalog. This is the
 * single convention every theme follows, so adding a theme never means
 * duplicating a localization layer.
 */

export type TranslateFn = (key: string) => string;

export type ArticleCellConfig = string | { key: string };

export type ArticleSectionConfig = {
  /** Message path segment under `pages.<theme>.<page>.sections`. */
  key: string;
  /** HTML id of the section. */
  id: string;
  subtle?: boolean;
  lead?: boolean;
  paragraphCount?: number;
  bulletCount?: number;
  cards?: ReadonlyArray<{ key: string; iconId?: string }>;
  tiles?: ReadonlyArray<{ key: string; href: string; iconId: string }>;
  statGrid?: ReadonlyArray<{ key: string; href?: string }>;
  table?: { headerCount: number; rowCells: ReadonlyArray<number>; note?: boolean };
  denominations?: ReadonlyArray<{ key: string; value: string; tag?: boolean; motif?: boolean }>;
  steps?: ReadonlyArray<{ key: string }>;
  /** Process / flow schema: each node resolves `flow.<key>`. */
  flow?: ReadonlyArray<{ key: string }>;
  /** Reference facts: each item resolves `facts.<key>.label` / `.value`. */
  facts?: ReadonlyArray<{ key: string }>;
  links?: ReadonlyArray<{ key: string; href: string }>;
  notice?: boolean;
  cta?: { href: string };
};

export type ArticleHeroConfig = {
  kickerKey: string;
  titleKey: string;
  leadKey: string;
  ctaKey?: string;
  ctaHref?: string;
  noticeKey?: string;
  /** Optional key figure shown in the hero (`hero.stat.value` / `.label`). */
  statKey?: string;
};

export type ArticleContent = {
  hero: ArticleHeroConfig;
  sections: ReadonlyArray<ArticleSectionConfig>;
  /** Shared cross-links: `key` resolves `pages.<theme>.related.<key>.*`, href is structural. */
  related: ReadonlyArray<{ key: string; href: string }>;
};

function range(count: number): number[] {
  return Array.from({ length: count }, (_, index) => index);
}

function sectionStrings(section: ArticleSectionConfig, t: TranslateFn): LocalizedSection {
  const base = `sections.${section.key}`;
  const localized: LocalizedSection = {
    id: section.id,
    kicker: t(`${base}.kicker`),
    title: t(`${base}.title`),
    lead: section.lead ? t(`${base}.lead`) : undefined,
    subtle: section.subtle,
  };

  if (section.paragraphCount) {
    localized.paragraphs = range(section.paragraphCount).map(
      (index) => t(`${base}.paragraphs.p${index + 1}`)
    );
  }

  if (section.bulletCount) {
    localized.bullets = range(section.bulletCount).map(
      (index) => t(`${base}.bullets.b${index + 1}`)
    );
  }

  if (section.cards) {
    localized.cards = section.cards.map((card) => ({
      key: card.key,
      title: t(`${base}.cards.${card.key}.title`),
      text: t(`${base}.cards.${card.key}.text`),
      iconId: card.iconId,
    }));
  }

  if (section.tiles) {
    localized.tiles = section.tiles.map((tile) => ({
      key: tile.key,
      title: t(`${base}.tiles.${tile.key}.title`),
      desc: t(`${base}.tiles.${tile.key}.desc`),
      href: tile.href,
      iconId: tile.iconId,
    }));
  }

  if (section.statGrid) {
    localized.statGrid = section.statGrid.map((item): StatItem => ({
      key: item.key,
      value: t(`${base}.statGrid.${item.key}.value`),
      label: t(`${base}.statGrid.${item.key}.label`),
      href: item.href,
    }));
  }

  if (section.table) {
    localized.table = {
      caption: t(`${base}.table.caption`),
      headers: range(section.table.headerCount).map((index) => t(`${base}.table.headers.${index}`)),
      rows: section.table.rowCells.map((cellCount, rowIndex) =>
        range(cellCount).map((cellIndex) =>
          t(`${base}.table.rows.${rowIndex}.${cellIndex}`)
        )
      ),
      note: section.table.note ? t(`${base}.table.note`) : undefined,
    };
  }

  if (section.denominations) {
    localized.denominations = section.denominations.map((denomination) => ({
      key: denomination.key,
      value: denomination.value,
      unit: t(`${base}.denominations.${denomination.key}.unit`),
      tag: denomination.tag ? t(`${base}.denominations.${denomination.key}.tag`) : undefined,
      motif: denomination.motif ? t(`${base}.denominations.${denomination.key}.motif`) : undefined,
    }));
  }

  if (section.steps) {
    localized.steps = section.steps.map((step) => ({
      key: step.key,
      title: t(`${base}.steps.${step.key}.title`),
      text: t(`${base}.steps.${step.key}.text`),
    }));
  }

  if (section.flow) {
    localized.flow = section.flow.map((node) => ({
      key: node.key,
      label: t(`${base}.flow.${node.key}`),
    }));
  }

  if (section.facts) {
    localized.facts = section.facts.map((fact) => ({
      key: fact.key,
      label: t(`${base}.facts.${fact.key}.label`),
      value: t(`${base}.facts.${fact.key}.value`),
    }));
  }

  if (section.links) {
    localized.links = section.links.map((link) => ({
      key: link.key,
      label: t(`${base}.links.${link.key}`),
      href: link.href,
    }));
  }

  if (section.notice) {
    localized.notice = t(`${base}.notice`);
  }

  if (section.cta) {
    localized.cta = { label: t(`${base}.cta`), href: section.cta.href };
  }

  return localized;
}

/**
 * Resolves a key-based content object into the localized `ThemeArticle` input
 * using the message catalogs of the active locale.
 *
 * @param content   Structural content of the page (`{page}-content.ts`).
 * @param t         Translator of the page (`pages.<theme>.<page>`).
 * @param tRelated  Translator of the shared cross-links (`pages.<theme>.related`).
 */
export function localizeArticle(
  content: ArticleContent,
  t: TranslateFn,
  tRelated: TranslateFn
): LocalizedArticle {
  return {
    hero: {
      kicker: t(content.hero.kickerKey),
      title: t(content.hero.titleKey),
      lead: t(content.hero.leadKey),
      cta:
        content.hero.ctaKey && content.hero.ctaHref
          ? { label: t(content.hero.ctaKey), href: content.hero.ctaHref }
          : undefined,
      notice: content.hero.noticeKey ? t(content.hero.noticeKey) : undefined,
      stat: content.hero.statKey
        ? {
            value: t(`${content.hero.statKey}.value`),
            label: t(`${content.hero.statKey}.label`),
          }
        : undefined,
    },
    sections: content.sections.map((section) => sectionStrings(section, t)),
    relatedTitle: tRelated("title"),
    relatedKicker: tRelated("kicker"),
    related: content.related.map(({ key, href }) => ({
      key,
      label: tRelated(`${key}.label`),
      desc: tRelated(`${key}.desc`),
      href,
    })),
  };
}
