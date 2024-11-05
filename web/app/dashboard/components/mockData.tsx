export interface Ingress {
  name: string;
  service: string;
  port: string;
}

export interface Repository {
  url: string;
  branch: string;
}

export interface TemplateSource {
  type: 'builtin' | 'external';
  value: string;
  instanceName?: string;
  targetRevision?: string;
}

export const fieldDescriptions = {
  tenantName: {
    label: "Tenant Name",
    tooltip: "The unique identifier for your organization or project"
  },
  appCode: {
    label: "App Code",
    tooltip: "Application identifier code used for resource management and tracking"
  },
  namespace: {
    label: "Namespace",
    tooltip: "Kubernetes namespace where the application will be deployed"
  },
  description: {
    label: "Description",
    tooltip: "Detailed information about the application and its purpose"
  },
  ingress: {
    label: "Ingress",
    tooltip: "Configure external access to services in your Kubernetes cluster",
    name: "Unique name for the ingress rule",
    service: "Target service to route traffic to",
    port: "Port number of the service"
  },
  builtinTemplate: {
    label: "Built-in Template",
    tooltip: "Pre-configured deployment templates maintained by the platform team"
  },
  externalTemplate: {
    label: "External Template",
    tooltip: "Custom deployment template from an external Git repository or URL"
  }
};

export interface ClusterQuota {
  cpu: string;
  memory: string;
  storage: string;
  pvcs: string;
  nodeports: string;
}

export interface ClusterDefaults {
  [key: string]: ClusterQuota;
}

export const clusterDefaults: ClusterDefaults = {
  SIT: {
    cpu: "2",
    memory: "4",
    storage: "20",
    pvcs: "3",
    nodeports: "2"
  },
  SIT1: {
    cpu: "2",
    memory: "4",
    storage: "20",
    pvcs: "3",
    nodeports: "2"
  },
  UAT: {
    cpu: "4",
    memory: "8",
    storage: "50",
    pvcs: "5",
    nodeports: "3"
  },
  PRD: {
    cpu: "8",
    memory: "16",
    storage: "100",
    pvcs: "10",
    nodeports: "5"
  }
};

export const mockYamlTemplate = (namespace: string | undefined, selectedCluster: string, clusterDefaults: ClusterDefaults) => {
  const defaultQuota = clusterDefaults['SIT'];
  const clusterQuota = clusterDefaults[selectedCluster] || defaultQuota;

  return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
  namespace: ${namespace || 'default'}
  labels:
    app: example
    environment: ${selectedCluster || 'dev'}
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
    spec:
      containers:
      - name: example-app
        image: nginx:1.14.2
        imagePullPolicy: Always
        ports:
        - name: http
          containerPort: 80
          protocol: TCP
        - name: metrics
          containerPort: 8080
          protocol: TCP
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        resources:
          limits:
            cpu: "${clusterQuota.cpu}m"
            memory: "${clusterQuota.memory}Gi"
          requests:
            cpu: "500m"
            memory: "1Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 5
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
        - name: secret-volume
          mountPath: /etc/secrets
          readOnly: true
      volumes:
      - name: config-volume
        configMap:
          name: example-config
      - name: secret-volume
        secret:
          secretName: example-secret
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
      serviceAccountName: example-sa`;
}

export interface TenantInfo {
  id: string;
  name: string;
  secretPath: string;
}

export const tenants: TenantInfo[] = [
  { id: '1', name: 'Tenant A', secretPath: 'secrets/data/tenant-a' },
  { id: '2', name: 'Tenant B', secretPath: 'secrets/data/tenant-b' },
  { id: '3', name: 'Tenant C', secretPath: 'secrets/data/tenant-c' },
];
