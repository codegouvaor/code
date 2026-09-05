"use client";

import { AuthProvider } from "@/context/AuthContext";
import { LicenseProvider } from "@/context/LicenseContext";
import { PlatformProvider } from "@/context/PlatformContext";
import { CastProvider } from "@/context/CastContext";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
      <LicenseProvider>
        <AuthProvider>
          <PlatformProvider>
            <CastProvider>
              {children}
              <Toaster richColors position="top-right" />
            </CastProvider>
          </PlatformProvider>
        </AuthProvider>
      </LicenseProvider>
    </ThemeProvider>
  );
}
