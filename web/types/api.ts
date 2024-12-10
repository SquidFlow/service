import type { ApplicationTemplate } from './application';
import type { ClusterInfo } from './cluster';
import type { SecretStore } from './security';
import type { TenantInfo } from './tenant';

export interface BaseResponse<T> {
  success: boolean;
  total: number;
  error: string;
  items: T[];
}

export interface ApplicationResponse extends BaseResponse<ApplicationTemplate> {}
export interface ClusterResponse extends BaseResponse<ClusterInfo> {}
export interface SecretStoreResponse extends BaseResponse<SecretStore> {}
export interface TenantResponse extends BaseResponse<TenantInfo> {}