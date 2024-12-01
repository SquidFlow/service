export interface ClusterInfo {
  id: number;
  name: string;
  env: string;
  status: 'active' | 'warning' | 'error';
  provider: 'GKE' | 'OCP' | 'AKS' | 'EKS';
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
  monitoring: {
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
  labels?: Record<string, string>;
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
  pvcs: string;
  nodeports: string;
}

export interface ClusterDefaults {
  [key: string]: ResourceQuota;
}

export type IconType = 'CheckCircle' | 'AlertTriangle' | 'XCircle' | 'Settings2' | 'ExternalLink' | 'Layout' | 'Server' | 'Cpu' | 'MemoryStick' | 'HardDrive' | 'Network' | 'Box';

export const resourceDescriptions = {
  cpu: {
    label: "CPU",
    tooltip: "Maximum CPU cores allocated in this cluster"
  },
  memory: {
    label: "Memory",
    tooltip: "Maximum RAM allocated in this cluster"
  },
  storage: {
    label: "Storage",
    tooltip: "Maximum storage space for persistent volumes"
  },
  pvcs: {
    label: "PVCs",
    tooltip: "Maximum number of persistent volumes allowed"
  },
  nodeports: {
    label: "NodePorts",
    tooltip: "Maximum number of NodePort services allowed"
  }
} as const;

export const providerColorMap = {
  GKE: 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400',
  OCP: 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400',
  AKS: 'bg-purple-100 text-purple-800 dark:bg-purple-900/20 dark:text-purple-400',
  EKS: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400',
  default: 'bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-400'
} as const;

export const healthStatusMap: {
  [key: string]: {
    bg: string;
    text: string;
    icon: IconType | null;
  };
} = {
  Healthy: {
    bg: 'bg-green-100 dark:bg-green-900/30',
    text: 'text-green-800 dark:text-green-400',
    icon: 'CheckCircle'
  },
  Warning: {
    bg: 'bg-yellow-100 dark:bg-yellow-900/30',
    text: 'text-yellow-800 dark:text-yellow-400',
    icon: 'AlertTriangle'
  },
  Degraded: {
    bg: 'bg-red-100 dark:bg-red-900/30',
    text: 'text-red-800 dark:text-red-400',
    icon: 'XCircle'
  },
  default: {
    bg: 'bg-gray-100 dark:bg-gray-900/30',
    text: 'text-gray-800 dark:text-gray-400',
    icon: null
  }
} as const;

export const monitoringTypeStyles = {
  prometheus: 'bg-blue-50 text-blue-700 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400',
  grafana: 'bg-orange-50 text-orange-700 hover:bg-orange-100 dark:bg-orange-900/20 dark:text-orange-400',
  alertmanager: 'bg-red-50 text-red-700 hover:bg-red-100 dark:bg-red-900/20 dark:text-red-400'
} as const;

export interface ClusterResponse {
  success: boolean;
  total: number;
  items: ClusterInfo[];
}