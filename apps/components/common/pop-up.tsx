"use client";

import * as React from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";

const STORAGE_KEY = "bce-cookie-consent";

/**
 * Floating cookie consent indicator, anchored at the bottom-left of the page.
 *
 * It only appears on the client, once, until the visitor accepts or declines.
 * The choice is persisted in `localStorage`. The `data-consent-banner="true"`
 * attribute lets the floating “back to top” bubble clear it vertically.
 */
export function PopUp() {
  const t = useTranslations("common.cookiePopup");
  const [visible, setVisible] = React.useState(false);

  React.useEffect(() => {
    try {
      if (window.localStorage.getItem(STORAGE_KEY) === null) {
        setVisible(true);
      }
    } catch {
      setVisible(true);
    }
  }, []);

  function recordConsent(value: "accepted" | "declined") {
    try {
      window.localStorage.setItem(STORAGE_KEY, value);
    } catch {
      // localStorage unavailable: still hide the banner for this visit.
    }
    setVisible(false);
  }

  if (!visible) {
    return null;
  }

  return (
    <aside
      data-consent-banner="true"
      role="dialog"
      aria-label={t("ariaLabel")}
      style={{
        position: "fixed",
        left: "1rem",
        bottom: "1rem",
        zIndex: 999,
        maxWidth: "26rem",
        padding: "1.25rem 1.5rem",
        background: "var(--ads-color-surface)",
        color: "var(--ads-color-text)",
        border: "1px solid var(--ads-color-border)",
        borderRadius: "var(--ads-radius-large)",
        boxShadow: "var(--ads-elevation-large)",
      }}
    >
      <p style={{ margin: 0, fontSize: "1rem", fontWeight: 600 }}>{t("title")}</p>
      <p
        style={{
          margin: "0.5rem 0 1rem",
          color: "var(--ads-color-text-muted)",
          fontSize: "0.875rem",
          lineHeight: 1.5,
        }}
      >
        {t("message")}
      </p>
      <div style={{ display: "flex", flexWrap: "wrap", gap: "0.5rem" }}>
        <button type="button" className="fr-btn" onClick={() => recordConsent("accepted")}>
          {t("accept")}
        </button>
        <button
          type="button"
          className="fr-btn fr-btn--secondary"
          onClick={() => recordConsent("declined")}
        >
          {t("decline")}
        </button>
      </div>
      <p style={{ margin: "0.75rem 0 0" }}>
        <Link href="/legal/cookies" style={{ fontSize: "0.875rem" }}>
          {t("manage")}
        </Link>
      </p>
    </aside>
  );
}