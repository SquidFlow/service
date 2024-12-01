export interface SecretStore {
  id: string;
  name: string;
  provider: 'AWS' | 'Vault' | 'GCP';
  type: 'SecretStore' | 'ClusterSecretStore';
  path: string;
  health: {
    status: 'Healthy' | 'Warning' | 'Error';
    message?: string;
  };
}