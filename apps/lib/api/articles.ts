import { apiRequest, apiListRequest } from "@/lib/api/client";
import type { Article, CreateArticleInput, UpdateArticleInput, PaginatedResponse } from "@/lib/api/types";

export interface ListArticlesOptions {
  status?: string;
  categoryId?: string;
  page?: number;
  pageSize?: number;
}

export const articlesApi = {
  list(options?: ListArticlesOptions): Promise<PaginatedResponse<Article>> {
    const params = new URLSearchParams();
    if (options?.status) params.set("status", options.status);
    if (options?.categoryId) params.set("categoryId", options.categoryId);
    if (options?.page) params.set("page", String(options.page));
    if (options?.pageSize) params.set("pageSize", String(options.pageSize));
    const suffix = params.size > 0 ? `?${params.toString()}` : "";
    return apiListRequest<Article>(`/articles${suffix}`);
  },

  get(id: string): Promise<Article> {
    return apiRequest<Article>(`/articles/${encodeURIComponent(id)}`);
  },

  getBySlug(slug: string): Promise<Article> {
    return apiRequest<Article>(`/articles/slug/${encodeURIComponent(slug)}`);
  },

  create(input: CreateArticleInput): Promise<Article> {
    return apiRequest<Article, CreateArticleInput>("/articles", {
      method: "POST",
      body: input,
    });
  },

  update(id: string, input: UpdateArticleInput): Promise<Article> {
    return apiRequest<Article, UpdateArticleInput>(`/articles/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: input,
    });
  },

  delete(id: string): Promise<{ deleted: boolean }> {
    return apiRequest<{ deleted: boolean }>(`/articles/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
  },

  publish(id: string): Promise<Article> {
    return apiRequest<Article>(`/articles/${encodeURIComponent(id)}/publish`, {
      method: "POST",
    });
  },

  schedule(id: string, scheduledAt: string): Promise<Article> {
    return apiRequest<Article, { scheduledAt: string }>(`/articles/${encodeURIComponent(id)}/schedule`, {
      method: "POST",
      body: { scheduledAt },
    });
  },

  archive(id: string): Promise<Article> {
    return apiRequest<Article>(`/articles/${encodeURIComponent(id)}/archive`, {
      method: "POST",
    });
  },
};
