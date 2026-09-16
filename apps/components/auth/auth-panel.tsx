"use client";

import * as React from "react";
import LoginForm from "./login-form";
import RegisterForm from "./register-form";

export type AuthTab = "login" | "register";

const toggleStyle: React.CSSProperties = {
  display: "flex",
  marginBottom: "1.25rem",
  border: "1px solid var(--ads-color-border)",
  borderRadius: "0.5rem",
  overflow: "hidden",
};

const tabButtonStyle = (active: boolean): React.CSSProperties => ({
  flex: 1,
  padding: "0.625rem 1rem",
  background: active ? "var(--ads-color-primary)" : "transparent",
  color: active ? "var(--ads-color-on-primary)" : "var(--ads-color-text)",
  fontWeight: 600,
  fontSize: "0.9375rem",
  border: "none",
  cursor: "pointer",
});

/**
 * Single auth component: a Login / Register toggle. Clicking one of the two
 * tabs shows the corresponding form, so both spaces live in one component.
 */
export function AuthPanel() {
  const [tab, setTab] = React.useState<AuthTab>("login");

  return (
    <div>
      <div role="tablist" style={toggleStyle} aria-label="Authentication">
        <button
          type="button"
          role="tab"
          aria-selected={tab === "login"}
          onClick={() => setTab("login")}
          style={tabButtonStyle(tab === "login")}
        >
          Sign in
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={tab === "register"}
          onClick={() => setTab("register")}
          style={tabButtonStyle(tab === "register")}
        >
          Create an account
        </button>
      </div>

      {tab === "login" ? <LoginForm /> : <RegisterForm />}
    </div>
  );
}