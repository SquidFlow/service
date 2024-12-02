import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { HelpCircle } from "lucide-react";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { useState } from "react";
import { ClusterInfo } from '@/types/cluster';
import { UIResourceQuota, quotaFields, QuotaField } from '@/types/quota';

interface ResourceQuotaDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  selectedCluster?: ClusterInfo;
  onSave?: (quotas: UIResourceQuota) => void;
}

export function ResourceQuotaDialog({
  isOpen,
  onOpenChange,
  selectedCluster,
  onSave
}: ResourceQuotaDialogProps) {
  const [quotas, setQuotas] = useState<UIResourceQuota>({
    cpu: selectedCluster?.quota?.cpu || "0",
    memory: selectedCluster?.quota?.memory || "0",
    storage: selectedCluster?.quota?.storage || "0",
    pods: selectedCluster?.quota?.pods || "0"
  });

  const handleSave = () => {
    onSave?.(quotas);
    onOpenChange(false);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            Resource Quota Settings
            {selectedCluster && (
              <span className="ml-2 text-sm text-muted-foreground">
                for {selectedCluster.name}
              </span>
            )}
          </DialogTitle>
        </DialogHeader>
        <div className="space-y-6">
          {quotaFields.map((desc: QuotaField) => (
            <div key={desc.name} className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <span className="text-sm font-medium">{desc.label}</span>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <HelpCircle className="h-4 w-4 text-gray-400" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p className="max-w-xs">{desc.tooltip}</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
              <div className="flex items-center space-x-2">
                <Input
                  type="number"
                  value={quotas[desc.name]}
                  onChange={(e) => setQuotas(prev => ({
                    ...prev,
                    [desc.name]: e.target.value
                  }))}
                  className="w-24 font-mono"
                />
                <span className="text-sm text-gray-500 font-mono">
                  {desc.unit}
                </span>
              </div>
            </div>
          ))}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave}>
            Save Changes
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}