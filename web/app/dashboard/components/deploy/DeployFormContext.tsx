import { createContext, useContext, useState, ReactNode } from 'react';
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
}

const DeployFormContext = createContext<DeployFormState | undefined>(undefined);

export function DeployFormProvider({ children }: { children: ReactNode }) {
  const [source, setSource] = useState<ApplicationSource>({
    url: "",
    path: "",
    targetRevision: "",
    name: "",
  });
  const [selectedClusters, setSelectedClusters] = useState<string[]>([]);
  const [clusterDetails, setClusterDetails] = useState<ClusterInfo[]>([]);

  return (
    <DeployFormContext.Provider
      value={{
        source,
        setSource,
        selectedClusters,
        setSelectedClusters,
        clusterDetails,
        setClusterDetails,
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