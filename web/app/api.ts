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
export const useApplications = (params: any) => {
  const { data, error } = useSWR(ARGOCDAPPLICATIONS, async (url) => {
    const response = await requestor.get(url, { params });
    return response.data;
  });

  const applications :any[]= [];
  return {
    applications,
    error,
  };
};

export const useKustomizationsData = () => {
  const { data, error } = useSWR(TEMPLATES, async (url) => {
    const response = await requestor.get(url);
    return response.data;
  });

  const kustomizationsData = data || [];
  return {
    kustomizationsData,
    error,
  };
};

export const usePostValidate = (payload: any) => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [data, setData] = useState(null);

  const triggerValidate = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await requestor.post(`${TEMPLATES}/validate`, payload);
      setData(response.data);
    } catch (err:any) {
      setError(err);
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
  const { data, error } = useSWR(TENANTS, async (url) => {
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
