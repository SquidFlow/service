"use client";

import { useState, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Plus,
  RefreshCw,
  Search,
  ExternalLink,
  Trash2,
  RotateCw,
  CheckCircle2
} from "lucide-react";
import { DeleteDialog } from './DeleteDialog';
import { useApplicationStore } from '@/store';
import type { ApplicationTemplate } from '@/types';
import { getStatusIcon } from './utils';

interface ApplicationListProps {
  onSelectApp: (app: ApplicationTemplate) => void;
  onCreateNew: () => void;
}

const formatDate = (dateString?: string) => {
  if (!dateString) return 'N/A';
  try {
    return new Date(dateString).toLocaleString();
  } catch (error) {
    console.error('Error formatting date:', error);
    return 'Invalid date';
  }
};

export function ApplicationList({ onSelectApp, onCreateNew }: ApplicationListProps) {
  const { data: applications, fetch: fetchApplications, deleteApplications } = useApplicationStore();
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedApps, setSelectedApps] = useState<string[]>([]);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  useEffect(() => {
    fetchApplications();
  }, [fetchApplications]);

  const filteredApps = applications.filter((app: ApplicationTemplate) => {
    const appName = (app.name || '').toLowerCase();
    const appOwner = (app.created_by || '').toLowerCase();
    const searchLower = searchTerm.toLowerCase();

    return appName.includes(searchLower) || appOwner.includes(searchLower);
  });

  const handleRefresh = async () => {
    try {
      await fetchApplications();
    } catch (error) {
      console.error('Failed to refresh applications:', error);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center space-x-4">
        <div className="relative flex-1">
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
        {selectedApps.length > 0 && (
          <>
            <Button
              variant="outline"
              className="text-red-600"
              onClick={() => setIsDeleteDialogOpen(true)}
            >
              <Trash2 className="h-4 w-4 mr-2" />
              Delete Selected
            </Button>
            <Button variant="outline">
              <RotateCw className="h-4 w-4 mr-2" />
              Sync Selected
            </Button>
          </>
        )}
        <Button onClick={onCreateNew} className="ml-auto">
          <Plus className="h-4 w-4 mr-2" />
          New Application
        </Button>
      </div>

      <Table>
        <TableHeader>
          <TableRow className="bg-muted/50">
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
            <TableHead>Name</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Health</TableHead>
            <TableHead>Repository</TableHead>
            <TableHead>Environments</TableHead>
            <TableHead>Last Updated</TableHead>
            <TableHead className="text-right">ArgoCD</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {filteredApps.map((app: ApplicationTemplate) => (
            <TableRow
              key={`${app.id || ''}-${app.name}`}
              className="border-b data-[state=selected]:bg-muted hover:bg-muted/50 transition-colors duration-200"
            >
              <TableCell>
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
              <TableCell>
                <Button
                  variant="link"
                  onClick={() => onSelectApp(app)}
                  className="p-0 h-auto text-sm font-semibold hover:text-blue-600"
                >
                  {app.name}
                </Button>
              </TableCell>
              <TableCell>
                <div className="flex items-center space-x-2">
                  {getStatusIcon(app.runtime_status.status)}
                  <span className="text-sm">{app.runtime_status.status}</span>
                </div>
              </TableCell>
              <TableCell>
                <Badge
                  className="inline-flex items-center rounded-md border font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 border-transparent bg-primary text-primary-foreground shadow hover:bg-primary/80 text-sm px-3 py-1"
                >
                  <div className="flex items-center space-x-2">
                    <CheckCircle2 className="h-4 w-4 mr-1 text-green-500" />
                    <span>{app.runtime_status.health}</span>
                  </div>
                </Badge>
              </TableCell>
              <TableCell>
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
              <TableCell>
                <div className="flex flex-wrap gap-1.5">
                  {app.deployed_environments?.map((env: string) => (
                    <Badge
                      key={`${app.id || app.name}-${env}`}
                      variant="outline"
                      className="text-sm px-3 py-1"
                    >
                      {env}
                    </Badge>
                  ))}
                </div>
              </TableCell>
              <TableCell>
                <time className="text-sm text-muted-foreground">
                  {formatDate(app.runtime_status.last_update)}
                </time>
              </TableCell>
              <TableCell>
                <Button
                  variant="link"
                  size="sm"
                  asChild
                  className="inline-flex items-center px-3 py-2 rounded-md text-base font-medium bg-blue-50 text-blue-600 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400 dark:hover:bg-blue-900/30 transition-colors duration-200 group"
                >
                  <a
                    href={app.argocd_url}
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
        onDelete={async () => {
          await deleteApplications(selectedApps);
          setSelectedApps([]);
          setIsDeleteDialogOpen(false);
        }}
      />
    </div>
  );
}