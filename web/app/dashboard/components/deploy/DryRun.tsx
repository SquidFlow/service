import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Light as SyntaxHighlighter } from "react-syntax-highlighter";
import yaml from "react-syntax-highlighter/dist/esm/languages/hljs/yaml";
import { vs2015 } from "react-syntax-highlighter/dist/esm/styles/hljs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

SyntaxHighlighter.registerLanguage("yaml", yaml);

interface DryRunProps {
  isOpen: boolean;
  yamls: { cluster: string; content: string }[];
  onClose: () => void;
}

export function DryRun({ isOpen, yamls, onClose }: DryRunProps) {
  const [selectedCluster, setSelectedCluster] = useState(yamls[0]?.cluster || '');
  const selectedYaml = yamls.find(y => y.cluster === selectedCluster)?.content || '';

  return (
    <div
      className={`fixed top-0 right-0 w-[600px] h-screen bg-background border-l border-border transform transition-transform duration-300 ease-in-out ${
        isOpen ? "translate-x-0" : "translate-x-full"
      }`}
    >
      <div className="h-full flex flex-col p-6">
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-semibold">Dry Run Results</h2>
          <Button variant="ghost" onClick={onClose}>
            Close
          </Button>
        </div>

        <div className="mb-4">
          <Select value={selectedCluster} onValueChange={setSelectedCluster}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="Select cluster" />
            </SelectTrigger>
            <SelectContent>
              {yamls.map(({ cluster }) => (
                <SelectItem key={cluster} value={cluster}>
                  {cluster}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex-1 relative rounded-md overflow-hidden">
          <SyntaxHighlighter
            language="yaml"
            style={vs2015}
            customStyle={{
              margin: 0,
              padding: '1rem',
              fontSize: '0.875rem',
              borderRadius: '0.375rem',
              height: '100%',
              overflow: 'auto'
            }}
          >
            {selectedYaml}
          </SyntaxHighlighter>
        </div>
      </div>
    </div>
  );
}