import { apiRequest, apiListRequest } from "@/lib/api/client";
import type { NewsletterSubscriber, PaginatedResponse } from "@/lib/api/types";

export interface ListSubscribersOptions {
  status?: string;
  page?: number;
  pageSize?: number;
}

export const newsletterApi = {
  listSubscribers(options?: ListSubscribersOptions): Promise<PaginatedResponse<NewsletterSubscriber>> {
    const params = new URLSearchParams();
    if (options?.status) params.set("status", options.status);
    if (options?.page) params.set("page", String(options.page));
    if (options?.pageSize) params.set("pageSize", String(options.pageSize));
    const suffix = params.size > 0 ? `?${params.toString()}` : "";
    return apiListRequest<NewsletterSubscriber>(`/newsletter/subscribers${suffix}`);
  },

  subscribe(email: string): Promise<{ subscribed: boolean }> {
    return apiRequest<{ subscribed: boolean }, { email: string }>("/newsletter/subscribe", {
      method: "POST",
      body: { email },
    });
  },

  unsubscribe(email: string): Promise<{ unsubscribed: boolean }> {
    return apiRequest<{ unsubscribed: boolean }, { email: string }>("/newsletter/unsubscribe", {
      method: "DELETE",
      body: { email },
    });
  },
};
