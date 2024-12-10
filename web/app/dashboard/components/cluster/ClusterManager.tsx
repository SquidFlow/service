import { useState, useEffect } from "react";
import { StatCards } from './StatCards';
import { ClusterList } from './ClusterList';
import { ResourceQuotaDialog } from './ResourceQuotaDialog';
import { ClusterInfo } from '@/types/cluster';
import { UIResourceQuota, toClusterQuota } from '@/types/quota';
import { useClusterStore } from '@/store';

export function ClusterManager() {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const {
    data: clusters,
    selectedCluster,
    setSelectedCluster,
    getClusterList,
    updateClusterQuota
  } = useClusterStore();

  useEffect(() => {
    getClusterList();
  }, [getClusterList]);

  const handleResourceQuota = (cluster: ClusterInfo) => {
    setSelectedCluster(cluster);
    setIsDialogOpen(true);
  };

  const handleSaveQuota = async (quotas: UIResourceQuota) => {
    if (!selectedCluster) return;
    try {
      const clusterQuota = toClusterQuota(quotas);
      await updateClusterQuota(selectedCluster.name, clusterQuota);
    } catch (error) {
      console.error('Failed to update resource quota:', error);
    }
  };

  return (
    <div className="space-y-6">
      <p className="text-muted-foreground">
        Manage and monitor your Kubernetes clusters across environments
      </p>

      <StatCards clusters={clusters} />

      <ClusterList onResourceQuota={handleResourceQuota} />

      <ResourceQuotaDialog
        isOpen={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        selectedCluster={selectedCluster}
        onSave={handleSaveQuota}
      />
    </div>
  );
}