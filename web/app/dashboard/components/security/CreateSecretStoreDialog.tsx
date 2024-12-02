"use client";

import { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Editor } from '@monaco-editor/react';
import { load as yamlLoad } from 'js-yaml';

interface YamlDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  defaultValue?: string;
  onSave: (yaml: string) => void;
}

export function YamlDialog({
  isOpen,
  onOpenChange,
  title,
  description,
  defaultValue,
  onSave
}: YamlDialogProps) {
  const [yamlContent, setYamlContent] = useState(defaultValue || '');
  const [yamlValidation, setYamlValidation] = useState<{
    isValid: boolean;
    error?: string;
  }>({ isValid: true });

  useEffect(() => {
    if (defaultValue) {
      setYamlContent(defaultValue);
    }
  }, [defaultValue]);

  const validateYAML = (yamlString: string) => {
    try {
      yamlLoad(yamlString);
      setYamlValidation({ isValid: true });
    } catch (error) {
      setYamlValidation({
        isValid: false,
        error: (error as Error).message
      });
    }
  };

  useEffect(() => {
    validateYAML(yamlContent);
  }, [yamlContent]);

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[800px] h-[600px]">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="flex-1 min-h-[400px]">
          <Editor
            height="400px"
            defaultLanguage="yaml"
            value={yamlContent}
            onChange={(value) => setYamlContent(value || '')}
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
          <div className="mt-2">
            {yamlValidation.isValid ? (
              <div className="flex items-center text-sm text-green-600 dark:text-green-400">
                <svg
                  className="w-4 h-4 mr-1"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                YAML syntax is valid
              </div>
            ) : (
              <div className="flex items-center text-sm text-red-600 dark:text-red-400">
                <svg
                  className="w-4 h-4 mr-1"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
                {yamlValidation.error || 'Invalid YAML syntax'}
              </div>
            )}
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={() => {
              onSave(yamlContent);
              onOpenChange(false);
            }}
            disabled={!yamlValidation.isValid}
          >
            {defaultValue ? 'Apply' : 'Create'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}