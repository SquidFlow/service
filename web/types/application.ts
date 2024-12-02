export type ApplicationHealth = "Healthy" | "Degraded" | "Progressing" | "Suspended" | "Missing";

export type ApplicationStatus =
  | "Succeeded"
  | "Synced"
  | "OutOfSync"
  | "Unknown"
  | "Progressing"
  | "Degraded";

export interface ApplicationTemplate {
  id?: string;
  name: string;
  tenant_name: string;
  appcode: string;
  description: string;
  created_by: string;
  template: {
    source: {
      type: string;
      url: string;
      targetRevision: string;
      path: string;
    };
    last_commit_info: {
      Creator: string;
      LastUpdater: string;
      LastCommitID: string;
      LastCommitMessage: string;
    };
  };
  destination_clusters: {
    clusters: string[];
    namespace: string;
  };
  runtime_status: {
    status: ApplicationStatus;
    health: ApplicationHealth;
    sync_status: string;
    deployed_clusters: string[] | null;
    resource_metrics: {
      cpu_cores: string;
      memory_usage: string;
    };
    last_update?: string;
  };
  deployed_environments?: string[];
  argocd_url?: string;
}

export interface DryRunResult {
  [clusterName: string]: string;
}

export interface ValidateResult {
  success: boolean;
  message?: string;
  details?: Record<string, any>;
}

export interface ValidatePayload {
  templateSource: string;
  targetRevision: string;
  path: string;
}

export interface CreateApplicationPayload {
  application_source: {
    type: string;
    url: string;
    targetRevision: string;
    path: string;
  };
  application_name: string;
  tenant_name: string;
  appcode: string;
  description: string;
  destination_clusters: {
    clusters: string[];
    namespace: string;
  };
  ingress?: {
    host: string;
    tls?: {
      enabled: boolean;
      secretName: string;
    };
  };
  security?: {
    external_secret?: {
      secret_store_ref: {
        id: string;
      };
    };
  };
  is_dryrun: boolean;
}

export interface CreateTemplatePayload {
  name: string;
  description: string;
  source: {
    url: string;
    targetRevision: string;
    path: string;
  };
}

// ... 其他类型定义