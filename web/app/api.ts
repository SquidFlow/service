import useSWR from "swr";
import requestor from "@/requestor";
import { useEffect, useState } from "react";
import {
  ApplicationTemplate,
  ApplicationResponse,
  ValidatePayload,
  CreateApplicationPayload,
  CreateTemplatePayload,
  Kustomization,
  ClusterInfo,
  SecretStore,
  TenantInfo,
  TenantResponse,
  ClusterResponse,
  SecretStoreResponse,
} from "@/types";

const ARGOCDAPPLICATIONS = "/api/v1/deploy/argocdapplications";
const TEMPLATES = "/api/v1/applications/templates";
const TENANTS = "/api/v1/tenants";
const CLUSTER = "/api/v1/destinationCluster";
const SECRETSTORE = "/api/v1/security/externalsecrets/secretstore";
const APPCODE = "/api/v1/appcode";
const RELEASEHISTORIES = "/api/v1/releasehistories";

interface SimpleTenantInfo {
  id: string;
  name: string;
  secretPath: string;
}

export const useGetApplicationList = () => {
  const { data, error, mutate } = useSWR<ApplicationResponse>(
    ARGOCDAPPLICATIONS,
    async (url: string) => {
      const response = await requestor.get<ApplicationResponse>(url);
      return response.data;
    }
  );

  const applications = data?.items || [];
  return {
    applications,
    error,
    mutate,
  };
};

export const useGetApplicationDetail = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [applicationDetailData, setApplicationDetailData] = useState(null);

  const triggerGetApplicationDetail = async (applicationId: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await requestor.get(
        `${ARGOCDAPPLICATIONS}/${applicationId}`
      );
      setApplicationDetailData(response.data);
      return response.data;
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Unknown error"));
    } finally {
      setIsLoading(false);
    }
  };

  return {
    applicationDetailData,
    error,
    isLoading,
    triggerGetApplicationDetail,
  };
};

export const useKustomizationsData = () => {
  return { kustomizationsData: [] };
};

export const usePostValidate = () => {
  const [isValidating, setIsValidating] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState<unknown | null>(null);

  const triggerValidate = async (payload: ValidatePayload) => {
    setIsValidating(true);
    setError(null);
    try {
      const response = await requestor.post(
        `${ARGOCDAPPLICATIONS}/validate`,
        payload
      );
      setData(response.data);
      return response.data; // 返回接口数据
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Unknown error"));
    } finally {
      setIsValidating(false);
    }
  };

  return {
    data,
    error,
    isValidating,
    triggerValidate,
  };
};

export const useGetTemplateDetail = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState<unknown | null>(null);

  const triggerGetTemplateDetail = async (payload: { id: string }) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await requestor.get(`${TEMPLATES}/${payload.id}`);
      setData(response.data);
      return response.data.item; // 返回接口数据
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Unknown error"));
    } finally {
      setIsLoading(false);
    }
  };

  return {
    data,
    error,
    isLoading,
    triggerGetTemplateDetail,
  };
};

export const useGetAvailableTenants = () => {
  const { data, error } = useSWR(TENANTS, async (url) => {
    const response = await requestor.get(url);
    return response.data;
  });

  return {
    availableTenants: (data?.items || []) as SimpleTenantInfo[],
    error,
  };
};

export const useGetClusterList = () => {
  const { data, error, mutate } = useSWR<{ items: ClusterInfo[] }>(
    CLUSTER,
    async (url: string) => {
      const response = await requestor.get(url);
      return response.data;
    }
  );

  return {
    clusterList: data?.items || [],
    error,
    mutate,
  };
};

export const useGetSecretStore = () => {
  const { data, error } = useSWR(SECRETSTORE, async (url) => {
    const response = await requestor.get(url);
    return response.data;
  });

  return {
    secretStoreList: (data?.items || []) as SecretStore[],
    error,
  };
};

interface DryRunPayload {}

// 自定义钩子函数用于 dryrun 操作，接收 payload 作为参数
export const useDryRun = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState(null);

  const triggerDryRun = async (payload: DryRunPayload) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await requestor.post(
        `${ARGOCDAPPLICATIONS}/dryruntemplate`,
        payload
      );
      setData(response.data);
      return response.data; // 返回接口数据，方便外部使用
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Unknown error"));
    } finally {
      setIsLoading(false);
    }
  };

  return {
    data,
    error,
    isLoading,
    triggerDryRun,
  };
};

// useDeleteTemplate 自定义钩子函数，支持传入数组批量删除模板
export const useDeleteTemplate = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [successCount, setSuccessCount] = useState(0); // 用于记录成功删除的模板数量

  const deleteTemplates = async (templateIds: string[]) => {
    setIsLoading(true);
    setError(null);
    setSuccessCount(0); // 在每次发起批量删除请求前，重置成功删除数量为 0
    try {
      for (const templateId of templateIds) {
        await requestor.delete(`${TEMPLATES}/${templateId}`);
        setSuccessCount((prevCount) => prevCount + 1); // 每成功删除一个，成功数量加 1
      }
      return successCount; // 返回成功删除的模板数量
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Unknown error"));
    } finally {
      setIsLoading(false);
    }
  };

  return {
    isLoading,
    error,
    successCount,
    deleteTemplates,
  };
};

export const useGetAppCode = () => {
  return { appCodeData: [] };
};

export const usePostCreateApplication = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [createdApplicationData, setCreatedApplicationData] = useState<ApplicationTemplate | null>(null);

  const triggerPostCreateApplication = async (applicationData: CreateApplicationPayload) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await requestor.post(
        ARGOCDAPPLICATIONS,
        applicationData
      );
      setCreatedApplicationData(response.data);
      return response.data;
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Unknown error"));
    } finally {
      setIsLoading(false);
    }
  };

  return {
    createdApplicationData,
    error,
    isLoading,
    triggerPostCreateApplication,
  };
};

export const usePostCreateTemplate = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [createdTemplateData, setCreatedTemplateData] = useState<ApplicationTemplate | null>(null);

  const triggerPostCreateTemplate = async (templateData: CreateTemplatePayload) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await requestor.post(TEMPLATES, templateData);
      setCreatedTemplateData(response.data);
      return response.data;
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Unknown error"));
    } finally {
      setIsLoading(false);
    }
  };

  return {
    createdTemplateData,
    error,
    isLoading,
    triggerPostCreateTemplate,
  };
};

// Add new hook for deleting applications
export const useDeleteApplications = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const deleteApplications = async (appNames: string[]) => {
    setIsLoading(true);
    setError(null);
    try {
      const deletePromises = appNames.map(async (appName) => {
        const response = await requestor.delete(`${ARGOCDAPPLICATIONS}/${appName}`);
        return response.data;
      });

      const results = await Promise.all(deletePromises);
      return results;
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Unknown error"));
      throw err; // Re-throw the error so it can be caught by the component
    } finally {
      setIsLoading(false);
    }
  };

  return {
    isLoading,
    error,
    deleteApplications,
  };
};

export const useGetReleaseHistories = () => {
  const { data, error } = useSWR(RELEASEHISTORIES, async (url) => {
    const response = await requestor.get(url);
    return response.data;
  });

  return {
    releaseHistories: data || { SIT: [], UAT: [], PRD: [] },
    error,
  };
};
