"use client";

import * as React from "react";
import { BackToTop } from "@codegouvaor/react-ads/BackToTop";
import { useTranslations } from "next-intl";

const EXTRA_OFFSET_PX = 16;

/**
 * Floating “back to top” bubble of the portal, wrapping the ADS `BackToTop`
 * component (the behavioural and visual source of truth).
 *
 * The only ministry-specific behaviour kept here is the vertical offset used
 * to clear the consent banner when one is displayed on the page: the banner
 * height is observed and forwarded to ADS through the `bottom` style.
 */
export function BackToTopButton() {
  const t = useTranslations("common");
  const [bottomOffset, setBottomOffset] = React.useState(EXTRA_OFFSET_PX);

  React.useEffect(() => {
    let frameId = 0;

    const updateOffset = () => {
      frameId = 0;

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

      frameId = window.requestAnimationFrame(updateOffset);
    };

    const observer = new MutationObserver(scheduleUpdate);
    observer.observe(document.body, { childList: true, subtree: true, attributes: true });

    window.addEventListener("resize", scheduleUpdate, { passive: true });
    scheduleUpdate();

    return () => {
      if (frameId !== 0) {
        window.cancelAnimationFrame(frameId);
      }

      observer.disconnect();
      window.removeEventListener("resize", scheduleUpdate);
    };
  }, []);

  return <BackToTop threshold={1} label={t("backToTop.ariaLabel")} style={{ bottom: `${bottomOffset}px` }} />;
}
