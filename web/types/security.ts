export interface SecretStore {
  id: string;
  name: string;
  provider: string;
  type: string;
  path: string;
  health: {
    status: 'Healthy' | 'Warning' | 'Error';
    message?: string;
  };
}