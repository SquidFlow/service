import { useState } from 'react';
import { SecretStoreList } from './SecretStoreList';
import { YAMLDialog } from './YAMLDialog';
import { CreateDialog } from './CreateDialog';
import { SecretStore } from '@/types/security';
import { getSecretStoreYAML } from './utils';

interface SecurityProps {
  activeSubMenu: string;
}

export function Security({ activeSubMenu }: SecurityProps) {
  const [isYAMLDialogOpen, setIsYAMLDialogOpen] = useState(false);
  const [editingYAML, setEditingYAML] = useState<string>('');
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);

  const handleYAMLClick = (store: SecretStore) => {
    setEditingYAML(getSecretStoreYAML(store));
    setIsYAMLDialogOpen(true);
  };

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">{activeSubMenu}</h1>
      </div>
      {activeSubMenu === 'ExternalSecrets' && (
        <SecretStoreList onYAMLClick={handleYAMLClick} onCreateNew={() => setIsCreateDialogOpen(true)} />
      )}

      <YAMLDialog
        isOpen={isYAMLDialogOpen}
        onOpenChange={setIsYAMLDialogOpen}
        yaml={editingYAML}
        onYAMLChange={setEditingYAML}
      />

      <CreateDialog
        isOpen={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
      />
    </>
  );
}