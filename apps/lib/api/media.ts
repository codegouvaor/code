import { apiRequest, apiListRequest } from "@/lib/api/client";
import type { MediaItem, CreateMediaInput, UpdateMediaInput, PaginatedResponse } from "@/lib/api/types";

export type { MediaItem as Media } from "@/lib/api/types";

export interface ListMediaOptions {
  mimeType?: string;
  page?: number;
  pageSize?: number;
}

export interface MediaApiResponse {
  success: boolean;
  data?: MediaItem[];
}

export const mediaApi = {
  list(options?: ListMediaOptions): Promise<MediaApiResponse> {
    const params = new URLSearchParams();
    if (options?.mimeType) params.set("mimeType", options.mimeType);
    if (options?.page) params.set("page", String(options.page));
    if (options?.pageSize) params.set("pageSize", String(options.pageSize));
    const suffix = params.size > 0 ? `?${params.toString()}` : "";
    return apiListRequest<MediaItem>(`/media${suffix}`).then((res) => ({
      success: true,
      data: res.data,
    }));
  },

  get(id: string): Promise<MediaItem> {
    return apiRequest<MediaItem>(`/media/${encodeURIComponent(id)}`);
  },

  create(input: CreateMediaInput): Promise<MediaItem> {
    return apiRequest<MediaItem, CreateMediaInput>("/media", {
      method: "POST",
      body: input,
    });
  },

  upload(file: File, metadata?: { name?: string; alt?: string; caption?: string }): Promise<{ success: boolean; data: MediaItem }> {
    const body = {
      name: metadata?.name ?? file.name,
      fileName: file.name,
      url: URL.createObjectURL(file),
      mimeType: file.type,
      size: file.size,
      alt: metadata?.alt ?? "",
      caption: metadata?.caption ?? "",
    };
    return apiRequest<MediaItem, typeof body>("/media", {
      method: "POST",
      body,
    }).then((data) => ({ success: true, data }));
  },

  update(id: string, input: UpdateMediaInput): Promise<MediaItem> {
    return apiRequest<MediaItem, UpdateMediaInput>(`/media/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: input,
    });
  },

  delete(id: string): Promise<{ deleted: boolean }> {
    return apiRequest<{ deleted: boolean }>(`/media/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
  },
};
