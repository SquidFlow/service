export interface EnvironmentInfo {
  environment: string;
  isValid: boolean;
}

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
    serviceAccounts: number;
    roles: number;
    roleBindings: number;
    networkPolicies: number;
    persistentVolumeClaims: number;
    horizontalPodAutoscalers: number;
    customResourceDefinitions: {
      externalSecrets: number;
      certificates: number;
      ingressRoutes: number;
      prometheusRules: number;
      serviceMeshPolicies: number;
      virtualServices: number;
    };
  };
  events: {
    time: string;
    type: "Normal" | "Warning";
    reason: string;
    message: string;
  }[];
}

export interface TemplateSource {
  type: "git";
  value: string;
  instanceName?: string;
  targetRevision?: string;
  path?: string;
}

export interface Repository {
  url: string;
  branch: string;
}

export interface Ingress {
  name: string;
  service: string;
  port: string;
}

export const fieldDescriptions = {
  tenantName: {
    label: "Tenant Name",
    tooltip: "The unique identifier for your organization or project",
  },
  appCode: {
    label: "App Code",
    tooltip: "Application identifier code used for resource management and tracking",
  },
  namespace: {
    label: "Namespace",
    tooltip: "Kubernetes namespace where the application will be deployed",
  },
  description: {
    label: "Description",
    tooltip: "Detailed information about the application and its purpose",
  },
  ingress: {
    label: "Ingress",
    tooltip: "Configure external access to services in your Kubernetes cluster",
    name: "Unique name for the ingress rule",
    service: "Target service to route traffic to",
    port: "Port number of the service",
  },
  gitRepository: {
    label: "Git Repository",
    tooltip: "Git repository containing your application deployment configuration",
  },
} as const;