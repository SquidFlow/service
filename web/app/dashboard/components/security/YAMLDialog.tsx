import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Editor } from '@monaco-editor/react';

interface YAMLDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  yaml: string;
  onYAMLChange: (value: string) => void;
}

export function YAMLDialog({ isOpen, onOpenChange, yaml, onYAMLChange }: YAMLDialogProps) {
  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[800px] h-[600px]">
        <DialogHeader>
          <DialogTitle>Edit SecretStore YAML</DialogTitle>
          <DialogDescription>
            Edit the YAML configuration for your SecretStore.
          </DialogDescription>
        </DialogHeader>
        <div className="flex-1 min-h-[400px]">
          <Editor
            height="400px"
            defaultLanguage="yaml"
            value={yaml}
            onChange={(value) => onYAMLChange(value || '')}
            theme="vs-dark"
            options={{
              minimap: { enabled: false },
              scrollBeyondLastLine: false,
              fontSize: 14,
              lineNumbers: 'on',
              readOnly: false,
              wordWrap: 'on',
              automaticLayout: true,
            }}
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={() => {
            console.log('Saving YAML:', yaml);
            onOpenChange(false);
          }}>
            Apply
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}