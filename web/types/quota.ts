export interface UIResourceQuota {
  cpu: string;
  memory: string;
  storage: string;
  pods: string;
}

export interface ClusterResourceQuota extends UIResourceQuota {
  pvcs: string;
  nodeports: string;
}

export interface QuotaField {
  name: keyof UIResourceQuota;
  label: string;
  unit: string;
  tooltip: string;
}

export const quotaFields: QuotaField[] = [
  {
    name: "cpu",
    label: "CPU Limit",
    unit: "cores",
    tooltip: "Maximum CPU cores that can be allocated",
  },
  {
    name: "memory",
    label: "Memory Limit",
    unit: "GiB",
    tooltip: "Maximum memory that can be allocated",
  },
  {
    name: "storage",
    label: "Storage Limit",
    unit: "GiB",
    tooltip: "Maximum storage space that can be allocated",
  },
  {
    name: "pods",
    label: "Pod Limit",
    unit: "pods",
    tooltip: "Maximum number of pods that can be created",
  },
];

export const toClusterQuota = (quota: UIResourceQuota): ClusterResourceQuota => ({
  ...quota,
  pvcs: "50",     // 默认值
  nodeports: "20", // 默认值
});