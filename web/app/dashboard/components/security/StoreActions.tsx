import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { FileText, MoreHorizontal, Trash2 } from "lucide-react";
import { SecretStore } from '@/types/security';

interface StoreActionsProps {
  store: SecretStore;
  onYAMLClick: (store: SecretStore) => void;
}

export function StoreActions({ store, onYAMLClick }: StoreActionsProps) {
  return (
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
          onClick={() => onYAMLClick(store)}
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
}