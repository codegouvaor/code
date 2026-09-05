import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin("./i18n/request.ts");
const isProduction = process.env.NODE_ENV === "production";
const isStaticWebBuild = process.env.BUILD_WEB_STATIC === "true";

// Shared image allow-list used by both the static export and the standard
// (server-rendered) builds so artwork hosts stay in sync. The explicit type
// keeps the object literals contextually typed (e.g. `protocol` stays the
// "https"/"http" literal union) exactly like the inline arrays did.
const imageLocalPatterns: NonNullable<NextConfig["images"]>["localPatterns"] = [
  // Allow local assets (public/) and the server-side Plex artwork
  // proxy — its encoded path arrives as a query string. `search` is
  // omitted so any query is accepted for these paths.
  { pathname: "/**" },
  { pathname: "/api/v1/integrations/plex/image" },
];
const imageRemotePatterns: NonNullable<NextConfig["images"]>["remotePatterns"] = [
  { protocol: "https", hostname: "info.gouv.aor", pathname: "/**" },
  { protocol: "https", hostname: "api.dicebear.com", pathname: "/**" },
  { protocol: "https", hostname: "avatars.githubusercontent.com", pathname: "/**" },
  { protocol: "https", hostname: "images.unsplash.com", pathname: "/**" },
  // AniList artwork CDN (cover art, banners, backdrops) serves images from
  { protocol: "http", hostname: "127.0.0.1", pathname: "/**" },
  { protocol: "http", hostname: "localhost", pathname: "/**" },
];

const nextConfig: NextConfig = {
  // Vidstack ships as ESM with non-standard `exports` conditions (e.g.
  // `development`, `deno`) and is verified to need transpilation under
  // Next.js 16 / Turbopack. See https://docs.vidstack.io/player/getting-started/installation/nextjs
  transpilePackages: ["@vidstack/react", "vidstack"],
  turbopack: {
    root: "../",
  },
  allowedDevOrigins: ["info.gouv.aor", "api.info.gouv.aor", "sso.info.gouv.aor", "info.gouv.localhost", "api.info.gouv.localhost", "sso.gouv.localhost"],
  outputFileTracingExcludes: {
    "*": ["test/**"],
  },

  reactStrictMode: true,
  poweredByHeader: false,

  ...(isStaticWebBuild
    ? {
        output: "export",
        trailingSlash: true,
        images: {
          unoptimized: true,
          localPatterns: imageLocalPatterns,
          remotePatterns: imageRemotePatterns,
        },
      }
    : {
        ...(isProduction && {
          output: "standalone",
          ...(process.env.ASSET_PREFIX && { assetPrefix: process.env.ASSET_PREFIX }),
        }),
      }),

  basePath: process.env.BASE_PATH || "",
  // TEMP-VERIFY: env override to render outside the root-owned .next dir.
  ...(process.env.NEXT_DIST_DIR ? { distDir: process.env.NEXT_DIST_DIR } : {}),

  ...(!isStaticWebBuild && {
    images: {
      localPatterns: imageLocalPatterns,
      remotePatterns: imageRemotePatterns,
    },
  }),

  ...(!isStaticWebBuild && {
    async headers() {
      const headers = [{ key: "Referrer-Policy", value: "origin-when-cross-origin" }];

      if (isProduction) {
        headers.push(
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "X-Frame-Options", value: "DENY" }
        );
      }

      return [{ source: "/(.*)", headers }];
    },

    async rewrites() {
      // Inside Docker the Go API runs in the `worker` container; in plain dev
      // it runs on localhost. This proxy is used for server-side fetches
      // (SSR + next/image optimization), not for browser requests (nginx).
      const apiBase = process.env.API_INTERNAL_URL
        ? process.env.API_INTERNAL_URL.replace(/\/api\/v1$/, "")
        : "http://localhost:8080";
      return [
        {
          source: "/api/:path*",
          destination: `${apiBase}/api/:path*`,
        },
      ];
    },
  }),
};

let configWithPlugins: NextConfig = nextConfig;

export default withNextIntl(configWithPlugins);
