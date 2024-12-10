import { createContext, useContext, useState, ReactNode, useEffect } from 'react';
import { ClusterInfo } from '@/types/cluster';

interface IngressRule {
  name: string;
  service: string;
  port: string;
}

interface ApplicationSource {
  url: string;
  path: string;
  targetRevision: string;
  name: string;
  description?: string;
  tenant?: string;
  appCode?: string;
  namespace?: string;
  application_specifier?: {
    helm_manifest_path?: string;
  };
  ingress?: IngressRule[];
  externalSecrets?: {
    enabled: boolean;
    secretStore?: string;
  };
}

interface DeployFormState {
  source: ApplicationSource;
  setSource: (source: ApplicationSource | ((prev: ApplicationSource) => ApplicationSource)) => void;
  selectedClusters: string[];
  setSelectedClusters: (clusters: string[] | ((prev: string[]) => string[])) => void;
  clusterDetails: ClusterInfo[];
  setClusterDetails: (clusters: ClusterInfo[]) => void;
  clearSavedData: () => void;
}

const DeployFormContext = createContext<DeployFormState | undefined>(undefined);

const FORM_STORAGE_KEY = 'deploy-form-draft';

export function DeployFormProvider({ children }: { children: ReactNode }) {
  const [source, setSource] = useState<ApplicationSource>(() => {
    // 尝试从 localStorage 恢复数据
    const saved = localStorage.getItem(FORM_STORAGE_KEY);
    if (saved) {
      try {
        return JSON.parse(saved);
      } catch (e) {
        console.error('Failed to parse saved form data:', e);
      }
    }
    return {
      url: "",
      path: "",
      targetRevision: "",
      name: "",
    };
  });

  const [selectedClusters, setSelectedClusters] = useState<string[]>([]);
  const [clusterDetails, setClusterDetails] = useState<ClusterInfo[]>([]);

  // 自动保存表单数据
  useEffect(() => {
    localStorage.setItem(FORM_STORAGE_KEY, JSON.stringify(source));
  }, [source]);

  // 清除保存的数据
  const clearSavedData = () => {
    localStorage.removeItem(FORM_STORAGE_KEY);
    setSource({
      url: "",
      path: "",
      targetRevision: "",
      name: "",
    });
  };

  return (
    <DeployFormContext.Provider
      value={{
        source,
        setSource,
        selectedClusters,
        setSelectedClusters,
        clusterDetails,
        setClusterDetails,
        clearSavedData, // 导出清除方法
      }}
    >
      {children}
    </DeployFormContext.Provider>
  );
}

export function useDeployForm() {
  const context = useContext(DeployFormContext);
  if (context === undefined) {
    throw new Error('useDeployForm must be used within a DeployFormProvider');
  }
  return context;
}