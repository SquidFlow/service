import { SecretStore } from '@/types/security';

export const getSecretStoreYAML = (store: SecretStore) => {
  const isCluster = store.type === 'ClusterSecretStore';
  return `apiVersion: external-secrets.io/v1beta1
kind: ${store.type}
metadata:
  name: ${store.name}
${!isCluster ? '  namespace: default\n' : ''}spec:
  provider:
    ${store.provider.toLowerCase()}:
      server: "https://vault.your-domain.com"
      path: ${store.path}
      version: v2
      auth:
        tokenSecretRef:
          name: "${store.provider.toLowerCase()}-token${isCluster ? '-global' : ''}"
          key: "token"
          ${isCluster ? 'namespace: external-secrets' : ''}`;
};

export const getProviderBadgeColor = (provider: string) => {
  const colors: Record<string, string> = {
    AWS: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
    GCP: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
    Azure: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
    Vault: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400',
    CyberArk: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
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