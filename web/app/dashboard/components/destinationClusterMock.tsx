// this data should be fetched from the server
export interface ClusterInfo {
  name: string;
  environment: string;
  status: 'active' | 'disabled';
  provider: 'GKE' | 'OCP' | 'AKS' | 'EKS';
  version: {
    kubernetes: string;
    platform: string;
  };
  nodeCount: number;
  region: string;
  resourceQuota: ResourceQuota;
  health: {
    status: 'Healthy' | 'Degraded' | 'Warning';
    message?: string;
  };
  nodes: {
    ready: number;
    total: number;
  };
  networkPolicy: boolean;
  ingressController: string;
  lastUpdated: string;
  consoleUrl?: string;
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
  builtin?: boolean;
}

export interface ResourceQuota {
  cpu: string;
  memory: string;
  storage: string;
  pvcs: string;
  nodeports: string;
}

// 添加资源描述
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
};

// 添加供应商颜色映射
export const providerColorMap = {
  GKE: 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400',
  OCP: 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400',
  AKS: 'bg-purple-100 text-purple-800 dark:bg-purple-900/20 dark:text-purple-400',
  EKS: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400',
  default: 'bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-400'
};

// 添加图标类型
export type IconType = 'CheckCircle' | 'AlertTriangle' | 'XCircle' | 'Settings2' | 'ExternalLink' | 'Layout' | 'Server' | 'Cpu' | 'MemoryStick' | 'HardDrive' | 'Network' | 'Box';

// 更新健康状态样式映射的类型
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
};

// 添加监控工具样式映射
export const monitoringTypeStyles = {
  prometheus: 'bg-blue-50 text-blue-700 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400',
  grafana: 'bg-orange-50 text-orange-700 hover:bg-orange-100 dark:bg-orange-900/20 dark:text-orange-400',
  alertmanager: 'bg-red-50 text-red-700 hover:bg-red-100 dark:bg-red-900/20 dark:text-red-400'
};

// 更新集群数据
export const clusters: ClusterInfo[] = [
  {
    name: 'SIT0',
    environment: 'SIT',
    status: 'active',
    provider: 'GKE',
    version: {
      kubernetes: 'v1.24.0',
      platform: 'GKE 1.24.0-gke.1000',
    },
    nodeCount: 3,
    region: 'ap-southeast-1',
    resourceQuota: {
      cpu: '8 cores',
      memory: '32Gi',
      storage: '500Gi',
      pvcs: '10',
      nodeports: '5'
    },
    health: {
      status: 'Healthy',
      message: 'All core components are healthy'
    },
    nodes: {
      ready: 3,
      total: 3
    },
    networkPolicy: true,
    ingressController: 'GCE',
    lastUpdated: '2024-03-15T10:00:00Z',
    consoleUrl: 'https://console.cloud.google.com/kubernetes/clusters/details/asia-southeast1/sit0',
    monitoring: {
      prometheus: true,
      grafana: true,
      alertmanager: true,
      urls: {
        prometheus: 'https://prometheus.sit0.example.com',
        grafana: 'https://grafana.sit0.example.com',
        alertmanager: 'https://alertmanager.sit0.example.com'
      }
    }
  },
  {
    name: 'SIT1',
    environment: 'SIT',
    status: 'active',
    provider: 'OCP',
    version: {
      kubernetes: 'v1.24.0',
      platform: 'OpenShift 4.12',
    },
    builtin: true,
    nodeCount: 3,
    region: 'SG IDC-A',
    resourceQuota: {
      cpu: '8 cores',
      memory: '32Gi',
      storage: '500Gi',
      pvcs: '10',
      nodeports: '5'
    },
    health: {
      status: 'Healthy',
      message: 'All components are healthy'
    },
    nodes: {
      ready: 3,
      total: 3
    },
    networkPolicy: true,
    ingressController: 'OpenShift Router',
    lastUpdated: '2024-03-15T09:00:00Z',
    consoleUrl: 'https://console.openshift.com/clusters/sit1',
    monitoring: {
      prometheus: true,
      grafana: true,
      alertmanager: true,
      urls: {
        prometheus: 'https://prometheus.sit1.example.com',
        grafana: 'https://grafana.sit1.example.com',
        alertmanager: 'https://alertmanager.sit1.example.com'
      }
    }
  },
  {
    name: 'UAT',
    environment: 'UAT',
    status: 'active',
    provider: 'GKE',
    version: {
      kubernetes: 'v1.25.0',
      platform: 'GKE 1.25.0-gke.1000',
    },
    nodeCount: 5,
    region: 'ap-southeast-1',
    resourceQuota: {
      cpu: '16 cores',
      memory: '64Gi',
      storage: '1000Gi',
      pvcs: '10',
      nodeports: '5'
    },
    health: {
      status: 'Healthy',
      message: 'All components are healthy'
    },
    nodes: {
      ready: 5,
      total: 5
    },
    networkPolicy: true,
    ingressController: 'GCE',
    lastUpdated: '2024-03-15T08:00:00Z',
    consoleUrl: 'https://console.cloud.google.com/kubernetes/clusters/details/asia-southeast1/uat',
    monitoring: {
      prometheus: true,
      grafana: true,
      alertmanager: true,
      urls: {
        prometheus: 'https://prometheus.uat.example.com',
        grafana: 'https://grafana.uat.example.com',
        alertmanager: 'https://alertmanager.uat.example.com'
      }
    }
  },
  {
    name: 'PDC',
    environment: 'PRD',
    status: 'active',
    provider: 'OCP',
    version: {
      kubernetes: 'v1.25.0',
      platform: 'OpenShift 4.13',
    },
    nodeCount: 12,
    region: 'SG IDC-A',
    resourceQuota: {
      cpu: '64 cores',
      memory: '256Gi',
      storage: '5000Gi',
      pvcs: '10',
      nodeports: '5'
    },
    health: {
      status: 'Healthy',
      message: 'All components are healthy'
    },
    nodes: {
      ready: 12,
      total: 12
    },
    networkPolicy: true,
    ingressController: 'OpenShift Router',
    lastUpdated: '2024-03-15T08:00:00Z',
    consoleUrl: 'https://console.openshift.com/clusters/pdc',
    monitoring: {
      prometheus: true,
      grafana: true,
      alertmanager: true,
      urls: {
        prometheus: 'https://prometheus.pdc.example.com',
        grafana: 'https://grafana.pdc.example.com',
        alertmanager: 'https://alertmanager.pdc.example.com'
      }
    }
  },
  {
    name: 'DDC',
    environment: 'PRD',
    status: 'active',
    provider: 'OCP',
    version: {
      kubernetes: 'v1.25.0',
      platform: 'OpenShift 4.13',
    },
    nodeCount: 12,
    region: 'SG IDC-B',
    resourceQuota: {
      cpu: '64 cores',
      memory: '256Gi',
      storage: '5000Gi',
      pvcs: '10',
      nodeports: '5'
    },
    health: {
      status: 'Healthy',
      message: 'All components are healthy'
    },
    nodes: {
      ready: 12,
      total: 12
    },
    networkPolicy: true,
    ingressController: 'OpenShift Router',
    lastUpdated: '2024-03-15T08:00:00Z',
    consoleUrl: 'https://console.openshift.com/clusters/ddc',
    monitoring: {
      prometheus: true,
      grafana: true,
      alertmanager: true,
      urls: {
        prometheus: 'https://prometheus.ddc.example.com',
        grafana: 'https://grafana.ddc.example.com',
        alertmanager: 'https://alertmanager.ddc.example.com'
      }
    }
  }
];
