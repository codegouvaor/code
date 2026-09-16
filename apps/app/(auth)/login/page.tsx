import type { Metadata } from "next";
import type { CSSProperties } from "react";
import { AdsProvider } from "@/components/public/ads/ads-provider";
import { AuthPanel } from "@/components/auth/auth-panel";
import "@codegouvaor/react-ads/main.css";

export const metadata: Metadata = {
  title: "Sign in or create an account — Official Portal",
  description:
    "Sign in to your personal space or create an account on the official portal of the Republic of Astoria.",
};

const pageStyle: CSSProperties = {
  minHeight: "100dvh",
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  padding: "1rem",
  background: "var(--ads-color-surface-muted)",
};

const containerStyle: CSSProperties = {
  width: "100%",
  maxWidth: "28rem",
};

const brandStyle: CSSProperties = {
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.625rem",
  marginBottom: "1.25rem",
  textDecoration: "none",
  color: "var(--ads-color-text)",
};

const brandTitleStyle: CSSProperties = {
  fontWeight: 700,
  fontSize: "1rem",
  lineHeight: 1.3,
};

const brandSubtitleStyle: CSSProperties = {
  fontSize: "0.8125rem",
  color: "var(--ads-color-text-muted)",
};

const cardStyle: CSSProperties = {
  background: "var(--ads-color-background)",
  border: "1px solid var(--ads-color-border)",
  borderRadius: "0.5rem",
  padding: "1.25rem",
  boxShadow: "0 0.375rem 1rem rgb(0 0 0 / 8%)",
};

const footerStyle: CSSProperties = {
  marginTop: "1rem",
};

const footerTextStyle: CSSProperties = {
  margin: "0 0 0.5rem",
  fontSize: "0.75rem",
  lineHeight: 1.4,
  textAlign: "center",
  color: "var(--ads-color-text-muted)",
};

const footerLinksStyle: CSSProperties = {
  listStyle: "none",
  margin: 0,
  padding: 0,
  display: "flex",
  flexWrap: "wrap",
  justifyContent: "center",
  gap: "0.125rem 1rem",
};

const footerLinkStyle: CSSProperties = {
  fontSize: "0.75rem",
  color: "var(--ads-color-link)",
  textUnderlineOffset: "0.2em",
};

const FOOTER_LINKS = [
  { href: "/legal/accessibility", label: "Accessibility" },
  { href: "/legal/mentions-legales", label: "Legal notices" },
  { href: "/legal/donnees-personnelles", label: "Personal data" },
  { href: "/legal/cookies", label: "Cookies" },
];

type PageProps = { searchParams: Promise<{ tab?: string }> };

export default async function LoginPage({ searchParams }: PageProps) {
  await searchParams;

  return (
    <AdsProvider lang="en">
      <div style={pageStyle}>
        <div style={containerStyle}>
          {/* Header — Republic branding */}
          <a href="/" style={brandStyle} title="Back to homepage">
            <img
              src="/astoria-gouv.png"
              alt="Republic of Astoria"
              width={56}
              height={56}
            />
            <div>
              <div style={brandTitleStyle}>Official Portal</div>
              <div style={brandSubtitleStyle}>Republic of Astoria</div>
            </div>
          </a>

          {/* Login / Register card */}
          <div style={cardStyle}>
            <AuthPanel />
          </div>

          {/* Footer links */}
          <div style={footerStyle}>
            <p style={footerTextStyle}>
              This site is protected by an authentication system. Any
              unauthorised access attempt may be subject to criminal
              prosecution.
            </p>
            <ul style={footerLinksStyle}>
              {FOOTER_LINKS.map((link) => (
                <li key={link.href}>
                  <a href={link.href} style={footerLinkStyle}>
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </AdsProvider>
  );
}