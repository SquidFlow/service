import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useState } from "react";
import { useDeleteApplications } from "@/app/api";
import { useToast } from "@/components/ui/use-toast";

interface DeleteDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  selectedApps: string[];
  onDelete: () => void;
}

export function DeleteDialog({ isOpen, onOpenChange, selectedApps, onDelete }: DeleteDialogProps) {
  const [confirmationInput, setConfirmationInput] = useState("");
  const { deleteApplications, isLoading } = useDeleteApplications();
  const { toast } = useToast();

  const handleDelete = async () => {
    if (confirmationInput !== selectedApps.join(", ")) {
      toast({
        title: "Error",
        description: "Please enter the correct application name(s) to confirm deletion.",
        variant: "destructive",
      });
      return;
    }

    try {
      await deleteApplications(selectedApps);
      toast({
        title: "Success",
        description: `Successfully deleted ${selectedApps.length} application(s)`,
      });
      onDelete();
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete applications. Please try again.",
        variant: "destructive",
      });
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Application(s)</DialogTitle>
          <DialogDescription>
            This action cannot be undone. Please type the application name(s) to confirm:
            <code className="mt-2 block bg-muted p-2 rounded">{selectedApps.join(", ")}</code>
          </DialogDescription>
        </DialogHeader>
        <Input
          value={confirmationInput}
          onChange={(e) => setConfirmationInput(e.target.value)}
          placeholder="Enter application name(s)"
        />
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={handleDelete} disabled={isLoading}>
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}