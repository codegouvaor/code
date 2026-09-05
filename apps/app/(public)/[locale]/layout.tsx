import type { Metadata } from "next";
import { NextIntlClientProvider } from "next-intl";
import { getMessages, getTranslations, setRequestLocale } from "next-intl/server";
import { routing, type Locale } from "@/i18n/routing";
import { pageAnchors } from "@/lib/site-structure";
import { AdsProvider } from "@/components/public/ads/ads-provider";
import { GovernmentHeader } from "@/components/public/header/government-header";
import { GovernmentFooter } from "@/components/public/footer/government-footer";
import { BackToTopButton } from "@/components/common/back-to-top-button";

// Astoria Design System stylesheet (icons + components). Imported here so the
// CSS is only shipped to the localized public routes.
import "@codegouvaor/react-ads/main.css";
// Portal layer (Astorian identity lockup, page chrome, content typography).
import "@/styles/globals.css";

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
        { url: "/icon-light-32x32.png" },
      ],
      apple: "/apple-icon.png",
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
    <html lang={locale} suppressHydrationWarning className="select-none">
      <body className="gov-ads">
        <NextIntlClientProvider locale={locale} messages={messages}>
          <AdsProvider lang={locale}>
            <div className="gov-page">
              <GovernmentHeader />
              <main id={pageAnchors.content} className="gov-main">
                {children}
              </main>
              <GovernmentFooter />
              <BackToTopButton />
            </div>
          </AdsProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
