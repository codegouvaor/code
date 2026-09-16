import { getTranslations } from "next-intl/server";
import type { Locale } from "@/i18n/routing";
import { NoticeCallout } from "./ads-fragments";

/**
 * Standard editorial page of the localized portal. Every word rendered here is
 * resolved from the message catalog of the active locale — no page-specific
 * logic is required to add or translate a page.
 */
export type SectionPageKind =
  | "government"
  | "organisation"
  | "news"
  | "services"
  | "contact"
  | "search"
  | "sitemap"
  | "decryptages"
  | "lEtatEtMoi"
  | "preventionDesRisques"
  | "suiviDesEngagements"
  | "liensUtiles"
  | "participation"
  | "accessibility"
  | "privacy"
  | "terms"
  | "cookies";

const legalPageKinds: ReadonlySet<SectionPageKind> = new Set([
  "accessibility",
  "privacy",
  "terms",
  "cookies",
]);

function messagesNamespace(kind: SectionPageKind): string {
  return legalPageKinds.has(kind) ? `pages.legal.${kind}` : `pages.${kind}`;
}

export async function SectionPage({
  locale,
  kind,
  paragraphCount = 1,
  showNotice = false,
  children,
}: {
  locale: Locale;
  kind: SectionPageKind;
  /** Number of translated paragraphs available under `paragraphs.pN`. */
  paragraphCount?: 1 | 2;
  showNotice?: boolean;
  children?: React.ReactNode;
}) {
  const t = await getTranslations({ locale, namespace: messagesNamespace(kind) });

  return (
    <div className="gov-section">
      <div className="gov-section__container gov-prose">
        <p className="gov-kicker">{t("kicker")}</p>
        <h1>{t("title")}</h1>
        <p className="gov-lead">{t("lead")}</p>

        {Array.from({ length: paragraphCount }, (_, index) => (
          <p key={index}>{t(`paragraphs.p${index + 1}`)}</p>
        ))}

        {children}

        {showNotice && (
          <div style={{ marginTop: "1.5rem" }}>
            <NoticeCallout iconId="fr-icon-information-line">{t("notice")}</NoticeCallout>
          </div>
        )}
      </div>
    </div>
  );
}
