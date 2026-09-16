import * as React from "react";
import { AuthProvider } from "@/context/AuthContext";
import { AuthGuard } from "./AuthGuard";

// ADS stylesheet (icons + components) — same as the public layout.
import "@codegouvaor/react-ads/main.css";

export default function AuthLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <AuthProvider>
          <AuthGuard>{children}</AuthGuard>
        </AuthProvider>
      </body>
    </html>
  );
}
