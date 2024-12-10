export interface ClusterInfo {
  name: string;
  environment: string;
  provider: string;
  isAvailable?: boolean;
  version: {
    kubernetes: string;
    platform: string;
  };
  region: string;
  builtin?: boolean;
  nodes: {
    ready: number;
    total: number;
  };
  resourceQuota: {
    cpu: string;
    memory: string;
    storage: string;
    pvcs: string | number;
    nodeports: string | number;
  };
  quota?: {
    cpu: string;
    memory: string;
    storage: string;
    pods: string;
  };
  health: {
    status: string;
    message?: string;
  };
  monitoring?: {
    prometheus?: boolean;
    grafana?: boolean;
    alertmanager?: boolean;
    urls?: {
      prometheus?: string;
      grafana?: string;
      alertmanager?: string;
    };
  };
  consoleUrl: string;
  networkPolicy: boolean;
  ingressController: string;
  lastUpdated: string;
}

export interface ResourceQuota {
  cpu: string;
  memory: string;
  storage: string;
  pods: string;
}

export interface ClusterDefaults {
  [key: string]: ResourceQuota;
}

export type IconType = 'CheckCircle' | 'AlertTriangle' | 'XCircle' | 'Settings2';