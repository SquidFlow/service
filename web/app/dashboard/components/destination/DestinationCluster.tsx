import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { StatCards } from './StatCards';
import { ClusterList } from './ClusterList';
import { ResourceQuotaDialog, ResourceQuota } from './ResourceQuotaDialog';
import { useGetClusterList } from '@/app/api';
import { ClusterInfo } from '@/types/cluster';

export function DestinationCluster() {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [selectedCluster, setSelectedCluster] = useState<ClusterInfo | undefined>();
  const { clusterList, error, mutate } = useGetClusterList();

  const handleResourceQuota = (cluster: ClusterInfo) => {
    setSelectedCluster(cluster);
    setIsDialogOpen(true);
  };

  const handleSaveQuota = async (quotas: ResourceQuota) => {
    try {
      // await updateClusterQuota(selectedCluster?.name, quotas);
      await mutate();
    } catch (error) {
      console.error('Failed to update resource quota:', error);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">Target Clusters</h1>
        <Button className="flex items-center space-x-2" variant="default">
          <Plus className="h-4 w-4" />
          <span>Add Cluster</span>
        </Button>
      </div>

      <p className="text-muted-foreground">
        Manage and monitor your Kubernetes clusters across environments
      </p>

      <StatCards clusters={clusterList} />

      <ClusterList
        clusters={clusterList}
        onResourceQuota={handleResourceQuota}
      />

      <ResourceQuotaDialog
        isOpen={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        selectedCluster={selectedCluster}
        onSave={handleSaveQuota}
      />
    </div>
  );
}