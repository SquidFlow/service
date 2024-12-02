"use client";

import { useEffect, useState } from 'react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { MoreHorizontal, FileText, Trash2 } from "lucide-react";
import { useSecretStore } from "@/store";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { YamlDialog } from './CreateSecretStoreDialog';
import type { SecretStore } from '@/types';

interface SecretStoreListProps {
  searchTerm: string;
}

const getProviderBadgeStyle = (provider: string) => {
  switch (provider.toLowerCase()) {
    case 'aws':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
    case 'vault':
      return 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400';
    case 'azure':
      return 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400';
    case 'gcp':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400';
    case 'cyberark':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-800/30 dark:text-gray-400';
  }
};

const getTypeStyle = (type: string) => {
  return type === 'ClusterSecretStore'
    ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
    : 'bg-gray-100 text-gray-800 dark:bg-gray-800/30 dark:text-gray-400';
};

const getHealthStyle = (status: string) => {
  switch (status.toLowerCase()) {
    case 'healthy':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
    case 'warning':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
    case 'error':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-800/30 dark:text-gray-400';
  }
};

export function SecretStoreList({ searchTerm }: SecretStoreListProps) {
  const { data: stores, fetch: fetchStores } = useSecretStore();
  const [selectedStore, setSelectedStore] = useState<SecretStore | null>(null);
  const [isEditorOpen, setIsEditorOpen] = useState(false);

  useEffect(() => {
    fetchStores();
  }, [fetchStores]);

  const filteredStores = stores.filter(store =>
    store.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    store.provider.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleEdit = (store: SecretStore) => {
    setSelectedStore(store);
    setIsEditorOpen(true);
  };

  const handleSaveYaml = async (yaml: string) => {
    if (!selectedStore) return;

    try {
      await useSecretStore.getState().update(selectedStore.name, yaml);
      setIsEditorOpen(false);
    } catch (error) {
      console.error('Failed to save YAML:', error);
    }
  };

  return (
    <>
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
          {filteredStores.map((store) => (
            <TableRow key={store.name}>
              <TableCell className="font-medium">{store.name}</TableCell>
              <TableCell>
                <Badge
                  variant="secondary"
                  className={getProviderBadgeStyle(store.provider)}
                >
                  {store.provider}
                </Badge>
              </TableCell>
              <TableCell>
                <Badge
                  variant="secondary"
                  className={getTypeStyle(store.type)}
                >
                  {store.type}
                </Badge>
              </TableCell>
              <TableCell className="font-mono text-sm">{store.path}</TableCell>
              <TableCell>
                <Badge
                  variant="secondary"
                  className={getHealthStyle(store.health.status)}
                >
                  {store.health.status}
                </Badge>
              </TableCell>
              <TableCell>{new Date(store.lastSynced).toLocaleString()}</TableCell>
              <TableCell className="text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 p-0"
                    >
                      <span className="sr-only">Open menu</span>
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem
                      className="flex items-center cursor-pointer"
                      onClick={() => handleEdit(store)}
                    >
                      <FileText className="mr-2 h-4 w-4" />
                      <span>Edit</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      className="flex items-center cursor-pointer text-red-600 focus:text-red-600 dark:focus:text-red-500"
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      <span>Delete</span>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      <YamlDialog
        isOpen={isEditorOpen}
        onOpenChange={setIsEditorOpen}
        title="Edit SecretStore YAML"
        description="Edit the YAML configuration for your SecretStore."
        defaultValue={selectedStore ? JSON.stringify(selectedStore, null, 2) : ''}
        onSave={handleSaveYaml}
      />
    </>
  );
}