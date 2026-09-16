"use client";

import * as React from "react";
import { useAuth } from "@/context/AuthContext";
import { useTranslations } from "next-intl";
import { siteAccountConfig } from "@/lib/site-config";

/**
 * Account menu displayed in the header when the user is authenticated.
 *
 * Renders a button with the user's display name and a dropdown containing
 * site-specific menu items (profile, settings, etc.) plus a logout action.
 *
 * The menu content is driven by `siteAccountConfig` so each site can
 * present a different account interface without touching this component.
 *
 * The appearance uses the ADS design tokens (`var(--ads-*)`, provided by
 * `@codegouvaor/react-ads/main.css`) through inline styles — no local
 * stylesheet.
 */
export function UserAccountMenu() {
  const { user, logout, isLoading } = useAuth();
  const t = useTranslations();
  const [isOpen, setIsOpen] = React.useState(false);
  const menuRef = React.useRef<HTMLDivElement>(null);
  const buttonRef = React.useRef<HTMLButtonElement>(null);

  // Close on outside click
  React.useEffect(() => {
    if (!isOpen) return;

    function handleClickOutside(event: MouseEvent) {
      if (
        menuRef.current &&
        !menuRef.current.contains(event.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    }

    function handleEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setIsOpen(false);
        buttonRef.current?.focus();
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleEscape);
    };
  }, [isOpen]);

  // Don't render anything while loading or if not authenticated
  if (isLoading || !user) return null;

  if (!siteAccountConfig.enabled) return null;

  const displayName = user.displayName || user.name || user.email;

  return (
    <div style={{ position: "relative", zIndex: 850 }}>
      <button
        ref={buttonRef}
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="true"
        aria-controls="account-menu-dropdown"
        style={{
          display: "inline-flex",
          alignItems: "center",
          gap: "0.5rem",
          padding: "0.5rem 0.75rem",
          fontSize: "0.875rem",
          fontWeight: 600,
          lineHeight: 1.5,
          color: "var(--ads-color-text)",
          background: "transparent",
          border: "1px solid var(--ads-color-border)",
          borderRadius: "0.25rem",
          cursor: "pointer",
        }}
      >
        {user.avatarUrl ? (
          <img
            src={user.avatarUrl}
            alt=""
            width={28}
            height={28}
            style={{ width: "1.75rem", height: "1.75rem", borderRadius: "50%", objectFit: "cover" }}
          />
        ) : (
          <span
            className="fr-icon-account-circle-line"
            aria-hidden="true"
            style={{ fontSize: "1.25rem", lineHeight: 1, color: "var(--ads-color-text-muted)" }}
          />
        )}
        <span style={{ maxWidth: "10rem", overflow: "hidden", textOverflow: "ellipsis" }}>
          {displayName}
        </span>
        <span
          className="fr-icon-arrow-down-s-line"
          aria-hidden="true"
          style={{
            fontSize: "0.75rem",
            lineHeight: 1,
            color: "var(--ads-color-text-muted)",
            transform: isOpen ? "rotate(180deg)" : undefined,
            transition: "transform 0.15s ease",
          }}
        />
      </button>

      {isOpen && (
        <div
          ref={menuRef}
          id="account-menu-dropdown"
          role="menu"
          aria-label={t(siteAccountConfig.labelKey)}
          style={{
            position: "absolute",
            top: "100%",
            right: 0,
            marginTop: "0.25rem",
            minWidth: "14rem",
            background: "var(--ads-color-background)",
            border: "1px solid var(--ads-color-border)",
            borderRadius: "0.25rem",
            boxShadow: "0 4px 12px rgba(28, 35, 43, 0.12)",
            zIndex: 800,
          }}
        >
          {/* User info header */}
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: "0.125rem",
              padding: "0.875rem 1rem",
              borderBottom: "1px solid var(--ads-color-border)",
            }}
          >
            <span style={{ fontSize: "0.9375rem", fontWeight: 700, lineHeight: 1.3 }}>
              {displayName}
            </span>
            <span
              style={{
                fontSize: "0.8125rem",
                lineHeight: 1.4,
                color: "var(--ads-color-text-muted)",
              }}
            >
              {user.email}
            </span>
          </div>

          {/* Menu items */}
          <ul role="none" style={{ listStyle: "none", margin: "0", padding: "0.25rem 0" }}>
            {siteAccountConfig.items.map((item) => (
              <li key={item.labelKey} role="none">
                <a
                  href={item.href}
                  role="menuitem"
                  onClick={() => setIsOpen(false)}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "0.625rem",
                    padding: "0.5rem 1rem",
                    fontSize: "0.875rem",
                    lineHeight: 1.5,
                    textDecoration: "none",
                  }}
                >
                  {item.iconId && (
                    <span
                      className={item.iconId}
                      aria-hidden="true"
                      style={{ fontSize: "1.125rem", lineHeight: 1, color: "var(--ads-color-text-muted)" }}
                    />
                  )}
                  {t(item.labelKey)}
                </a>
              </li>
            ))}
          </ul>

          {/* Logout */}
          <div style={{ padding: "0.25rem 0", borderTop: "1px solid var(--ads-color-border)" }}>
            <button
              type="button"
              role="menuitem"
              onClick={() => {
                setIsOpen(false);
                void logout();
              }}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.625rem",
                width: "100%",
                padding: "0.5rem 1rem",
                fontSize: "0.875rem",
                lineHeight: 1.5,
                textAlign: "left",
                color: "var(--ads-color-danger)",
                background: "transparent",
                border: "none",
                cursor: "pointer",
              }}
            >
              <span
                className="fr-icon-logout-box-r-line"
                aria-hidden="true"
                style={{ fontSize: "1.125rem", lineHeight: 1, color: "var(--ads-color-danger)" }}
              />
              {t(siteAccountConfig.logoutLabelKey)}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
