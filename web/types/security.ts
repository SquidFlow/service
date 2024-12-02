export type SecretStoreType = 'SecretStore' | 'ClusterSecretStore';
export type SecretStoreProvider = 'vault' | 'aws' | 'azure';
export type SecretStoreStatus = 'Active' | 'Inactive';

export interface SecretStore {
  id: string;
  name: string;
  provider: SecretStoreProvider;
  type: SecretStoreType;
  status: SecretStoreStatus;
  environment: string[];
  path: string;
  lastSynced: string;
  createdAt: string;
  lastUpdated: string;
  health: {
    status: 'Healthy' | 'Warning' | 'Error';
    message: string;
  };
}

export interface SecretStoreResponse {
  success: boolean;
  total: number;
  items: SecretStore[];
  message: string;
}