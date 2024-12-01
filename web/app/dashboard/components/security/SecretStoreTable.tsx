import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { StoreActions } from './StoreActions';
import { SecretStore } from '@/types/security';
import { getHealthBadgeColor, getProviderBadgeColor, getTypeBadgeColor } from './utils';

interface SecretStoreTableProps {
  stores: SecretStore[];
  onYAMLClick: (store: SecretStore) => void;
}

export function SecretStoreTable({ stores, onYAMLClick }: SecretStoreTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Provider</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Path</TableHead>
          <TableHead>Health</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {stores.map((store) => (
          <TableRow key={store.id}>
            <TableCell className="font-medium">{store.name}</TableCell>
            <TableCell>
              <Badge className={getProviderBadgeColor(store.provider)}>
                {store.provider}
              </Badge>
            </TableCell>
            <TableCell>
              <Badge className={getTypeBadgeColor(store.type)}>
                {store.type}
              </Badge>
            </TableCell>
            <TableCell className="font-mono text-sm">{store.path}</TableCell>
            <TableCell>
              <Badge className={getHealthBadgeColor(store.health.status)}>
                {store.health.status}
              </Badge>
            </TableCell>
            <TableCell className="text-right">
              <StoreActions store={store} onYAMLClick={onYAMLClick} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}