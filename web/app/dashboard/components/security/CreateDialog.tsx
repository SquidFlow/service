import { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Editor } from '@monaco-editor/react';
import { load as yamlLoad } from 'js-yaml';

interface CreateDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

const defaultYAML = `apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: my-secret-store
  namespace: default
spec:
  provider:
    vault:
      server: "https://vault.your-domain.com"
      path: "path/to/secrets"
      version: v2
      auth:
        tokenSecretRef:
          name: "vault-token"
          key: "token"`;

export function CreateDialog({ isOpen, onOpenChange }: CreateDialogProps) {
  const [createYAML, setCreateYAML] = useState(defaultYAML);
  const [yamlValidation, setYamlValidation] = useState<{
    isValid: boolean;
    error?: string;
  }>({ isValid: true });

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
    validateYAML(createYAML);
  }, [createYAML]);

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[800px] h-[600px]">
        <DialogHeader>
          <DialogTitle>Create SecretStore</DialogTitle>
          <DialogDescription>
            Create a new SecretStore by providing the YAML configuration.
          </DialogDescription>
        </DialogHeader>
        <div className="flex-1 min-h-[400px]">
          <Editor
            height="400px"
            defaultLanguage="yaml"
            value={createYAML}
            onChange={(value) => setCreateYAML(value || '')}
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
              console.log('Creating SecretStore:', createYAML);
              onOpenChange(false);
            }}
            disabled={!yamlValidation.isValid}
          >
            Create
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}