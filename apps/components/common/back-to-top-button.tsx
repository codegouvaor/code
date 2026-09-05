"use client";

import * as React from "react";
import { ArrowUp } from "lucide-react";
import { useTranslations } from "next-intl";

const EXTRA_OFFSET_PX = 16;

export function BackToTopButton() {
  const t = useTranslations("common");
  const [visible, setVisible] = React.useState(false);
  const [bottomOffset, setBottomOffset] = React.useState(EXTRA_OFFSET_PX);

  React.useEffect(() => {
    let frameId = 0;

    const updateState = () => {
      frameId = 0;

      // Appears as soon as the page is scrolled, so the visitor can always
      // jump back to the top without scrolling all the way up.
      setVisible(window.scrollY > 0);

      const consentBanner = document.querySelector<HTMLElement>('[data-consent-banner="true"]');
      const nextOffset =
        consentBanner && consentBanner.offsetParent !== null
          ? consentBanner.getBoundingClientRect().height + EXTRA_OFFSET_PX
          : EXTRA_OFFSET_PX;

      setBottomOffset((current) => (current === nextOffset ? current : nextOffset));
    };

    const scheduleUpdate = () => {
      if (frameId !== 0) {
        return;
      }

      frameId = window.requestAnimationFrame(updateState);
    };

    const observer = new MutationObserver(scheduleUpdate);
    observer.observe(document.body, { childList: true, subtree: true, attributes: true });

    window.addEventListener("scroll", scheduleUpdate, { passive: true });
    window.addEventListener("resize", scheduleUpdate, { passive: true });
    scheduleUpdate();

    return () => {
      if (frameId !== 0) {
        window.cancelAnimationFrame(frameId);
      }

      observer.disconnect();
      window.removeEventListener("scroll", scheduleUpdate);
      window.removeEventListener("resize", scheduleUpdate);
    };
  }, []);

  const handleClick = React.useCallback(() => {
    window.scrollTo({ top: 0, behavior: "smooth" });
  }, []);

  // Compact circular bubble, always shown in the bottom-right corner when the
  // page is scrolled. The portal pages do not load Tailwind utilities, so the
  // appearance lives entirely in `.gov-back-to-top` (globals.css): a fixed
  // overlay independent from the page flow and from the footer.
  return (
    <button
      type="button"
      aria-label={t("backToTop.ariaLabel")}
      aria-hidden={!visible}
      tabIndex={visible ? 0 : -1}
      onClick={handleClick}
      className={`gov-back-to-top${visible ? " is-visible" : ""}`}
      style={{ bottom: `${bottomOffset}px` }}
    >
      <ArrowUp strokeWidth={2.5} aria-hidden="true" />
    </button>
  );
}
