"use client";

import * as React from "react";
import { Input } from "@codegouvaor/react-ads/Input";
import { PasswordInput } from "@codegouvaor/react-ads/blocks/PasswordInput";
import { Button } from "@codegouvaor/react-ads/Button";
import { Checkbox } from "@codegouvaor/react-ads/Checkbox";
import { useAuth } from "@/context/AuthContext";

export default function LoginForm() {
  const { login, isLoading } = useAuth();

  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [rememberMe, setRememberMe] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = React.useState(false);

  const emailId = React.useId();
  const passwordId = React.useId();

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);

    if (!email.trim()) {
      setError("Please enter your email address.");
      return;
    }

    if (!password) {
      setError("Please enter your password.");
      return;
    }

    setIsSubmitting(true);

    try {
      await login(email, password, rememberMe);
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

      {/* Email */}
      <Input
        id={emailId}
        label="Email address"
        hintText="Enter the email address linked to your account."
        nativeInputProps={{
          type: "email",
          autoComplete: "email",
          autoFocus: true,
          value: email,
          onChange: (e) => {
            setEmail(e.target.value);
            if (error) setError(null);
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
          autoComplete: "current-password",
          value: password,
          onChange: (e) => {
            setPassword(e.target.value);
            if (error) setError(null);
          },
          disabled: isSubmitting || isLoading,
        }}
      />

      {/* Remember me + forgot password */}
      <div className="gov-login-form__options">
        <Checkbox
          type="checkbox"
          legend={<span className="sr-only">Options</span>}
          options={[
            {
              label: "Remember me",
              nativeInputProps: {
                name: "remember-me",
                checked: rememberMe,
                onChange: (e) => setRememberMe(e.target.checked),
                disabled: isSubmitting || isLoading,
              },
            },
          ]}
        />

        <a href="/forgot-password" className="gov-login-form__forgot-link">
          Forgot your password?
        </a>
      </div>

      {/* Submit */}
      <div className="gov-login-form__submit">
        <Button
          type="submit"
          priority="primary"
          size="large"
          disabled={isSubmitting || isLoading}
          iconId={
            isSubmitting ? "fr-icon-refresh-line" : "fr-icon-lock-line"
          }
          iconPosition="left"
        >
          {isSubmitting ? "Signing in…" : "Sign in"}
        </Button>
      </div>

      {/* Divider */}
      <div className="gov-login-form__divider">
        <span className="gov-login-form__divider-text">or</span>
      </div>

      {/* Register link */}
      <div className="gov-login-form__register">
        <p className="gov-login-form__register-text">
          Don&apos;t have an account?{" "}
          <a href="/register" className="gov-login-form__register-link">
            Create one
          </a>
        </p>
      </div>
    </form>
  );
}
