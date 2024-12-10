"use client";

import { useState, useEffect } from "react";
import { useClusterStore } from '@/store';
import { ClusterTable } from './ClusterTable';
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ClusterInfo } from '@/types/cluster';

interface ClusterListProps {
  onResourceQuota: (cluster: ClusterInfo) => void;
}

export function ClusterList({ onResourceQuota }: ClusterListProps) {
  const {
    searchTerm,
    selectedEnvironment,
    selectedProvider,
    setSearchTerm,
    setSelectedEnvironment,
    setSelectedProvider,
    getFilteredClusters,
    getClusterList
  } = useClusterStore();

  const [selectedClusters, setSelectedClusters] = useState<string[]>([]);
  const filteredClusters = getFilteredClusters();

  const handleRefresh = async () => {
    try {
      await getClusterList();
    } catch (error) {
      console.error('Failed to refresh clusters:', error);
    }
  };

  useEffect(() => {
    handleRefresh();
  }, []);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Clusters</h2>
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <Select value={selectedEnvironment} onValueChange={setSelectedEnvironment}>
              <SelectTrigger className="w-[180px] bg-background">
                <SelectValue placeholder="All Environments" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="All Environments">All Environments</SelectItem>
                <SelectItem value="SIT">SIT</SelectItem>
                <SelectItem value="UAT">UAT</SelectItem>
                <SelectItem value="PRD">PRD</SelectItem>
              </SelectContent>
            </Select>
            <Select value={selectedProvider} onValueChange={setSelectedProvider}>
              <SelectTrigger className="w-[180px] bg-background">
                <SelectValue placeholder="All Providers" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="All Providers">All Providers</SelectItem>
                <SelectItem value="GKE">GKE</SelectItem>
                <SelectItem value="OCP">OCP</SelectItem>
                <SelectItem value="AKS">AKS</SelectItem>
                <SelectItem value="EKS">EKS</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button
            variant="outline"
            disabled
            className="opacity-50 cursor-not-allowed"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Cluster
          </Button>
        </div>
      </div>

      <ClusterTable
        clusters={filteredClusters}
        onResourceQuota={onResourceQuota}
        selectedClusters={selectedClusters}
        onSelectedChange={setSelectedClusters}
      />
    </div>
  );
}