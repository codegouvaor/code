"use client";

import * as React from "react";
import type { ComponentType } from "react";
import { useTranslations } from "next-intl";
import { NewsletterForm } from "./newsletter-form";
import { FacebookIcon } from "@/components/ui/icons/FacebookIcon";
import { InstagramIcon } from "@/components/ui/icons/InstagramIcon";
import { LinkedinIcon } from "@/components/ui/icons/LinkedinIcon";
import { ThreadsIcon } from "@/components/ui/icons/ThreadsIcon";
import { TwitterIcon } from "@/components/ui/icons/TwitterIcon";

/** Icon of each social account, keyed by the `follow` list below. */
const SOCIAL_ICONS: Record<string, ComponentType<{ className?: string }>> = {
  x: TwitterIcon,
  facebook: FacebookIcon,
  linkedin: LinkedinIcon,
  instagram: InstagramIcon,
  threads: ThreadsIcon,
};

/**
 * Social accounts of the ministry. The hrefs are placeholders (the ministry's
 * real accounts do not exist yet); labels resolve under
 * `stayInTouch.follow.items.<key>`.
 */
const FOLLOW_LINKS = [
  { key: "x", href: "https://x.com" },
  { key: "facebook", href: "https://www.facebook.com" },
  { key: "linkedin", href: "https://www.linkedin.com" },
  { key: "instagram", href: "https://www.instagram.com" },
  { key: "threads", href: "https://www.threads.net" },
] as const;

/**
 * “Stay in touch” band of the Government Footer: newsletter subscription and
 * the ministry's social accounts.
 *
 * Newsletter and social accounts are deliberately absent from the homepage —
 * they live in the footer, a secondary zone of the portal, so they never
 * compete with the public services of the page.
 *
 * The layout reuses the ADS primitives of `@codegouvaor/react-ads/main.css`
 * (`.gov-container`, DSFR grid) and the `--ads-*` tokens; no local stylesheet.
 */
export function StayInTouch() {
  const t = useTranslations("stayInTouch");

  return (
    <div
      style={{
        background: "var(--ads-color-surface-muted)",
        paddingBlock: "2rem",
      }}
    >
      <div className="gov-container">
        <div className="fr-grid-row fr-grid-row--gutters" style={{ alignItems: "start" }}>
          <div className="fr-col-12 fr-col-lg-8">
            <h2 style={{ margin: "0 0 0.5rem", fontSize: "1.125rem", lineHeight: 1.4 }}>
              {t("newsletter.title")}
            </h2>
            <p
              style={{
                margin: "0 0 1rem",
                maxWidth: "36rem",
                fontSize: "0.875rem",
                lineHeight: 1.6,
                color: "var(--ads-color-text-muted)",
              }}
            >
              {t("newsletter.desc")}
            </p>
            <NewsletterForm />
          </div>
          <div
            className="fr-col-12 fr-col-lg-4"
            style={{ paddingLeft: 0 }}
          >
            <h2
              style={{
                margin: "0 0 1rem",
                fontSize: "0.9375rem",
                lineHeight: 1.4,
              }}
            >
              {t("follow.title")}
            </h2>
            <ul
              role="list"
              style={{
                listStyle: "none",
                margin: "0",
                padding: "0",
                display: "flex",
                flexWrap: "wrap",
                gap: "0.625rem",
              }}
            >
              {FOLLOW_LINKS.map((item) => {
                const Icon = SOCIAL_ICONS[item.key];
                return (
                  <li key={item.key}>
                    <a
                      href={item.href}
                      aria-label={t(`follow.items.${item.key}`)}
                      style={{
                        display: "inline-flex",
                        alignItems: "center",
                        justifyContent: "center",
                        width: "2.25rem",
                        height: "2.25rem",
                        color: "var(--ads-color-text)",
                        background: "var(--ads-color-background)",
                        border: "1px solid var(--ads-color-border)",
                        borderRadius: "50%",
                        textDecoration: "none",
                      }}
                    >
                      <Icon />
                    </a>
                  </li>
                );
              })}
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
