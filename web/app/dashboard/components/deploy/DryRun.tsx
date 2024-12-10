import { useState } from "react";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism';

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
      <SheetContent
        side="right"
        className="w-[90vw] sm:max-w-[1200px] p-0"
      >
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b bg-background">
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
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                  currentYaml.is_valid
                    ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                    : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                }`}>
                  {currentYaml.is_valid ? 'Valid' : 'Invalid'}
                </span>
              )}
            </div>
          </div>

          {/* Content */}
          <ScrollArea className="flex-1">
            <div className="p-6">
              <SyntaxHighlighter
                language="yaml"
                style={oneDark}
                customStyle={{
                  margin: 0,
                  borderRadius: '0.5rem',
                  fontSize: '0.875rem',
                  fontFamily: 'JetBrains Mono, Consolas, Monaco, monospace',
                }}
                showLineNumbers
                wrapLines
                wrapLongLines
              >
                {manifest}
              </SyntaxHighlighter>
            </div>
          </ScrollArea>
        </div>
      </SheetContent>
    </Sheet>
  );
}