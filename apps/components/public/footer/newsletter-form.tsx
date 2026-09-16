"use client";

import * as React from "react";
import { useTranslations } from "next-intl";

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/**
 * Newsletter subscription form of the footer “stay in touch” zone.
 *
 * The form validates the address client-side and confirms the subscription.
 * There is no public subscription API yet (the server endpoint is restricted
 * to the admin area), so the submit is simulated: this component is the seam
 * where a future public endpoint plugs in — swap the simulated success for a
 * call to the API (e.g. `newsletterApi.subscribe(email)`) and surface its
 * response, without changing the markup or the styles.
 *
 * Styles come from the ADS tokens (`var(--ads-*)`) of
 * `@codegouvaor/react-ads/main.css` and the `ads-sr-only` helper — no local
 * stylesheet.
 */
export function NewsletterForm() {
  const t = useTranslations("stayInTouch.newsletter");

  const [email, setEmail] = React.useState("");
  const [status, setStatus] = React.useState<"idle" | "error" | "success">("idle");
  const [error, setError] = React.useState<string | null>(null);

  const inputId = React.useId();
  const errorId = React.useId();

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    const value = email.trim();
    if (!EMAIL_PATTERN.test(value)) {
      setStatus("error");
      setError(t("invalidEmail"));
      return;
    }

    setStatus("success");
    setError(null);
    setEmail("");
  }

  const feedbackStyle: React.CSSProperties = {
    margin: "0.625rem 0 0",
    fontSize: "0.8125rem",
    lineHeight: 1.4,
  };

  return (
    <form
      onSubmit={handleSubmit}
      noValidate
      style={{ maxWidth: "27rem", marginTop: "0.5rem" }}
    >
      <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
        <label className="ads-sr-only" htmlFor={inputId}>
          {t("label")}
        </label>
        <input
          id={inputId}
          type="email"
          autoComplete="email"
          placeholder={t("placeholder")}
          value={email}
          onChange={(e) => {
            setEmail(e.target.value);
            if (status === "error") {
              setStatus("idle");
              setError(null);
            }
          }}
          aria-invalid={status === "error"}
          aria-describedby={status === "error" ? errorId : undefined}
          style={{
            flex: "1 1 10rem",
            minWidth: 0,
            minHeight: "2.5rem",
            padding: "0 0.75rem",
            fontSize: "0.9375rem",
            lineHeight: 1.5,
            color: "var(--ads-color-text)",
            background: "var(--ads-color-background)",
            border: `1px solid ${
              status === "error"
                ? "var(--ads-color-danger)"
                : "var(--ads-color-border)"
            }`,
            borderRadius: "0.25rem",
          }}
        />
        <button
          type="submit"
          style={{
            display: "inline-flex",
            alignItems: "center",
            justifyContent: "center",
            minHeight: "2.5rem",
            padding: "0 1.125rem",
            fontSize: "0.9375rem",
            fontWeight: 600,
            lineHeight: 1.5,
            color: "var(--ads-color-on-primary)",
            background: "var(--ads-color-primary)",
            border: "none",
            borderRadius: "0.25rem",
            cursor: "pointer",
          }}
        >
          {t("submit")}
        </button>
      </div>

      {status === "error" && error && (
        <p id={errorId} role="alert" style={{ ...feedbackStyle, fontWeight: 600, color: "var(--ads-color-danger)" }}>
          {error}
        </p>
      )}
      {status === "success" && (
        <p role="status" style={{ ...feedbackStyle, color: "var(--ads-color-success)" }}>
          {t("success")}
        </p>
      )}

      <p style={{ ...feedbackStyle, color: "var(--ads-color-text-muted)" }}>{t("hint")}</p>
    </form>
  );
}
