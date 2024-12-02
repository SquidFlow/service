export interface ClusterInfo {
  id: number;
  name: string;
  env: string;
  status: 'active' | 'warning' | 'error';
  provider: string;
  version: {
    kubernetes: string;
    platform: string;
  };
  nodes: {
    ready: number;
    total: number;
  };
  resources: {
    cpu: string;
    memory: string;
    storage: string;
    pods: number;
  };
  monitoring?: {
    prometheus: boolean;
    grafana: boolean;
    alertmanager: boolean;
    urls?: {
      prometheus?: string;
      grafana?: string;
      alertmanager?: string;
    };
  };
  consoleUrl?: string;
  quota?: {
    cpu: string;
    memory: string;
    storage: string;
    pods: string;
  };
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