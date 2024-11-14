import useSWR from 'swr';
import requestor from '@/requestor';
import { useState } from 'react';
import { SecretStore } from './dashboard/components/securityMock';
import { ClusterInfo } from './dashboard/components/destinationClusterMock';
import { TenantInfo } from './dashboard/components/mockData';

const ARGOCDAPPLICATIONS = '/api/v1/deploy/argocdapplications';
const TEMPLATES = '/api/v1/applications/templates';
const TENANTS = '/api/v1/tenants';
const CLUSTER = '/api/v1/destinationCluster';
const SECRETSTORE = '/api/v1/security/externalsecrets/secretstore';

export interface ApplicationTemplate {
  id: number;
  name: string;
  owner: string;
  description?: string;
  path: string;
  environments: string[];
  appType: 'kustomize' | 'helm' | 'helm+kustomize';
  source: {
    url: string;
    targetRevision: string;
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
  status: 'Synced' | 'OutOfSync' | 'Unknown' | 'Progressing' | 'Degraded';
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
  deployedEnvironments: string[];
  health: {
    status: 'Healthy' | 'Degraded' | 'Progressing' | 'Suspended' | 'Missing';
    message?: string;
  };
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

interface ApplicationParams {
  id?: number;
  name?: string;
  project?: string;
  appType?: string;
  owner?: string;
  validated?: string;
}

interface ApplicationResponse {
  success: boolean;
  total: number;
  items: ApplicationTemplate[];
}

// Add new interface for validate payload
interface ValidatePayload {
  name: string;
  path: string;
  owner: string;
  source: {
    url: string;
    targetRevision: string;
    type: 'git';
  };
  appType: 'kustomize' | 'helm' | 'helm+kustomize';
  description?: string;
}

export const useApplications = (params: ApplicationParams) => {
  const { data, error } = useSWR<ApplicationResponse>(ARGOCDAPPLICATIONS, async (url: string) => {
    const response = await requestor.get<ApplicationResponse>(url, { params });
    return response.data;
  });

  const applications = data?.items || [];
  return {
    applications,
    error,
  };
};

export const useKustomizationsData = () => {
  const { data, error } = useSWR(TEMPLATES, async (url: string) => {
    const response = await requestor.get(url);
    return response.data;
  });

  const kustomizationsData = data || [];
  return {
    kustomizationsData,
    error,
  };
};

export const usePostValidate = (payload: ValidatePayload) => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState<unknown | null>(null);

  const triggerValidate = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await requestor.post(`${TEMPLATES}/validate`, payload);
      setData(response.data);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Unknown error'));
    } finally {
      setIsLoading(false);
    }
  };

  return {
    data,
    error,
    isLoading,
    triggerValidate,
  };
};

export const useGetAvailableTenants = () => {
  const { data, error } = useSWR(TENANTS, async (url: string) => {
    const response = await requestor.get(url);
    return response.data;
  });

  const availableTenants: TenantInfo[] = data?.projects || [];
  return {
    availableTenants,
    error,
  };
};

export const useGetClusterList = () => {
  const { data, error } = useSWR(CLUSTER, async (url) => {
    const response = await requestor.get(url);
    return response.data;
  });

  const clusterList: ClusterInfo[] = data || [];
  return {
    clusterList,
    error,
  };
};

export const useGetSecretStore = () => {
  const { data, error } = useSWR(SECRETSTORE, async (url) => {
    const response = await requestor.get(url);
    return response.data;
  });

  const secretStoreList: SecretStore[] = data || [];
  return {
    secretStoreList,
    error,
  };
};
