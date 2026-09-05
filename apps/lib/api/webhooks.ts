import { apiRequest, apiListRequest } from "@/lib/api/client";
import type { Webhook, CreateWebhookInput, UpdateWebhookInput, WebhookDelivery, PaginatedResponse } from "@/lib/api/types";

export const webhooksApi = {
  list(): Promise<PaginatedResponse<Webhook>> {
    return apiListRequest<Webhook>("/webhooks");
  },

  get(id: string): Promise<Webhook> {
    return apiRequest<Webhook>(`/webhooks/${encodeURIComponent(id)}`);
  },

  create(input: CreateWebhookInput): Promise<Webhook> {
    return apiRequest<Webhook, CreateWebhookInput>("/webhooks", {
      method: "POST",
      body: input,
    });
  },

  update(id: string, input: UpdateWebhookInput): Promise<Webhook> {
    return apiRequest<Webhook, UpdateWebhookInput>(`/webhooks/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: input,
    });
  },

  delete(id: string): Promise<{ deleted: boolean }> {
    return apiRequest<{ deleted: boolean }>(`/webhooks/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
  },

  test(id: string): Promise<{ delivered: boolean; deliveryId: string }> {
    return apiRequest<{ delivered: boolean; deliveryId: string }>(`/webhooks/${encodeURIComponent(id)}/test`, {
      method: "POST",
    });
  },

  listDeliveries(webhookId: string, limit?: number): Promise<PaginatedResponse<WebhookDelivery>> {
    const params = new URLSearchParams();
    if (limit) params.set("limit", String(limit));
    const suffix = params.size > 0 ? `?${params.toString()}` : "";
    return apiListRequest<WebhookDelivery>(`/webhooks/${encodeURIComponent(webhookId)}/deliveries${suffix}`);
  },
};
