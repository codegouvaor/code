import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { localizedAlternates, resolveLocaleParam } from "@/lib/localized-metadata";
import { localizedContactRecipients } from "@/lib/content/contact/contact-recipients";
import { FactList, ThemeHero, ThemeSection } from "@/components/public/content/theme-page";
import ContactForm from "@/components/public/contact/contact-form";

const PAGE_PATH = "/contact";
const PAGE_NAMESPACE = "pages.contact";

type PageProps = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);
  const t = await getTranslations({ locale, namespace: PAGE_NAMESPACE });

  return {
    title: t("meta.title"),
    description: t("meta.description"),
    ...localizedAlternates(locale, PAGE_PATH),
  };
}

/**
 * Contact page of the Banque centrale d'Astoria.
 *
 * Server component: it resolves the localized recipients of the contact form
 * and composes the institutional hero, the form section and the contact
 * channels. Every display string lives in the message catalogs
 * (`pages.contact.*`); no data is stored server-side — the form delegates to
 * the user's email client through a `mailto:` link.
 */
export default async function ContactPage({ params }: PageProps) {
  const { locale: rawLocale } = await params;
  const locale = resolveLocaleParam(rawLocale);
  setRequestLocale(locale);

  const t = await getTranslations({ locale, namespace: PAGE_NAMESPACE });
  const recipients = localizedContactRecipients((id) => t(`recipients.${id}`));

  return (
    <>
      <ThemeHero kicker={t("hero.kicker")} title={t("hero.title")} lead={t("hero.lead")} />

      <ThemeSection
        id="contact-form"
        kicker={t("formSection.kicker")}
        title={t("formSection.title")}
        lead={t("formSection.lead")}
      >
        <ContactForm recipients={recipients} />
      </ThemeSection>

      <ThemeSection
        id="contact-channels"
        kicker={t("channels.kicker")}
        title={t("channels.title")}
        lead={t("channels.lead")}
        subtle
      >
        <FactList
          items={[
            { key: "email", label: t("channels.facts.email.label"), value: t("channels.facts.email.value") },
            { key: "phone", label: t("channels.facts.phone.label"), value: t("channels.facts.phone.value") },
            { key: "address", label: t("channels.facts.address.label"), value: t("channels.facts.address.value") },
            { key: "hours", label: t("channels.facts.hours.label"), value: t("channels.facts.hours.value") },
          ]}
        />
      </ThemeSection>
    </>
  );
}