import { Card, CardHeader, CardContent, CardTitle } from "@/components/ui/card";
import { BarChart } from "lucide-react";
import { useGetClusterList } from '@/app/api';
import { ClusterInfo } from '@/types/cluster';

export function ResourceQuotaSection() {
  const { clusterList } = useGetClusterList();

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-3">
          <BarChart className="h-6 w-6 text-green-500" />
          <span>Resource Quotas</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-3 gap-6">
          {clusterList.map((cluster: ClusterInfo) => (
            <Card key={cluster.name} className="bg-gray-50 dark:bg-gray-800">
              <CardHeader>
                <CardTitle className="text-base">{cluster.name}</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-gray-500">CPU Usage</span>
                    <span className="font-medium">{cluster.resources?.cpu ?? '0'}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-gray-500">Memory Usage</span>
                    <span className="font-medium">{cluster.resources?.memory ?? '0'}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-gray-500">Storage Usage</span>
                    <span className="font-medium">{cluster.resources?.storage ?? '0'}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}