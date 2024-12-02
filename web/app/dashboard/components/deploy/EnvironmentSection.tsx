import { useState, useEffect } from 'react';
import { Card, CardHeader, CardContent, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Server } from "lucide-react";
import { useClusterStore } from '@/store';
import { ClusterInfo } from '@/types/cluster';
import { useDeployForm } from './DeployFormContext';

export function EnvironmentSection() {
  const { data: clusterList, getClusterList } = useClusterStore();
  const { selectedClusters, setSelectedClusters } = useDeployForm();

  useEffect(() => {
    getClusterList();
  }, [getClusterList]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-3">
          <Server className="h-6 w-6 text-blue-500" />
          <span>Destination Clusters</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
              Select Target Clusters
            </h3>
            <span className="text-sm text-gray-500">
              {selectedClusters.length} cluster(s) selected
            </span>
          </div>
          <div className="flex flex-wrap gap-2">
            {clusterList.map((cluster: ClusterInfo) => (
              <Button
                key={cluster.name}
                className="flex items-center gap-2 group relative"
                variant={selectedClusters.includes(cluster.name) ? "default" : "outline"}
                onClick={() => {
                  setSelectedClusters((prev: string[]) =>
                    prev.includes(cluster.name)
                      ? prev.filter((c) => c !== cluster.name)
                      : [...prev, cluster.name]
                  );
                }}
              >
                <span className="font-mono font-medium">{cluster.name}</span>
                <span className={`text-xs px-2 py-0.5 rounded-full ${
                  cluster.environment === "SIT"
                    ? "bg-blue-100 text-blue-700"
                    : cluster.environment === "UAT"
                      ? "bg-green-100 text-green-700"
                      : "bg-purple-100 text-purple-700"
                }`}>
                  {cluster.environment}
                </span>
              </Button>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}