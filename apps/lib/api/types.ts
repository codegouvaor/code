export interface ApiMeta {
  requestId?: string;
}

export interface PaginationMeta extends ApiMeta {
  nextCursor?: string;
  hasMore?: boolean;
}

export interface ApiSuccessEnvelope<T> {
  data: T;
  meta?: ApiMeta;
}

export interface ApiListEnvelope<T> {
  data: T[];
  meta?: PaginationMeta;
}

export interface ApiFailureEnvelope {
  error: {
    code?: string;
    message?: string;
    details?: unknown;
  };
  meta?: ApiMeta;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: PaginationMeta;
}

export interface User {
  id: string;
  email: string;
  name?: string;
  displayName: string;
  avatarUrl?: string;
  status: string;
  presenceStatus?: string;
  workspaceId?: string;
  roles?: string[];
  permissions?: string[];
  createdAt: string;
  updatedAt: string;
  lastSeenAt?: string;
  disabledAt?: string;
  emailVerifiedAt?: string;
  passwordChangedAt?: string;
}

export interface Workspace {
  id: string;
  name: string;
  slug: string;
  description?: string;
  visibility: string;
  ownerId: string;
  createdAt: string;
  updatedAt: string;
  archivedAt?: string;
}

export type WorkspaceSsoProvider = "oidc" | "saml";

export interface WorkspaceSsoConfig {
  id?: string;
  workspaceId: string;
  provider: WorkspaceSsoProvider;
  enabled: boolean;
  enforceSso: boolean;
  allowPasswordAuth: boolean;
  allowAutoProvision: boolean;
  allowIdpInitiated: boolean;
  domainHint?: string;
  issuerUrl?: string;
  ssoUrl?: string;
  entityId?: string;
  clientId?: string;
  clientSecretConfigured: boolean;
  certificate?: string;
  allowedDomains: string[];
  defaultRole: WorkspaceMemberRole;
  attributeMapping: Record<string, string>;
  createdAt?: string;
  updatedAt?: string;
}

export interface UpdateWorkspaceSsoInput {
  provider: WorkspaceSsoProvider;
  enabled: boolean;
  enforceSso: boolean;
  allowPasswordAuth: boolean;
  allowAutoProvision: boolean;
  allowIdpInitiated: boolean;
  domainHint?: string;
  issuerUrl?: string;
  ssoUrl?: string;
  entityId?: string;
  clientId?: string;
  clientSecret?: string;
  clearClientSecret?: boolean;
  certificate?: string;
  allowedDomains: string[];
  defaultRole: WorkspaceMemberRole;
  attributeMapping: Record<string, string>;
}

export interface WorkspaceMember {
  id: string;
  workspaceId: string;
  userId: string;
  role: string;
  joinedAt: string;
  lastSeenAt?: string;
  createdAt: string;
  updatedAt: string;
  displayName?: string;
  email?: string;
  avatarUrl?: string;
  status?: string;
  presenceStatus?: string;
}

export type WorkspaceMemberRole = "owner" | "admin" | "member" | "guest";

export interface Team {
  id: string;
  workspaceId: string;
  name: string;
  description?: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  archivedAt?: string;
}

export interface Channel {
  id: string;
  workspaceId: string;
  teamId?: string;
  name: string;
  slug: string;
  description?: string;
  type: string;
  visibility: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  archivedAt?: string;
}

export interface Conversation {
  id: string;
  workspaceId: string;
  channelId?: string;
  type: string;
  name?: string;
  memberIds?: string[];
  participants?: Array<{
    userId: string;
    displayName: string;
    email: string;
    role: string;
    status: string;
    presenceStatus: string;
  }>;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  archivedAt?: string;
}

export interface ConversationMember {
  id: string;
  conversationId: string;
  userId: string;
  role: string;
  joinedAt: string;
  lastReadMessageId?: string;
  lastReadAt?: string;
  mutedUntil?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Message {
  id: string;
  workspaceId: string;
  conversationId: string;
  authorId: string;
  parentMessageId?: string;
  type: string;
  content: string;
  metadata?: Record<string, unknown>;
  editedAt?: string;
  deletedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface MessageReaction {
  id: string;
  messageId: string;
  userId: string;
  emoji: string;
  createdAt: string;
  updatedAt: string;
}

export interface Meeting {
  id: string;
  workspaceId: string;
  conversationId?: string;
  provider: string;
  providerRoomId?: string;
  title: string;
  status: string;
  createdBy: string;
  startedAt?: string;
  endedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface MeetingParticipant {
  id: string;
  meetingId: string;
  userId: string;
  role: string;
  status: string;
  metadata?: Record<string, unknown>;
  joinedAt?: string;
  leftAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface MeetingSession {
  id: string;
  meetingId: string;
  workspaceId: string;
  provider: string;
  providerRoomName: string;
  providerRoomId?: string;
  nodeId: string;
  status: string;
  publicUrl: string;
  signalingUrl: string;
  connectionDetails?: Record<string, unknown>;
  startedAt?: string;
  endedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface MeetingJoinCredentials {
  meetingId: string;
  sessionId: string;
  roomName: string;
  participantIdentity: string;
  token: string;
  signalingUrl: string;
  expiresAt: string;
  iceServers?: RTCIceServer[];
}

export interface JoinTokenResponse {
  token: string;
  meetingId: string;
  sessionId: string;
  roomName: string;
  participantIdentity: string;
  signalingUrl: string;
  expiresAt: string;
  iceServers?: RTCIceServer[];
}

export interface TokenResponse {
  accessToken: string;
  refreshToken?: string;
  tokenType?: string;
  expiresIn?: number;
  user: User;
  sessionId?: string;
}

export interface AuthSessionInfo {
  id: string;
  userAgent?: string;
  ipAddress?: string;
  createdAt: string;
  lastUsedAt: string;
  expiresAt: string;
  current: boolean;
}

export interface Application {
  id: string;
  workspaceId: string;
  provider: string;
  name: string;
  status: string;
  configuration?: Record<string, unknown>;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuditLog {
  id: string;
  workspaceId: string;
  actorId: string;
  action: string;
  resourceType: string;
  resourceId: string;
  metadata?: Record<string, unknown>;
  ipAddress?: string;
  userAgent?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Notification {
  id: string;
  workspaceId: string;
  userId: string;
  type: string;
  title: string;
  body: string;
  resourceType?: string;
  resourceId?: string;
  metadata?: Record<string, unknown>;
  readAt?: string;
  idempotencyKey?: string;
  createdAt: string;
  updatedAt: string;
}

export interface NotificationPreferences {
  directMessages: boolean;
  mentions: boolean;
  channelMessages: boolean;
  meetingReminders: boolean;
  incomingCalls: boolean;
  emailNotifications: boolean;
  sounds: boolean;
  desktopNotifications: boolean;
}

export interface UserPreferences {
  theme: "light" | "dark" | "system";
  language: string;
  locale: string;
  timezone: string;
  statusMessage?: string;
  density: "comfortable" | "compact";
  contrast: "default" | "high";
  soundEnabled: boolean;
  secureSession: boolean;
}

export interface ProvisionWorkspaceUserInput {
  email: string;
  displayName: string;
  role: WorkspaceMemberRole;
  temporaryPassword: string;
}

export interface CreateWorkspaceMemberInput {
  userId?: string;
  email?: string;
  role: WorkspaceMemberRole;
}

export interface UpdateWorkspaceMemberInput {
  role: WorkspaceMemberRole;
}

export interface Contact {
  id: string;
  workspaceId?: string;
  name: string;
  email?: string;
  phone?: string;
  [key: string]: unknown;
}

export interface ContactGroup {
  id: string;
  workspaceId?: string;
  name: string;
  description?: string;
  [key: string]: unknown;
}

export interface Task {
  id: string;
  workspaceId: string;
  title: string;
  description?: string;
  status: string;
  priority?: string;
  project?: string;
  assigneeUserId?: string;
  assigneeName?: string;
  createdBy?: string;
  dueAt?: string;
  completedAt?: string;
  createdAt?: string;
  updatedAt?: string;
  [key: string]: unknown;
}

export interface TaskComment {
  id: string;
  taskId?: string;
  content?: string;
  [key: string]: unknown;
}

export interface Project {
  id: string;
  workspaceId: string;
  name: string;
  summary?: string;
  status: string;
  progress?: number;
  cadence?: string;
  ownerUserId?: string;
  ownerName?: string;
  createdBy?: string;
  [key: string]: unknown;
}

export interface ProjectMember {
  id: string;
  projectId?: string;
  userId?: string;
  role?: string;
  [key: string]: unknown;
}

export interface FileRecord {
  id: string;
  workspaceId?: string;
  name?: string;
  contentType?: string;
  status?: string;
  sizeBytes?: number;
  storageKey?: string;
  metadata?: Record<string, unknown>;
  createdAt?: string;
  updatedAt?: string;
}

export interface Document {
  id: string;
  workspaceId?: string;
  title?: string;
  [key: string]: unknown;
}

export interface Resource {
  id: string;
  workspaceId?: string;
  title?: string;
  [key: string]: unknown;
}

export interface CallHistoryItem {
  id: string;
  workspaceId?: string;
  [key: string]: unknown;
}

export interface Voicemail {
  id: string;
  callId?: string;
  [key: string]: unknown;
}

export interface RealtimeEvent<T = unknown> {
  id: string;
  type: string;
  workspaceId?: string;
  resourceId?: string;
  occurredAt: string;
  data: T;
}

// ─── Content Types ─────────────────────────────────────────────

export type ArticleType = "article" | "annonce" | "note_technique" | "dossier" | "communique" | "analyse";
export type ArticleStatus = "draft" | "writing" | "review" | "scheduled" | "published" | "archived";

export interface Article {
  id: string;
  title: string;
  slug: string;
  excerpt: string;
  content: string;
  type: ArticleType;
  status: ArticleStatus;
  categoryId?: string;
  category?: Category;
  team: string;
  authorId: string;
  author?: User;
  publishedAt?: string;
  scheduledAt?: string;
  tags: string[];
  seoTitle: string;
  seoDescription: string;
  seoOgImage: string;
  priority: string;
  channel: string;
  views: number;
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateArticleInput {
  title: string;
  excerpt?: string;
  content?: string;
  type?: ArticleType;
  categoryId?: string;
  team?: string;
  tags?: string[];
  seoTitle?: string;
  seoDescription?: string;
  seoOgImage?: string;
  seoKeywords?: string[];
  priority?: string;
  channel?: string;
}

export interface UpdateArticleInput {
  title?: string;
  excerpt?: string;
  content?: string;
  type?: ArticleType;
  categoryId?: string;
  team?: string;
  tags?: string[];
  seoTitle?: string;
  seoDescription?: string;
  seoOgImage?: string;
  priority?: string;
  channel?: string;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  description: string;
  color: string;
  sortOrder: number;
  parentId?: string;
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateCategoryInput {
  name: string;
  description?: string;
  color?: string;
  parentId?: string;
}

export interface UpdateCategoryInput {
  name?: string;
  description?: string;
  color?: string;
  sortOrder?: number;
}

export interface Tag {
  id: string;
  name: string;
  slug: string;
  createdAt: string;
  updatedAt: string;
}

export interface MediaItem {
  id: string;
  name: string;
  fileName: string;
  originalName?: string;
  filename?: string;
  url: string;
  mimeType: string;
  size: number;
  width?: number;
  height?: number;
  alt: string;
  caption: string;
  uploadedBy: string;
  workspaceId?: string;
  metadata?: Record<string, unknown>;
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export type Media = MediaItem;

export interface CreateMediaInput {
  name: string;
  fileName: string;
  url: string;
  mimeType: string;
  size: number;
  alt?: string;
  caption?: string;
}

export interface UpdateMediaInput {
  name?: string;
  alt?: string;
  caption?: string;
}

export interface Webhook {
  id: string;
  workspaceId: string;
  provider: string;
  url: string;
  events: string[];
  active: boolean;
  lastTriggeredAt?: string;
  failureCount: number;
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateWebhookInput {
  provider: string;
  url: string;
  secret?: string;
  events: string[];
}

export interface UpdateWebhookInput {
  url?: string;
  secret?: string;
  events?: string[];
  active?: boolean;
}

export interface WebhookDelivery {
  id: string;
  webhookId: string;
  event: string;
  status: number;
  duration: number;
  error?: string;
  createdAt: string;
  updatedAt: string;
}

export interface SeoConfig {
  id: string;
  pagePath: string;
  title: string;
  description: string;
  ogImage: string;
  canonical: string;
  noIndex: boolean;
  keywords: string;
  locale: string;
  createdAt: string;
  updatedAt: string;
}

export interface UpsertSeoConfigInput {
  pagePath: string;
  title: string;
  description?: string;
  ogImage?: string;
  canonical?: string;
  noIndex?: boolean;
  keywords?: string;
  locale?: string;
}

export interface NewsletterSubscriber {
  id: string;
  email: string;
  status: string;
  workspaceId?: string;
  subscribedAt: string;
  unsubscribedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Schedule {
  id: string;
  entityType: string;
  entityId: string;
  title: string;
  scheduledAt: string;
  status: string;
  publishedAt?: string;
  cancelledAt?: string;
  workspaceId?: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateScheduleInput {
  entityType: string;
  entityId?: string;
  title: string;
  scheduledAt: string;
}
