"use client";

import * as React from "react";
import { Input } from "@codegouvaor/react-ads/Input";
import { PasswordInput } from "@codegouvaor/react-ads/blocks/PasswordInput";
import { Button } from "@codegouvaor/react-ads/Button";
import { useAuth } from "@/context/AuthContext";

export default function RegisterForm() {
  const { register, isLoading } = useAuth();

  const [displayName, setDisplayName] = React.useState("");
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [error, setError] = React.useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = React.useState(false);

  const nameId = React.useId();
  const emailId = React.useId();
  const passwordId = React.useId();
  const confirmId = React.useId();

  function clearError() {
    if (error) setError(null);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);

    if (!displayName.trim()) {
      setError("Please enter your name.");
      return;
    }

    if (!email.trim()) {
      setError("Please enter your email address.");
      return;
    }

    if (!password) {
      setError("Please choose a password.");
      return;
    }

    if (password.length < 8) {
      setError("Your password must be at least 8 characters.");
      return;
    }

    if (password !== confirmPassword) {
      setError("The passwords do not match.");
      return;
    }

    setIsSubmitting(true);

    try {
      await register({ displayName: displayName.trim(), email, password });
    } catch (err: unknown) {
      const message =
        err instanceof Error
          ? err.message
          : "An error occurred. Please try again.";
      setError(message);
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <form className="gov-login-form" onSubmit={handleSubmit} noValidate>
      {/* Error alert */}
      {error && (
        <div className="gov-login-form__alert" role="alert">
          <span className="fr-icon-error-line" aria-hidden="true" />
          <p className="gov-login-form__alert-text">{error}</p>
        </div>
      )}

      {/* Name */}
      <Input
        id={nameId}
        label="Full name"
        hintText="The name displayed on your profile."
        nativeInputProps={{
          type: "text",
          autoComplete: "name",
          autoFocus: true,
          value: displayName,
          onChange: (e) => {
            setDisplayName(e.target.value);
            clearError();
          },
          disabled: isSubmitting || isLoading,
        }}
      />

      {/* Email */}
      <Input
        id={emailId}
        label="Email address"
        hintText="You will use this address to sign in."
        nativeInputProps={{
          type: "email",
          autoComplete: "email",
          value: email,
          onChange: (e) => {
            setEmail(e.target.value);
            clearError();
          },
          disabled: isSubmitting || isLoading,
        }}
      />

      {/* Password */}
      <PasswordInput
        id={passwordId}
        label="Password"
        hintText=""
        messagesHint=""
        nativeInputProps={{
          autoComplete: "new-password",
          value: password,
          onChange: (e) => {
            setPassword(e.target.value);
            clearError();
          },
          disabled: isSubmitting || isLoading,
        }}
      />

      {/* Confirm password */}
      <PasswordInput
        id={confirmId}
        label="Confirm password"
        hintText=""
        messagesHint=""
        nativeInputProps={{
          autoComplete: "new-password",
          value: confirmPassword,
          onChange: (e) => {
            setConfirmPassword(e.target.value);
            clearError();
          },
          disabled: isSubmitting || isLoading,
        }}
      />

      {/* Submit */}
      <div className="gov-login-form__submit">
        <Button
          type="submit"
          priority="primary"
          size="large"
          disabled={isSubmitting || isLoading}
          iconId={
            isSubmitting ? "fr-icon-refresh-line" : "fr-icon-user-add-line"
          }
          iconPosition="left"
        >
          {isSubmitting ? "Creating account…" : "Create account"}
        </Button>
      </div>

      {/* Divider */}
      <div className="gov-login-form__divider">
        <span className="gov-login-form__divider-text">or</span>
      </div>

      {/* Login link */}
      <div className="gov-login-form__register">
        <p className="gov-login-form__register-text">
          Already have an account?{" "}
          <a href="/login" className="gov-login-form__register-link">
            Sign in
          </a>
        </p>
      </div>
    </form>
  );
}
