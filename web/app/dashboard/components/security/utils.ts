import type { SecretStore } from '@/types';

export const getSecretStoreYAML = (store: SecretStore) => {
  const getProviderConfig = () => {
    switch (store.provider) {
      case 'aws':
        return `
      aws:
        region: ap-southeast-1
        auth:
          secretRef:
            accessKeyIDSecretRef:
              name: aws-secret
              key: access-key
            secretAccessKeySecretRef:
              name: aws-secret
              key: secret-key`;
      case 'vault':
        return `
      vault:
        server: "https://vault.example.com"
        path: ${store.path}
        version: "v2"
        auth:
          tokenSecretRef:
            name: vault-token
            key: token`;
      case 'azure':
        return `
      azurekv:
        authType: WorkloadIdentity
        vaultUrl: "https://${store.path}.vault.azure.net"
        serviceAccountRef:
          name: workload-identity-sa`;
      default:
        return '';
    }
  };

  return `apiVersion: external-secrets.io/v1beta1
kind: ${store.type}
metadata:
  name: ${store.name}
spec:
  provider:${getProviderConfig()}`;
};

export const getProviderBadgeColor = (provider: string) => {
  const colors: Record<string, string> = {
    aws: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
    vault: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
    azure: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400'
  };
  return colors[provider] || 'bg-gray-100 text-gray-800';
};

export const getTypeBadgeColor = (type: string) => {
  return type === 'ClusterSecretStore'
    ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
    : 'bg-gray-100 text-gray-800 dark:bg-gray-800/30 dark:text-gray-400';
};

export const getHealthBadgeColor = (status: string) => {
  switch (status.toLowerCase()) {
    case 'healthy':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
    case 'warning':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
    case 'error':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400';
  }
};