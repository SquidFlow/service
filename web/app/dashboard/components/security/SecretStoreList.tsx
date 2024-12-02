import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { SecretStore } from '@/types/security';
import { useSecretStore } from '@/store';
import { SecretStoreTable } from './SecretStoreTable';

interface SecretStoreListProps {
  onYAMLClick: (store: SecretStore) => void;
  onCreateNew: () => void;
}

export function SecretStoreList({ onYAMLClick, onCreateNew }: SecretStoreListProps) {
  const { data: secretStoreList, fetch: fetchSecretStores } = useSecretStore();
  const [searchTerm, setSearchTerm] = useState("");

  // 在组件挂载时获取数据
  useEffect(() => {
    fetchSecretStores();
  }, [fetchSecretStores]);

  const filteredSecretStores = secretStoreList.filter(store =>
    store.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    store.provider.toLowerCase().includes(searchTerm.toLowerCase()) ||
    store.type.toLowerCase().includes(searchTerm.toLowerCase())
  );

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
        <Button
          className="flex items-center space-x-2"
          onClick={onCreateNew}
        >
          <Plus className="h-4 w-4" />
          <span>Create SecretStore</span>
        </Button>
      </CardHeader>
      <CardContent>
        <SecretStoreTable
          stores={filteredSecretStores}
          onYAMLClick={onYAMLClick}
        />
      </CardContent>
    </Card>
  );
}