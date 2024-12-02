export interface Kustomization {
  id: string;
  name: string;
  path: string;
  validated: boolean;
  owner: string;
  environments: string[];
  lastApplied: string;
  appType: "kustomization";
  source: {
    type: "git";
    url: string;
    branch?: string;
    tag?: string;
    commit?: string;
  };
  description?: string;
  resources: {
    deployments: number;
    services: number;
    configmaps: number;
    secrets: number;
    ingresses: number;
  };
}

export interface Repository {
  url: string;
  branch: string;
}