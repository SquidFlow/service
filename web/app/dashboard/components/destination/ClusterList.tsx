"use client";

import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, Plus, RefreshCw, Trash2 } from "lucide-react";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
import { useClusterStore } from '@/store';
import { ClusterTable } from './ClusterTable';
import { useState } from "react";

interface ClusterListProps {
  onResourceQuota: (cluster: any) => void;
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
    fetch: refreshClusters
  } = useClusterStore();

  const [selectedClusters, setSelectedClusters] = useState<string[]>([]);
  const filteredClusters = getFilteredClusters();

  const handleRefresh = async () => {
    try {
      await refreshClusters();
    } catch (error) {
      console.error('Failed to refresh clusters:', error);
    }
  };

  return (
    <Card>
      <div className="flex items-center justify-between p-4 border-b bg-muted/50">
        <div className="flex items-center space-x-4 flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              placeholder="Search clusters..."
              className="w-[300px] pl-9 bg-background"
              type="search"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          <Button variant="outline" size="icon" onClick={handleRefresh}>
            <RefreshCw className="h-4 w-4" />
          </Button>
          {selectedClusters.length > 0 && (
            <Button
              variant="outline"
              className="text-red-600"
              onClick={() => {
                // 处理删除逻辑
                setSelectedClusters([]);
              }}
            >
              <Trash2 className="h-4 w-4 mr-2" />
              Delete Selected
            </Button>
          )}
        </div>
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
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            Add Cluster
          </Button>
        </div>
      </div>
      <div className="p-4">
        <ClusterTable
          clusters={filteredClusters}
          onResourceQuota={onResourceQuota}
          selectedClusters={selectedClusters}
          onSelectedChange={setSelectedClusters}
        />
      </div>
    </Card>
  );
}