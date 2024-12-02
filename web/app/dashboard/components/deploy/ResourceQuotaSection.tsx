import { useEffect } from "react";
import { Card, CardHeader, CardContent, CardTitle } from "@/components/ui/card";
import { BarChart } from "lucide-react";
import type { ClusterInfo } from '@/types';
import { useClusterStore } from '@/store';

export function ResourceQuotaSection() {
  const { data: clusters, getClusterList } = useClusterStore();

  useEffect(() => {
    getClusterList();
  }, [getClusterList]);

  const getTotalResources = () => {
    return clusters.reduce(
      (acc, cluster) => {
        const resources = cluster.resources;
        return {
          cpu: acc.cpu + parseFloat(resources.cpu || '0'),
          memory: acc.memory + parseFloat(resources.memory || '0'),
          storage: acc.storage + parseFloat(resources.storage || '0'),
          pods: acc.pods + (resources.pods || 0),
        };
      },
      { cpu: 0, memory: 0, storage: 0, pods: 0 }
    );
  };

  const totalResources = getTotalResources();

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-3">
          <BarChart className="h-6 w-6 text-blue-500" />
          <span>Resource Quota Overview</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
              Total CPU Cores
            </h3>
            <p className="mt-2 text-3xl font-bold">
              {totalResources.cpu.toFixed(1)}
            </p>
          </div>
          <div>
            <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
              Total Memory (GiB)
            </h3>
            <p className="mt-2 text-3xl font-bold">
              {totalResources.memory.toFixed(1)}
            </p>
          </div>
          <div>
            <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
              Total Storage (GiB)
            </h3>
            <p className="mt-2 text-3xl font-bold">
              {totalResources.storage.toFixed(1)}
            </p>
          </div>
          <div>
            <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
              Total Pods
            </h3>
            <p className="mt-2 text-3xl font-bold">
              {totalResources.pods}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}