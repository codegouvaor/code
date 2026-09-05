import { apiRequest, apiListRequest } from "@/lib/api/client";
import type { Category, CreateCategoryInput, UpdateCategoryInput, PaginatedResponse } from "@/lib/api/types";

export interface CategoriesApiResponse {
  success: boolean;
  data?: Category[];
}

export const categoriesApi = {
  list(): Promise<CategoriesApiResponse> {
    return apiListRequest<Category>("/categories").then((res) => ({
      success: true,
      data: res.data,
    }));
  },

  get(id: string): Promise<Category> {
    return apiRequest<Category>(`/categories/${encodeURIComponent(id)}`);
  },

  create(input: CreateCategoryInput): Promise<{ success: boolean; data: Category }> {
    return apiRequest<Category, CreateCategoryInput>("/categories", {
      method: "POST",
      body: input,
    }).then((data) => ({ success: true, data }));
  },

  update(id: string, input: UpdateCategoryInput): Promise<Category> {
    return apiRequest<Category, UpdateCategoryInput>(`/categories/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: input,
    });
  },

  delete(id: string): Promise<{ deleted: boolean }> {
    return apiRequest<{ deleted: boolean }>(`/categories/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
  },
};
