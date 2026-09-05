import { apiRequest, apiListRequest } from "@/lib/api/client";
import type { Schedule, CreateScheduleInput, PaginatedResponse } from "@/lib/api/types";

export interface ListScheduleOptions {
  from?: string;
  to?: string;
  page?: number;
  pageSize?: number;
}

export const scheduleApi = {
  list(options?: ListScheduleOptions): Promise<PaginatedResponse<Schedule>> {
    const params = new URLSearchParams();
    if (options?.from) params.set("from", options.from);
    if (options?.to) params.set("to", options.to);
    if (options?.page) params.set("page", String(options.page));
    if (options?.pageSize) params.set("pageSize", String(options.pageSize));
    const suffix = params.size > 0 ? `?${params.toString()}` : "";
    return apiListRequest<Schedule>(`/schedule${suffix}`);
  },

  get(id: string): Promise<Schedule> {
    return apiRequest<Schedule>(`/schedule/${encodeURIComponent(id)}`);
  },

  create(input: CreateScheduleInput): Promise<Schedule> {
    return apiRequest<Schedule, CreateScheduleInput>("/schedule", {
      method: "POST",
      body: input,
    });
  },

  cancel(id: string): Promise<Schedule> {
    return apiRequest<Schedule>(`/schedule/${encodeURIComponent(id)}/cancel`, {
      method: "POST",
    });
  },

  delete(id: string): Promise<{ deleted: boolean }> {
    return apiRequest<{ deleted: boolean }>(`/schedule/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
  },
};
