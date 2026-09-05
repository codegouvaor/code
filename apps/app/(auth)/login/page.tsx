import type { Metadata } from "next";
import { AdsProvider } from "@/components/public/ads/ads-provider";
import LoginForm from "@/components/auth/login-form";
import "@codegouvaor/react-ads/main.css";

export const metadata: Metadata = {
  title: "Login — Official Portal",
  description:
    "Sign in to your personal space on the official portal of the Republic of Astoria.",
};

export default function LoginPage() {
  return (
    <AdsProvider lang="en">
      <div className="gov-login">
        <div className="gov-login__container">
          {/* Header — Republic branding */}
          <div className="gov-login__header">
            <a href="/" className="gov-login__logo-link" title="Back to homepage">
              <img
                src="/astoria-gouv.png"
                alt="Republic of Astoria"
                className="gov-login__logo"
                width={56}
                height={56}
              />
              <div className="gov-login__brand">
                <span className="gov-login__brand-title">Official Portal</span>
                <span className="gov-login__brand-subtitle">
                  Republic of Astoria
                </span>
              </div>
            </a>
          </div>

          {/* Login card */}
          <div className="gov-login__card">
            <div className="gov-login__card-header">
              <h1 className="gov-login__title">Sign in</h1>
              <p className="gov-login__subtitle">
                Access your personal space to manage your services and
                documents.
              </p>
            </div>

            <div className="gov-login__card-body">
              <LoginForm />
            </div>
          </div>

          {/* Footer links */}
          <div className="gov-login__footer">
            <p className="gov-login__footer-text">
              This site is protected by an authentication system. Any
              unauthorised access attempt may be subject to criminal
              prosecution.
            </p>
            <ul className="gov-login__footer-links">
              <li>
                <a href="/legal/accessibility">Accessibility</a>
              </li>
              <li>
                <a href="/legal/mentions-legales">Legal notices</a>
              </li>
              <li>
                <a href="/legal/donnees-personnelles">Personal data</a>
              </li>
              <li>
                <a href="/legal/cookies">Cookies</a>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </AdsProvider>
  );
}
