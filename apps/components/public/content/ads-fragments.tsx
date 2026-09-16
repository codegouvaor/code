"use client";

import { CallOut } from "@codegouvaor/react-ads/CallOut";
import { Tile } from "@codegouvaor/react-ads/Tile";
import { Card } from "@codegouvaor/react-ads/Card";
import { Tag } from "@codegouvaor/react-ads/Tag";
import { ButtonsGroup } from "@codegouvaor/react-ads/ButtonsGroup";
import type { FrIconClassName } from "@codegouvaor/react-ads/fr";

/**
 * Thin serializable wrappers around ADS building blocks.
 *
 * Server-rendered page content cannot pass functions or elements to client
 * components, so the portal exposes the ADS primitives it needs through these
 * tiny client boundaries (strings and hrefs only). The visual and behavioural
 * source of truth stays in @codegouvaor/react-ads.
 */

export function NoticeCallout({
  iconId,
  title,
  children,
}: {
  iconId?: FrIconClassName;
  title?: string;
  children: string;
}) {
  return (
    <CallOut iconId={iconId} title={title}>
      {children}
    </CallOut>
  );
}

/**
 * Clickable ADS tag used for the hero's popular searches: an accessible
 * in-page link that stays consistent with the DSFR tag styling.
 */
export function SearchSuggestionTag({ label, href }: { label: string; href: string }) {
  return (
    <Tag as="a" small linkProps={{ href }}>
      {label}
    </Tag>
  );
}

export type LinkTileProps = {
  title: string;
  desc?: string;
  href: string;
  small?: boolean;
  grey?: boolean;
  horizontal?: boolean;
  iconId?: FrIconClassName;
};

export function LinkTile({ title, desc, href, small, grey, horizontal, iconId }: LinkTileProps) {
  return (
    <Tile
      small={small}
      grey={grey}
      orientation={horizontal ? "horizontal" : "vertical"}
      title={title}
      desc={desc}
      pictogram={
        <span aria-hidden="true" className={iconId ?? "fr-icon-arrow-right-line"} />
      }
      linkProps={{ href }}
    />
  );
}

/**
 * Editorial card of the home page (news article, public policy…).
 *
 * The whole card links to `href`; the optional `tag` renders as an ADS `Tag`
 * above the title and `date` as the card detail line, following the DSFR news
 * card pattern.
 *
 * No `iconId` is passed to the Card: the DSFR stylesheet draws the enlarge-link
 * arrow itself (bottom-right, same as the tiles) and adding an `fr-icon-*`
 * class on the card root would paint a second, stray arrow at the top.
 */
export type ArticleCardProps = {
  title: string;
  desc?: string;
  tag?: string;
  date?: string;
  href: string;
  size?: "small" | "medium" | "large";
  /** White card on an alternating background section. */
  background?: boolean;
  border?: boolean;
  titleAs?: `h${2 | 3 | 4 | 5 | 6}`;
};

export function ArticleCard({
  title,
  desc,
  tag,
  date,
  href,
  size,
  background,
  border,
  titleAs = "h3",
}: ArticleCardProps) {
  return (
    <Card
      enlargeLink
      linkProps={{ href }}
      title={title}
      titleAs={titleAs}
      desc={desc}
      size={size}
      background={background}
      border={border}
      start={
        tag ? (
          <Tag as="span" small>
            {tag}
          </Tag>
        ) : undefined
      }
      detail={date}
    />
  );
}

export type CtaButton = {
  children: string;
  href: string;
  title?: string;
  priority?: "primary" | "secondary" | "tertiary" | "tertiary no outline";
  iconId?: FrIconClassName;
  iconPosition?: "left" | "right";
};

export function CtaButtonsGroup({
  buttons,
  alignment = "left",
}: {
  buttons: [CtaButton, ...CtaButton[]];
  alignment?: "left" | "center" | "right" | "between";
}) {
  const adsButtons = buttons.map((button) => {
    const common = {
      priority: button.priority ?? "primary",
      children: button.children,
      title: button.title,
      linkProps: { href: button.href },
    };

    return button.iconId
      ? {
          ...common,
          iconId: button.iconId,
          iconPosition: button.iconPosition ?? "right",
        }
      : common;
  });

  return (
    <ButtonsGroup
      inlineLayoutWhen="md and up"
      alignment={alignment}
      buttons={adsButtons as Parameters<typeof ButtonsGroup>[0]["buttons"]}
    />
  );
}
