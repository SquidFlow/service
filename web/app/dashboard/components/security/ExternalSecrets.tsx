"use client";

import { useState } from 'react';
import { Button } from "@/components/ui/button";
import { Plus, Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { SecretStoreList } from './SecretStoreList';
import { YamlDialog } from './CreateSecretStoreDialog';
import { useSecretStore } from "@/store";

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

export function ExternalSecrets() {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const secretStore = useSecretStore();

  const handleCreateStore = async (yaml: string) => {
    try {
      await secretStore.create(yaml);
      setIsDialogOpen(false);
    } catch (error) {
      console.error('Failed to create secret store:', error);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <p className="text-muted-foreground">
          Manage external secrets and secret stores across environments
        </p>
        <Button onClick={() => setIsDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          New Secret Store
        </Button>
      </div>

      <div className="flex items-center justify-between p-4 border-b bg-muted/50 rounded-lg">
        <div className="flex items-center space-x-4 flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              placeholder="Search secret stores..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-[300px] pl-9 bg-background"
            />
          </div>
        </div>
      </div>

      <SecretStoreList searchTerm={searchTerm} />

      <YamlDialog
        isOpen={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        title="Create SecretStore"
        description="Create a new SecretStore by providing the YAML configuration."
        defaultValue={defaultYAML}
        onSave={handleCreateStore}
      />
    </div>
  );
}