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
    <div className="gov-account-menu">
      <button
        ref={buttonRef}
        type="button"
        className="gov-account-menu__trigger"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="true"
        aria-controls="account-menu-dropdown"
      >
        {user.avatarUrl ? (
          <img
            src={user.avatarUrl}
            alt=""
            className="gov-account-menu__avatar"
            width={28}
            height={28}
          />
        ) : (
          <span className="fr-icon-account-circle-line gov-account-menu__icon" aria-hidden="true" />
        )}
        <span className="gov-account-menu__name">{displayName}</span>
        <span
          className={`fr-icon-arrow-down-s-line gov-account-menu__chevron ${isOpen ? "gov-account-menu__chevron--open" : ""}`}
          aria-hidden="true"
        />
      </button>

      {isOpen && (
        <div
          ref={menuRef}
          id="account-menu-dropdown"
          className="gov-account-menu__dropdown"
          role="menu"
          aria-label={t(siteAccountConfig.labelKey)}
        >
          {/* User info header */}
          <div className="gov-account-menu__header">
            <span className="gov-account-menu__header-name">{displayName}</span>
            <span className="gov-account-menu__header-email">{user.email}</span>
          </div>

          {/* Menu items */}
          <ul className="gov-account-menu__list" role="none">
            {siteAccountConfig.items.map((item) => (
              <li key={item.labelKey} role="none">
                <a
                  href={item.href}
                  className="gov-account-menu__item"
                  role="menuitem"
                  onClick={() => setIsOpen(false)}
                >
                  {item.iconId && (
                    <span className={`${item.iconId} gov-account-menu__item-icon`} aria-hidden="true" />
                  )}
                  {t(item.labelKey)}
                </a>
              </li>
            ))}
          </ul>

          {/* Logout */}
          <div className="gov-account-menu__footer">
            <button
              type="button"
              className="gov-account-menu__item gov-account-menu__item--destructive"
              role="menuitem"
              onClick={() => {
                setIsOpen(false);
                void logout();
              }}
            >
              <span className="fr-icon-logout-box-r-line gov-account-menu__item-icon" aria-hidden="true" />
              {t(siteAccountConfig.logoutLabelKey)}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
