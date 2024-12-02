import { Card, CardHeader, CardContent, CardTitle } from "@/components/ui/card";
import { Gauge } from "lucide-react";
import { ClusterInfo } from '@/types/cluster';

interface ResourceQuotaSectionProps {
  clusters: ClusterInfo[];
}

export function ResourceQuotaSection({ clusters }: ResourceQuotaSectionProps) {
  const parseResourceValue = (value: string) => {
    const match = value.match(/^(\d+(?:\.\d+)?)(Mi|Gi|m|cores)?$/);
    return match ? parseFloat(match[1]) : 0;
  };

  const getTotalResources = (clusters: ClusterInfo[]) => {
    return clusters.reduce(
      (acc, cluster) => {
        const quota = cluster.resourceQuota;
        return {
          cpu: acc.cpu + parseResourceValue(quota.cpu),
          memory: acc.memory + parseResourceValue(quota.memory),
          storage: acc.storage + parseResourceValue(quota.storage),
          pvcs: acc.pvcs + (typeof quota.pvcs === 'number' ? quota.pvcs : parseFloat(quota.pvcs || '0')),
          nodeports: acc.nodeports + (typeof quota.nodeports === 'number' ? quota.nodeports : parseFloat(quota.nodeports || '0'))
        };
      },
      { cpu: 0, memory: 0, storage: 0, pvcs: 0, nodeports: 0 }
    );
  };

  const formatResourceValue = (value: number, type: string) => {
    switch (type) {
      case 'cpu':
        return `${value} cores`;
      case 'memory':
        return `${value}Gi`;
      case 'storage':
        return `${value}Gi`;
      default:
        return value.toString();
    }
  };

  const totalResources = getTotalResources(clusters);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-3">
          <Gauge className="h-6 w-6 text-blue-500" />
          <span>Resource Quota</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
          <div className="space-y-2">
            <p className="text-sm font-medium text-gray-500 dark:text-gray-400">CPU</p>
            <p className="text-2xl font-bold">{formatResourceValue(totalResources.cpu, 'cpu')}</p>
          </div>
          <div className="space-y-2">
            <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Memory</p>
            <p className="text-2xl font-bold">{formatResourceValue(totalResources.memory, 'memory')}</p>
          </div>
          <div className="space-y-2">
            <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Storage</p>
            <p className="text-2xl font-bold">{formatResourceValue(totalResources.storage, 'storage')}</p>
          </div>
          <div className="space-y-2">
            <p className="text-sm font-medium text-gray-500 dark:text-gray-400">PVCs</p>
            <p className="text-2xl font-bold">{totalResources.pvcs}</p>
          </div>
          <div className="space-y-2">
            <p className="text-sm font-medium text-gray-500 dark:text-gray-400">NodePorts</p>
            <p className="text-2xl font-bold">{totalResources.nodeports}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}