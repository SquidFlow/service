import { useState } from "react";
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
import { useToast } from "@/components/ui/use-toast";
import { useApplicationStore } from '@/store';

interface DeleteDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  selectedApps: string[];
  onDelete: () => void;
}

export function DeleteDialog({ isOpen, onOpenChange, selectedApps, onDelete }: DeleteDialogProps) {
  const [confirmationInput, setConfirmationInput] = useState("");
  const { deleteApplications, isLoading } = useApplicationStore();
  const { toast } = useToast();

  const handleDelete = async () => {
    if (confirmationInput !== "delete") {
      toast({
        variant: "destructive",
        title: "Invalid confirmation",
        description: "Please type 'delete' to confirm",
      });
      return;
    }

    try {
      await deleteApplications(selectedApps);
      toast({
        title: "Success",
        description: "Applications deleted successfully",
      });
      onDelete();
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Error",
        description: error instanceof Error ? error.message : "Failed to delete applications",
      });
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Applications</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete {selectedApps.length} application(s)? This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Type <span className="font-mono text-destructive">delete</span> to confirm
          </p>
          <Input
            value={confirmationInput}
            onChange={(e) => setConfirmationInput(e.target.value)}
            placeholder="Type 'delete' to confirm"
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={confirmationInput !== "delete" || isLoading}
          >
            {isLoading ? "Deleting..." : "Delete"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}