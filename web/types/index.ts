import { SecretStore } from './security';

export interface CreateTemplatePayload {
  name: string;
  description: string;
  source: {
    type: string;
    url: string;
    targetRevision: string;
  };
  path: string;
  owner: string;
  appType: string;
}

export interface SecretStoreResponse {
  success: boolean;
  total: number;
  items: SecretStore[];
}

export * from './application';
export * from './template';
export * from './cluster';
export * from './security';
export * from './release';
export * from './tenant';