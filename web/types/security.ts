export interface SecretStore {
  id: string;
  name: string;
  provider: "AWS" | "GCP" | "Azure" | "Vault" | "CyberArk";
  type: string;
  status: "Active" | "Inactive" | "Error";
  path?: string;
  lastSynced: string;
  createdAt: string;
  lastUpdated: string;
  health: {
    status: "Healthy" | "Warning" | "Error";
    message?: string;
  };
}