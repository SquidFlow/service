export interface TenantInfo {
  id: string;
  name: string;
  description: string;
  owner: string;
  status: string;
  createdAt: string;
  updatedAt: string;
  secretPath: string;
}

export interface SimpleTenantInfo {
  name: string;
  id: string;
  type: string;
  createdAt: string;
}

export interface TenantResponse {
  success: boolean;
  total: number;
  items: TenantInfo[];
}

export interface AppCodeResponse {
  appCodes: string[];
}