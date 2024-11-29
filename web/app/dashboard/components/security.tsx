"use client"

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Plus, MoreHorizontal, Trash2, FileText } from "lucide-react"
import { SecretStore } from '@/types/security';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Editor } from '@monaco-editor/react';
import { load as yamlLoad } from 'js-yaml';
import { useGetSecretStore } from '@/app/api';

const getSecretStoreYAML = (store: SecretStore) => {
  const isCluster = store.type === 'ClusterSecretStore';
  return `apiVersion: external-secrets.io/v1beta1
kind: ${store.type}
metadata:
  name: ${store.name}
${!isCluster ? '  namespace: default\n' : ''}spec:
  provider:
    ${store.provider.toLowerCase()}:
      server: "https://vault.your-domain.com"
      path: ${store.path}
      version: v2
      auth:
        tokenSecretRef:
          name: "${store.provider.toLowerCase()}-token${isCluster ? '-global' : ''}"
          key: "token"
          ${isCluster ? 'namespace: external-secrets' : ''}`;
};

// 添加导出函数用于获取可用的 SecretStore
// export const getAvailableSecretStores = () => {
//   return secretStoreList.filter(store => store.health.status === 'Healthy');
// };

// 添加导出函数用于获取特定 SecretStore 的详细信息
// export const getSecretStoreDetails = (name: string) => {
//   return secretStoreList.find(store => store.name === name);
// };

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

type HealthStatus = 'Healthy' | 'Warning' | 'Error';

const getHealthBadgeColor = (status: string) => {
  switch (status.toLowerCase()) {
    case 'healthy':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
    case 'warning':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
    case 'error':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400';
  }
};

const getProviderBadgeColor = (provider: string) => {
  const colors: Record<string, string> = {
    AWS: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
    GCP: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
    Azure: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
    Vault: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400',
    CyberArk: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
  };
  return colors[provider] || 'bg-gray-100 text-gray-800';
};

const getTypeBadgeColor = (type: string) => {
  return type === 'ClusterSecretStore'
    ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
    : 'bg-gray-100 text-gray-800 dark:bg-gray-800/30 dark:text-gray-400';
};

export function Security({ activeSubMenu }: { activeSubMenu: string }) {
  const { secretStoreList } = useGetSecretStore();
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedStores, setSelectedStores] = useState<number[]>([]);
  const [isYAMLDialogOpen, setIsYAMLDialogOpen] = useState(false);
  const [editingYAML, setEditingYAML] = useState<string>('');
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [createYAML, setCreateYAML] = useState(defaultYAML);
  const [yamlValidation, setYamlValidation] = useState<{
    isValid: boolean;
    error?: string;
  }>({ isValid: true });

  const filteredSecretStores = secretStoreList.filter(store =>
    store.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    store.provider.toLowerCase().includes(searchTerm.toLowerCase()) ||
    store.type.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const renderSecretStores = () => {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <div className="space-y-2">
            <CardTitle>Secret Stores</CardTitle>
            <Input
              placeholder="Search by name, provider, or type..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-[300px]"
            />
          </div>
          <div>
            <Button
              className="flex items-center space-x-2"
              onClick={() => {
                setCreateYAML(defaultYAML);
                setIsCreateDialogOpen(true);
              }}
            >
              <Plus className="h-4 w-4" />
              <span>Create SecretStore</span>
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Provider</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Path</TableHead>
                <TableHead>Health</TableHead>
                <TableHead>Last Synced</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredSecretStores.map((store) => (
                <TableRow key={store.id}>
                  <TableCell className="font-medium">{store.name}</TableCell>
                  <TableCell>
                    <Badge className={getProviderBadgeColor(store.provider)}>
                      {store.provider}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant={store.type === 'ClusterSecretStore' ? 'default' : 'secondary'}
                      className={getTypeBadgeColor(store.type)}
                    >
                      {store.type}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-mono text-sm">{store.path}</TableCell>
                  <TableCell>
                    <Badge className={getHealthBadgeColor(store.health.status)}>
                      {store.health.status}
                    </Badge>
                  </TableCell>
                  <TableCell>{new Date(store.lastSynced).toLocaleString()}</TableCell>
                  <TableCell className="text-right">
                    {renderDropdownMenu(store)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    );
  };

  const handleYAMLClick = (store: SecretStore) => {
    setEditingYAML(getSecretStoreYAML(store));
    setIsYAMLDialogOpen(true);
  };

  const renderYAMLDialog = () => {
    return (
      <Dialog open={isYAMLDialogOpen} onOpenChange={setIsYAMLDialogOpen}>
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
              value={editingYAML}
              onChange={(value) => setEditingYAML(value || '')}
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
            <Button variant="outline" onClick={() => setIsYAMLDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={() => {
              console.log('Saving YAML:', editingYAML);
              setIsYAMLDialogOpen(false);
            }}>
              Apply
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  };

  const renderDropdownMenu = (store: SecretStore) => (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="h-8 w-8 p-0">
          <span className="sr-only">Open menu</span>
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem
          className="flex items-center cursor-pointer"
          onClick={() => handleYAMLClick(store)}
        >
          <FileText className="mr-2 h-4 w-4" />
          <span>Edit</span>
        </DropdownMenuItem>
        <DropdownMenuItem
          className="flex items-center cursor-pointer text-red-600 focus:text-red-600 dark:focus:text-red-500"
          onClick={() => console.log('Delete:', store.name)}
        >
          <Trash2 className="mr-2 h-4 w-4" />
          <span>Delete</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );

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

  const renderCreateDialog = () => {
    return (
      <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
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
            <Button variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={() => {
                console.log('Creating SecretStore:', createYAML);
                setIsCreateDialogOpen(false);
              }}
              disabled={!yamlValidation.isValid}
            >
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  };

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">{activeSubMenu}</h1>
      </div>
      {activeSubMenu === 'ExternalSecrets' && renderSecretStores()}
      {renderYAMLDialog()}
      {renderCreateDialog()}
    </>
  );
}