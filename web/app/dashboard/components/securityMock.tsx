"use client"

export interface SecretStore {
  id: number;
  name: string;
  provider: 'AWS' | 'GCP' | 'Azure' | 'Vault' | 'CyberArk';
  type: string;
  status: 'Active' | 'Inactive' | 'Error';
  path?: string;
  lastSynced: string;
  createdAt: string;
  lastUpdated: string;
  health: {
    status: 'Healthy' | 'Warning' | 'Error';
    message?: string;
  };
}

export const secretStoreMockData: SecretStore[] = [
  {
    id: 1,
    name: 'aws-secrets-manager',
    provider: 'AWS',
    type: 'SecretStore',
    status: 'Active',
    path: 'aws/data/applications/*',
    lastSynced: '2024-03-15T08:30:00Z',
    createdAt: '2024-01-15',
    lastUpdated: '2024-03-15',
    health: {
      status: 'Healthy',
      message: 'Connected to AWS Secrets Manager'
    }
  },
  {
    id: 2,
    name: 'vault-kv',
    provider: 'Vault',
    type: 'ClusterSecretStore',
    status: 'Active',
    path: 'secret/data/applications',
    lastSynced: '2024-03-15T08:25:00Z',
    createdAt: '2024-01-20',
    lastUpdated: '2024-03-15',
    health: {
      status: 'Healthy',
      message: 'Connected to Vault server'
    }
  },
  {
    id: 3,
    name: 'azure-key-vault',
    provider: 'Azure',
    type: 'SecretStore',
    status: 'Active',
    path: 'azure/data/certificates',
    lastSynced: '2024-03-15T08:20:00Z',
    createdAt: '2024-02-01',
    lastUpdated: '2024-03-15',
    health: {
      status: 'Warning',
      message: 'High latency detected'
    }
  },
  {
    id: 4,
    name: 'gcp-secret-manager',
    provider: 'GCP',
    type: 'ClusterSecretStore',
    status: 'Active',
    path: 'gcp/data/projects/*/secrets',
    lastSynced: '2024-03-15T08:15:00Z',
    createdAt: '2024-02-15',
    lastUpdated: '2024-03-15',
    health: {
      status: 'Healthy',
      message: 'Connected to GCP Secret Manager'
    }
  },
  {
    id: 5,
    name: 'cyberark-conjur',
    provider: 'CyberArk',
    type: 'SecretStore',
    status: 'Error',
    path: 'cyberark/data/apps/credentials',
    lastSynced: '2024-03-15T07:00:00Z',
    createdAt: '2024-03-01',
    lastUpdated: '2024-03-15',
    health: {
      status: 'Error',
      message: 'Authentication failed'
    }
  }
];

