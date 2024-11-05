import React, { useState } from 'react';
import { Button } from "@/components/ui/button";
import { Light as SyntaxHighlighter } from 'react-syntax-highlighter';
import yaml from 'react-syntax-highlighter/dist/esm/languages/hljs/yaml';
import { vs2015 } from 'react-syntax-highlighter/dist/esm/styles/hljs';
import { ChevronLeft } from 'lucide-react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

SyntaxHighlighter.registerLanguage('yaml', yaml);

interface DryRunProps {
  isOpen: boolean;
  yamls: { cluster: string; content: string }[];
  onClose: () => void;
}

export function DryRun({ isOpen, yamls, onClose }: DryRunProps) {
  const [selectedCluster, setSelectedCluster] = useState<string>(yamls[0]?.cluster || '');

  if (!isOpen) return null;

  const currentYaml = yamls.find(y => y.cluster === selectedCluster)?.content || '';

  return (
    <div className={`fixed right-0 top-0 h-full w-[45%] bg-white dark:bg-gray-900 shadow-xl transform transition-transform duration-300 ease-in-out ${
      isOpen ? 'translate-x-0' : 'translate-x-full'
    }`}>
      <div className="h-full flex flex-col">
        {/* Header */}
        <div className="p-4 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
          <div className="flex justify-between items-center mb-4">
            <div className="flex items-center space-x-2">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">DryRun Preview</h3>
              <span className="px-2 py-1 text-xs font-medium rounded-full bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">
                Preview Mode
              </span>
            </div>
            <Button
              variant="ghost"
              size="icon"
              onClick={onClose}
              className="hover:bg-gray-200 dark:hover:bg-gray-700 rounded-full"
            >
              <ChevronLeft className="h-6 w-6" />
            </Button>
          </div>
          {/* Cluster Selection */}
          <div className="flex items-center space-x-4">
            <span className="text-sm font-medium">Target Cluster:</span>
            <Select
              value={selectedCluster}
              onValueChange={setSelectedCluster}
            >
              <SelectTrigger className="w-[200px]">
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
        </div>

        {/* YAML Content */}
        <div className="flex-1 p-4 overflow-auto bg-gray-50 dark:bg-gray-900">
          <div className="rounded-lg overflow-hidden border border-gray-200 dark:border-gray-600 shadow-sm">
            {/* YAML Header */}
            <div className="bg-white dark:bg-gray-800 px-4 py-2 border-b border-gray-200 dark:border-gray-600 flex items-center justify-between">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-200">
                {selectedCluster}/deployment.yaml
              </span>
              <Button
                variant="ghost"
                size="sm"
                className="text-gray-600 hover:text-gray-900 dark:text-gray-300 dark:hover:text-gray-100"
                onClick={() => {
                  navigator.clipboard.writeText(currentYaml);
                }}
              >
                Copy
              </Button>
            </div>
            {/* YAML Content */}
            <div className="bg-[#1E1E1E] dark:bg-gray-950">
              <SyntaxHighlighter
                language="yaml"
                style={vs2015}
                customStyle={{
                  margin: 0,
                  padding: '1.25rem',
                  background: 'transparent',
                  fontSize: '0.875rem',
                  lineHeight: '1.5',
                }}
                showLineNumbers={true}
                lineNumberStyle={{
                  color: '#6B7280',
                  paddingRight: '1em',
                  borderRight: '1px solid #374151',
                  marginRight: '1em',
                }}
              >
                {currentYaml}
              </SyntaxHighlighter>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700">
          <div className="flex justify-end space-x-4">
            <Button
              variant="outline"
              onClick={onClose}
              className="w-24 bg-white dark:bg-gray-900 hover:bg-gray-50 dark:hover:bg-gray-800"
            >
              Cancel
            </Button>
            <Button
              onClick={() => {
                console.log('Submitting DryRun configuration...');
                onClose();
              }}
              className="w-24 bg-blue-600 hover:bg-blue-700 text-white"
            >
              Submit
            </Button>
          </div>
          <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
            Review the generated YAML configuration before submitting
          </p>
        </div>
      </div>
    </div>
  );
}
