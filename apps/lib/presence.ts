import type { User } from "@/lib/api/types";

export type PresenceStatus = "offline" | "online" | "idle" | "dnd" | "invisible";

export interface ResolveUserPresenceOptions {
  isAuthenticated?: boolean;
  isRealtimeConnected?: boolean;
  isCurrentSession?: boolean;
}

const NON_AUTHENTICATED_USER_STATUSES = new Set(["disabled", "suspended", "deleted"]);

function normalizePresenceStatus(value: string | undefined): PresenceStatus | null {
  if (value === "online" || value === "idle" || value === "dnd" || value === "invisible") {
    return value;
  }
  return null;
}

export function resolveUserPresenceStatus(
  user: Pick<User, "status" | "presenceStatus" | "lastSeenAt"> | null,
  options: ResolveUserPresenceOptions = {}
): PresenceStatus {
  if (!user || !options.isAuthenticated) {
    return "offline";
  }

  if (NON_AUTHENTICATED_USER_STATUSES.has(user.status)) {
    return "offline";
  }

  const explicit = normalizePresenceStatus(user.presenceStatus);
  if (explicit && explicit !== "online") {
    return explicit;
  }

  if (options.isCurrentSession && options.isRealtimeConnected) {
    return "online";
  }

  return "offline";
}
