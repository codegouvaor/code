import * as React from "react";
import { AuthProvider } from "@/context/AuthContext";
import { AuthGuard } from "./AuthGuard";

// ADS stylesheet (icons + components) — same as the public layout.
import "@codegouvaor/react-ads/main.css";
// Portal layer (login page styles live here).
import "@/styles/globals.css";

export default function AuthLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning className="select-none">
      <body className="gov-ads">
        <AuthProvider>
          <AuthGuard>{children}</AuthGuard>
        </AuthProvider>
      </body>
    </html>
  );
}
