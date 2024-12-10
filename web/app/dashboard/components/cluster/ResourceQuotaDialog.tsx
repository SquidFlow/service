"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { ClusterInfo } from '@/types';

interface ResourceQuotaDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  selectedCluster: ClusterInfo | null;
  onSave: (quotas: any) => void;
}

export function ResourceQuotaDialog({
  isOpen,
  onOpenChange,
  selectedCluster,
  onSave
}: ResourceQuotaDialogProps) {
  const [quotas, setQuotas] = useState({
    cpu: selectedCluster?.quota?.cpu || "0",
    memory: selectedCluster?.quota?.memory || "0",
    storage: selectedCluster?.quota?.storage || "0",
    pods: selectedCluster?.quota?.pods || "0"
  });

  useEffect(() => {
    if (selectedCluster?.quota) {
      setQuotas({
        cpu: selectedCluster.quota.cpu,
        memory: selectedCluster.quota.memory,
        storage: selectedCluster.quota.storage,
        pods: selectedCluster.quota.pods
      });
    }
  }, [selectedCluster]);

  const handleSave = () => {
    onSave(quotas);
    onOpenChange(false);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Resource Quota</DialogTitle>
          <DialogDescription>
            Configure resource quotas for {selectedCluster?.name}
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="cpu">CPU</Label>
            <Input
              id="cpu"
              value={quotas.cpu}
              onChange={(e) => setQuotas(prev => ({ ...prev, cpu: e.target.value }))}
              placeholder="e.g., 4 cores"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="memory">Memory</Label>
            <Input
              id="memory"
              value={quotas.memory}
              onChange={(e) => setQuotas(prev => ({ ...prev, memory: e.target.value }))}
              placeholder="e.g., 8Gi"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="storage">Storage</Label>
            <Input
              id="storage"
              value={quotas.storage}
              onChange={(e) => setQuotas(prev => ({ ...prev, storage: e.target.value }))}
              placeholder="e.g., 100Gi"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="pods">Pods</Label>
            <Input
              id="pods"
              value={quotas.pods}
              onChange={(e) => setQuotas(prev => ({ ...prev, pods: e.target.value }))}
              placeholder="e.g., 10"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}