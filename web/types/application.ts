export interface ApplicationTemplate {
  id: number;
  name: string;
  owner: string;
  description?: string;
  path: string;
  environments: string[];
  destination_clusters: {
    clusters: string[];
  };
  appType: "kustomize" | "helm" | "helm+kustomize";
  source: {
    url: string;
    targetRevision: string;
  };
  runtime_status: {
    health: ApplicationHealth;
    status: ApplicationStatus;
  };
  template: {
    last_commit_info: {
      LastUpdater: string;
    };
    source: {
      url: string;
    };
  };
  uri: string;
  lastUpdate: string;
  creator: string;
  lastUpdater: string;
  lastCommitId: string;
  lastCommitLog: string;
  podCount: number;
  cpuCount: string;
  memoryUsage: string;
  storageUsage: string;
  memoryAmount: string;
  secretCount: number;
  resources: {
    [cluster: string]: {
      cpu: string;
      memory: string;
      storage: string;
      pods: number;
    };
  };
  deploymentStats: {
    deployments: number;
    services: number;
    configmaps: number;
  };
  worklog: Array<{
    date: string;
    action: string;
    user: string;
  }>;
  remoteRepo: {
    url: string;
    branch: string;
    baseCommitUrl: string;
    latestCommit: {
      id: string;
      message: string;
      author: string;
      timestamp: string;
    };
  };
  deployed_environments: string[];
  health: ApplicationHealth;
  argocdUrl: string;
  events: Array<{
    time: string;
    type: string;
  }>;
  metadata: {
    createdAt: string;
    updatedAt: string;
    version: string;
  };
}

export type ApplicationHealth = "Healthy" | "Degraded" | "Progressing" | "Suspended" | "Missing";

export type ApplicationStatus =
  | "Succeeded"
  | "Synced"
  | "OutOfSync"
  | "Unknown"
  | "Progressing"
  | "Degraded";

export interface ApplicationParams {
  id?: number;
  name?: string;
  project?: string;
  appType?: string;
  owner?: string;
  validated?: string;
}

export interface ApplicationResponse {
  success: boolean;
  total: number;
  items: ApplicationTemplate[];
}

export interface ValidatePayload {
  templateSource: string;
  targetRevision: string;
  path: string;
}

export interface CreateApplicationPayload {
  name: string;
  description?: string;
  source: {
    url: string;
    targetRevision: string;
    path: string;
    appType: "kustomize" | "helm" | "helm+kustomize";
  };
  destination: {
    clusters: string[];
  };
  tenant_name?: string;
  appcode?: string;
  security?: {
    external_secret?: {
      secret_store_ref: {
        id: string;
      };
    };
  };
  is_dryrun?: boolean;
}