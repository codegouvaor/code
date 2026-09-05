import { apiRequest } from "@/lib/api/client";

export interface SiteSettings {
  siteName: string;
  siteDescription: string;
  siteUrl: string;
  email: string;
  smtpHost: string;
  smtpPort: string;
  smtpUser: string;
  smtpPassword: string;
  fromName: string;
  fromEmail: string;
  maintenanceMode: boolean;
  registrationOpen: boolean;
  commentsEnabled: boolean;
  newsletterEnabled: boolean;
  analyticsEnabled: boolean;
  sslEnforced: boolean;
}

export interface UpdateSiteSettingsInput {
  siteName?: string;
  siteDescription?: string;
  siteUrl?: string;
  email?: string;
  maintenanceMode?: boolean;
  registrationOpen?: boolean;
  commentsEnabled?: boolean;
  newsletterEnabled?: boolean;
  analyticsEnabled?: boolean;
  sslEnforced?: boolean;
}

export const settingsApi = {
  async get(): Promise<{ success: boolean; data: SiteSettings }> {
    try {
      const data = await apiRequest<SiteSettings>("/admin/settings");
      return { success: true, data };
    } catch {
      return {
        success: false,
        data: {
          siteName: "The Etheria Times",
          siteDescription: "L'information au service du citoyen",
          siteUrl: "https://etheriatimes.com",
          email: "contact@etheriatimes.com",
          smtpHost: "",
          smtpPort: "587",
          smtpUser: "",
          smtpPassword: "",
          fromName: "",
          fromEmail: "",
          maintenanceMode: false,
          registrationOpen: true,
          commentsEnabled: true,
          newsletterEnabled: true,
          analyticsEnabled: true,
          sslEnforced: true,
        },
      };
    }
  },

  async update(input: UpdateSiteSettingsInput): Promise<{ success: boolean }> {
    try {
      await apiRequest("/admin/settings", {
        method: "PATCH",
        body: input,
      });
      return { success: true };
    } catch {
      return { success: false };
    }
  },
};
