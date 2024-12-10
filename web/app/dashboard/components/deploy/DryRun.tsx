import { useState } from "react";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ScrollArea } from "@/components/ui/scroll-area";
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";

interface DryRunProps {
  isOpen: boolean;
  onClose: () => void;
  yamls: Array<{
    environment: string;
    manifest: string;
    is_valid: boolean;
  }>;
}

export function DryRun({ isOpen, onClose, yamls }: DryRunProps) {
  const [selectedEnv, setSelectedEnv] = useState(yamls[0]?.environment || '');

  const currentYaml = yamls.find(y => y.environment === selectedEnv);
  const manifest = currentYaml?.manifest || '';

  return (
    <Sheet open={isOpen} onOpenChange={onClose}>
      <SheetContent side="right" className="w-[90vw] sm:max-w-[1200px] p-0">
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b">
            <div className="flex items-center space-x-4">
              <span className="text-sm font-medium">Environment:</span>
              <Select value={selectedEnv} onValueChange={setSelectedEnv}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="Select environment" />
                </SelectTrigger>
                <SelectContent>
                  {yamls.map(yaml => (
                    <SelectItem key={yaml.environment} value={yaml.environment}>
                      {yaml.environment}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {currentYaml && (
                <span className={`text-sm ${currentYaml.is_valid ? 'text-green-500' : 'text-red-500'}`}>
                  {currentYaml.is_valid ? 'Valid' : 'Invalid'}
                </span>
              )}
            </div>
            <Button variant="ghost" size="icon" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>

          {/* Content */}
          <ScrollArea className="flex-1 p-6">
            <pre className="text-sm whitespace-pre-wrap font-mono">
              {manifest}
            </pre>
          </ScrollArea>
        </div>
      </SheetContent>
    </Sheet>
  );
}