import { apiRequest, apiListRequest } from "@/lib/api/client";
import type { SeoConfig, UpsertSeoConfigInput, PaginatedResponse } from "@/lib/api/types";

export const seoApi = {
  list(): Promise<PaginatedResponse<SeoConfig>> {
    return apiListRequest<SeoConfig>("/seo");
  },

  getConfig(pagePath: string): Promise<SeoConfig> {
    const params = new URLSearchParams({ pagePath });
    return apiRequest<SeoConfig>(`/seo/config?${params.toString()}`);
  },

  upsert(input: UpsertSeoConfigInput): Promise<SeoConfig> {
    return apiRequest<SeoConfig, UpsertSeoConfigInput>("/seo/config", {
      method: "POST",
      body: input,
    });
  },
};
