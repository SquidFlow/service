import { useState } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, RefreshCw, Search, ExternalLink, GitBranch, CircleCheckBig } from "lucide-react";
import { DeleteDialog } from './DeleteDialog';
import { useGetApplicationList } from '@/app/api';
import { ApplicationTemplate } from '@/types/application';
import { getStatusIcon } from './utils';

interface ApplicationListProps {
  onSelectApp: (app: ApplicationTemplate) => void;
  onCreateNew: () => void;
}

export function ApplicationList({ onSelectApp, onCreateNew }: ApplicationListProps) {
  const { applications, mutate: refreshApplications } = useGetApplicationList();
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedApps, setSelectedApps] = useState<string[]>([]);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  const filteredApps = applications.filter((app: ApplicationTemplate) =>
    app.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    app.owner.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleRefresh = async () => {
    try {
      await refreshApplications();
    } catch (error) {
      console.error('Failed to refresh applications:', error);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">Applications</h1>
        <Button onClick={onCreateNew} variant="default">
          <Plus className="h-4 w-4 mr-2" />
          New Application
        </Button>
      </div>

      <p className="text-muted-foreground">
        Manage and monitor your ArgoCD applications across environments
      </p>

      <div className="flex items-center justify-between p-4 border-b bg-muted/50 rounded-lg">
        <div className="flex items-center space-x-4 flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              placeholder="Search applications..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-[300px] pl-9 bg-background"
            />
          </div>
          <Button variant="outline" size="icon" onClick={handleRefresh}>
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
        {selectedApps.length > 0 && (
          <Button
            variant="outline"
            className="text-red-600"
            onClick={() => setIsDeleteDialogOpen(true)}
          >
            Delete Selected
          </Button>
        )}
      </div>

      <Table>
        <TableHeader>
          <TableRow className="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted bg-muted/50">
            <TableHead className="w-12">
              {filteredApps.length > 0 && (
                <Checkbox
                  checked={selectedApps.length === filteredApps.length && filteredApps.length > 0}
                  onCheckedChange={(checked) => {
                    if (checked) {
                      setSelectedApps(filteredApps.map(app => app.name));
                    } else {
                      setSelectedApps([]);
                    }
                  }}
                />
              )}
            </TableHead>
            <TableHead className="text-base font-semibold">Name</TableHead>
            <TableHead className="text-base font-semibold">Status</TableHead>
            <TableHead className="text-base font-semibold">Health</TableHead>
            <TableHead className="text-base font-semibold">Repository</TableHead>
            <TableHead className="text-base font-semibold">Environments</TableHead>
            <TableHead className="text-base font-semibold">Last Updated</TableHead>
            <TableHead className="text-base font-semibold text-right">ArgoCD</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {filteredApps.map((app: ApplicationTemplate) => (
            <TableRow key={app.id} className="border-b data-[state=selected]:bg-muted hover:bg-muted/50 transition-colors duration-200">
              <TableCell className="p-2 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px] py-4">
                <Checkbox
                  checked={selectedApps.includes(app.name)}
                  onCheckedChange={(checked) => {
                    if (checked) {
                      setSelectedApps([...selectedApps, app.name]);
                    } else {
                      setSelectedApps(selectedApps.filter(name => name !== app.name));
                    }
                  }}
                />
              </TableCell>
              <TableCell className="p-2 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px] py-4">
                <Button
                  variant="link"
                  onClick={() => onSelectApp(app)}
                  className="p-0 h-auto text-sm font-semibold hover:text-blue-600"
                >
                  {app.name}
                </Button>
              </TableCell>
              <TableCell className="p-2 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px] py-4">
                <div className="flex items-center space-x-2">
                  {getStatusIcon(app.runtime_status.status)}
                  <span className="text-sm">{app.runtime_status.status}</span>
                </div>
              </TableCell>
              <TableCell className="p-2 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px] py-4">
                <Badge
                  className="inline-flex items-center rounded-md border font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 border-transparent bg-primary text-primary-foreground shadow hover:bg-primary/80 text-sm px-3 py-1"
                >
                  <div className="flex items-center space-x-2">
                    <CircleCheckBig className="h-4 w-4 mr-1 text-green-500" />
                    <span>{app.runtime_status.health}</span>
                  </div>
                </Badge>
              </TableCell>
              <TableCell className="p-2 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px] py-4">
                <div className="flex items-center space-x-2">
                  <code className="px-3 py-1.5 bg-muted rounded text-sm">main</code>
                  <a
                    href={app.template?.source?.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-muted-foreground hover:text-primary"
                  >
                    {app.template?.source?.url?.replace('https://github.com/', '')}
                  </a>
                </div>
              </TableCell>
              <TableCell className="p-2 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px] py-4">
                <div className="flex flex-wrap gap-1.5">
                  {app.deployed_environments?.map((env: string) => (
                    <Badge key={env} variant="outline" className="text-sm px-3 py-1">
                      {env}
                    </Badge>
                  ))}
                </div>
              </TableCell>
              <TableCell className="p-2 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px] py-4">
                <time className="text-sm text-muted-foreground">
                  {new Date(app.lastUpdate).toLocaleString()}
                </time>
              </TableCell>
              <TableCell className="p-2 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px] text-right py-4">
                <Button
                  variant="link"
                  size="sm"
                  asChild
                  className="inline-flex items-center px-3 py-2 rounded-md text-base font-medium bg-blue-50 text-blue-600 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400 dark:hover:bg-blue-900/30 transition-colors duration-200 group"
                >
                  <a
                    href={app.argocdUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <ExternalLink className="h-4 w-4 mr-1 group-hover:translate-x-0.5 transition-transform duration-200" />
                    Console
                  </a>
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      <DeleteDialog
        isOpen={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        selectedApps={selectedApps}
        onDelete={() => {
          setSelectedApps([]);
          setIsDeleteDialogOpen(false);
          refreshApplications();
        }}
      />
    </div>
  );
}