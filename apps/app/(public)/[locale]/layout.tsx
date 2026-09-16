import type { Metadata } from "next";
import { NextIntlClientProvider } from "next-intl";
import { getMessages, getTranslations, setRequestLocale } from "next-intl/server";
import { routing, type Locale } from "@/i18n/routing";
import { pageAnchors } from "@/lib/site-structure";
import { AdsProvider } from "@/components/public/ads/ads-provider";
import { GovernmentHeader } from "@/components/public/header/government-header";
import { GovernmentFooter } from "@/components/public/footer/government-footer";
import { BackToTopButton } from "@/components/common/back-to-top-button";
import { PopUp } from "@/components/common/pop-up";

import "@codegouvaor/react-ads/main.css";
  
const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL ?? "https://code.gouv.aor";

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale: localeParam } = await params;
  const locale = routing.locales.includes(localeParam as Locale)
    ? (localeParam as Locale)
    : routing.defaultLocale;

  const tMeta = await getTranslations({ locale, namespace: "meta" });

  return {
    metadataBase: new URL(SITE_URL),
    title: {
      default: tMeta("defaultTitle"),
      template: `%s | ${tMeta("suffix")}`,
    },
    description: tMeta("description"),
    icons: {
      icon: [
        {
          url: "/astoria-gouv.png",
          type: "image/svg+xml",
        },
        { url: "/astoria-gouv.png" },
      ],
      apple: "/astoria-gouv.png",
    },
  };
}

export default async function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale: localeParam } = await params;

  if (!routing.locales.includes(localeParam as Locale)) {
    return null;
  }

  const locale = localeParam as Locale;
  setRequestLocale(locale);

  const messages = await getMessages();

  return (
    <html lang={locale} suppressHydrationWarning>
      <body className="select-none">
        <NextIntlClientProvider locale={locale} messages={messages}>
          <AdsProvider lang={locale}>
            <div className="gov-page">
              <GovernmentHeader />
              <main id={pageAnchors.content} className="gov-main">
                {children}
              </main>
              <GovernmentFooter />
              <BackToTopButton />
              <PopUp />
            </div>
          </AdsProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}