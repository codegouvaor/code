import type { Metadata } from "next";
import { AdsProvider } from "@/components/public/ads/ads-provider";
import RegisterForm from "@/components/auth/register-form";
import "@codegouvaor/react-ads/main.css";

export const metadata: Metadata = {
  title: "Create account — Official Portal",
  description:
    "Create your personal space on the official portal of the Republic of Astoria.",
};

export default function RegisterPage() {
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

          {/* Register card */}
          <div className="gov-login__card">
            <div className="gov-login__card-header">
              <h1 className="gov-login__title">Create an account</h1>
              <p className="gov-login__subtitle">
                Set up your personal space to access government services and
                manage your documents.
              </p>
            </div>

            <div className="gov-login__card-body">
              <RegisterForm />
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
