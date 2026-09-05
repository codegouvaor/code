import { apiRequest } from "@/lib/api/client";

export interface UpdateCheckResult {
  hasUpdate: boolean;
  currentVersion?: string;
  latestVersion?: string;
}

export interface UpdateResult {
  success: boolean;
  message: string;
}

export interface DockerLogsResult {
  success: boolean;
  data?: { logs: string[] };
  logs?: string[];
}

export interface DockerExecResult {
  success: boolean;
  output: string;
}

export const dockerApi = {
  async checkForUpdates(): Promise<UpdateCheckResult> {
    try {
      return await apiRequest<UpdateCheckResult>("/admin/docker/updates", {
        skipRefresh: true,
      });
    } catch {
      return { hasUpdate: false };
    }
  },

  async updateContainer(image: string): Promise<UpdateResult> {
    try {
      return await apiRequest<UpdateResult, { image: string }>(
        "/admin/docker/update",
        {
          method: "POST",
          body: { image },
        }
      );
    } catch (error) {
      const message = error instanceof Error ? error.message : "Update failed";
      return { success: false, message };
    }
  },

  async getLogs(
    service: string,
    tail = 100
  ): Promise<DockerLogsResult> {
    try {
      return await apiRequest<DockerLogsResult>(
        `/admin/docker/logs?service=${service}&tail=${tail}`,
        { skipRefresh: true }
      );
    } catch {
      return { success: false };
    }
  },

  async execCommand(
    command: string,
    service = "server"
  ): Promise<DockerExecResult> {
    try {
      return await apiRequest<DockerExecResult, { command: string; service: string }>(
        "/admin/docker/exec",
        {
          method: "POST",
          body: { command, service },
        }
      );
    } catch {
      return { success: false, output: "" };
    }
  },
};
