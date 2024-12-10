export type ApplicationHealth = "Healthy" | "Degraded" | "Progressing" | "Suspended" | "Missing";

export type ApplicationStatus =
  | "Committed"
  | "WaitApproved"
  | "Succeeded"
  | "Synced"
  | "OutOfSync"
  | "Unknown"
  | "Progressing"
  | "Degraded";

export interface ApplicationSource {
  repo: string;
  path: string;
  target_revision: string;
  submodules: boolean;
  application_specifier?: {
    helm_manifest_path?: string;
  };
}

export interface ApplicationInstantiation {
  application_name: string;
  tenant_name: string;
  appcode: string;
  security: {
    external_secret: {
      secret_store_ref: {
        id: string;
      };
    };
  };
}

export interface ApplicationTarget {
  cluster: string;
  namespace: string;
}

export interface ApplicationRuntime {
  status: ApplicationStatus;
  health: ApplicationHealth;
  sync_status: string;
  git_info: any[];
  resource_metrics: {
    pod_count: number;
    secret_count: number;
    cpu: string;
    memory: string;
  };
  argocd_url: string;
  created_at?: string;
  created_by?: string;
  last_updated_at?: string;
  last_updated_by?: string;
}

export interface ApplicationTemplate {
  application_source: ApplicationSource;
  application_instantiation: ApplicationInstantiation;
  application_target: ApplicationTarget[];
  application_runtime: ApplicationRuntime;
}

export interface ApplicationResponse {
  total: number;
  success: boolean;
  message: string;
  applications: ApplicationTemplate[];
}

export interface DryRunResult {
  success: boolean;
  message: string;
  total: number;
  environments: Array<{
    environment: string;
    manifest: string;
    is_valid: boolean;
  }>;
}

export interface ValidateResult {
  success: boolean;
  type: string;
  message?: string;
  suiteable_env?: Array<{
    environments: string;
    valid: boolean;
    error: string;
  }>;
  details?: Record<string, any>;
}

export interface ValidatePayload {
  repo: string;
  target_revision: string;
  path: string;
  submodules: boolean;
  application_specifier?: {
    helm_manifest_path?: string;
  };
}

export interface CreateApplicationPayload {
  application_source: {
    repo: string;
    target_revision: string;
    path: string;
    submodules: boolean;
    application_specifier?: {
      helm_manifest_path?: string;
    };
  };
  application_instantiation: {
    application_name: string;
    tenant_name: string;
    appcode: string;
    description: string;
  };
  application_target: Array<{
    cluster: string;
    namespace: string;
  }>;
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