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
import { useRouter } from 'next/navigation';

interface ApplicationListProps {
  onSelectApp: (app: ApplicationTemplate) => void;
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

export function ApplicationList({ onSelectApp }: ApplicationListProps) {
  const { data: applications, fetch: fetchApplications } = useApplicationStore();
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedApps, setSelectedApps] = useState<string[]>([]);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const router = useRouter();

  useEffect(() => {
    fetchApplications();
  }, [fetchApplications]);

  const filteredApps = applications.filter((app) => {
    const appName = (app.application_instantiation?.application_name || '').toLowerCase();
    const searchLower = searchTerm.toLowerCase();
    return appName.includes(searchLower);
  });

  const handleRefresh = async () => {
    try {
      await fetchApplications();
    } catch (error) {
      console.error('Failed to refresh applications:', error);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 p-4 rounded-lg border shadow-sm">
        <div className="flex items-center space-x-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search applications..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-[300px] pl-9 bg-background/50"
            />
          </div>
          <Button
            variant="outline"
            size="icon"
            onClick={handleRefresh}
            className="hover:bg-muted/80"
          >
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>

        <div className="flex items-center space-x-3">
          {selectedApps.length > 0 && (
            <>
              <Button
                variant="outline"
                className="text-destructive hover:bg-destructive/10"
                onClick={() => setIsDeleteDialogOpen(true)}
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Delete Selected
              </Button>
              <Button variant="outline" className="hover:bg-primary/10">
                <RotateCw className="h-4 w-4 mr-2" />
                Sync Selected
              </Button>
            </>
          )}
          <Button
            onClick={() => router.push('/dashboard/deploy/application/new')}
            className="bg-primary hover:bg-primary/90"
          >
            <Plus className="h-4 w-4 mr-2" />
            New Application
          </Button>
        </div>
      </div>

      <div className="rounded-lg border bg-card shadow-sm">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/50 hover:bg-muted/60">
              <TableHead className="w-12">
                {filteredApps.length > 0 && (
                  <Checkbox
                    checked={selectedApps.length === filteredApps.length && filteredApps.length > 0}
                    onCheckedChange={(checked) => {
                      if (checked) {
                        setSelectedApps(filteredApps.map(app => app.application_instantiation.application_name));
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
              <TableHead>Clusters</TableHead>
              <TableHead>Last Updated</TableHead>
              <TableHead className="text-right">ArgoCD</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredApps.map((app) => (
              <TableRow
                key={app.application_instantiation.application_name}
                className="border-b transition-colors hover:bg-muted/30 data-[state=selected]:bg-muted"
              >
                <TableCell>
                  <Checkbox
                    checked={selectedApps.includes(app.application_instantiation.application_name)}
                    onCheckedChange={(checked) => {
                      if (checked) {
                        setSelectedApps([...selectedApps, app.application_instantiation.application_name]);
                      } else {
                        setSelectedApps(selectedApps.filter(name => name !== app.application_instantiation.application_name));
                      }
                    }}
                  />
                </TableCell>
                <TableCell>
                  <Button
                    variant="link"
                    onClick={() => onSelectApp(app)}
                    className="p-0 h-auto text-sm font-semibold text-primary hover:text-primary/80"
                  >
                    {app.application_instantiation.application_name || 'N/A'}
                  </Button>
                </TableCell>
                <TableCell>
                  <div className="flex items-center space-x-2">
                    {getStatusIcon(app.application_runtime?.status || 'Unknown')}
                    <span className="text-sm">{app.application_runtime?.status || 'N/A'}</span>
                  </div>
                </TableCell>
                <TableCell>
                  <Badge
                    className="inline-flex items-center rounded-md border font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 border-transparent bg-primary text-primary-foreground shadow hover:bg-primary/80 text-sm px-3 py-1"
                  >
                    <div className="flex items-center space-x-2">
                      <CheckCircle2 className="h-4 w-4 mr-1 text-green-500" />
                      <span>{app.application_runtime?.health || 'N/A'}</span>
                    </div>
                  </Badge>
                </TableCell>
                <TableCell>
                  <div className="flex items-center space-x-2">
                    <code className="px-3 py-1.5 bg-muted rounded text-sm">main</code>
                    <a
                      href={app.application_source?.repo}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-muted-foreground hover:text-primary"
                    >
                      {app.application_source?.repo?.replace('https://github.com/', '') || 'N/A'}
                    </a>
                  </div>
                </TableCell>
                <TableCell>
                  <div className="flex flex-wrap gap-1.5">
                    {app.application_target?.map((target) => (
                      <Badge
                        key={`${app.application_instantiation.application_name}-${target.cluster}`}
                        variant="outline"
                        className="text-sm px-3 py-1"
                      >
                        {target.cluster}
                      </Badge>
                    )) || <span>N/A</span>}
                  </div>
                </TableCell>
                <TableCell>
                  <time className="text-sm text-muted-foreground">
                    {formatDate(app.application_runtime?.last_updated_at) || 'N/A'}
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
                      href="#" // TODO: 需要添加实际的 ArgoCD URL
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
      </div>

      <DeleteDialog
        isOpen={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        selectedApps={selectedApps}
        onDelete={async () => {
          setSelectedApps([]);
          setIsDeleteDialogOpen(false);
        }}
      />
    </div>
  );
}