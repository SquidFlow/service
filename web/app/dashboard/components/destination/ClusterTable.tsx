import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ExternalLink, AlertCircle } from "lucide-react";
import { ClusterInfo } from '@/types/cluster';
import { getClusterStatusColor } from './utils';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

interface ClusterTableProps {
  clusters: ClusterInfo[];
  onResourceQuota: (cluster: ClusterInfo) => void;
}

export function ClusterTable({ clusters, onResourceQuota }: ClusterTableProps) {
  const formatResourceValue = (value?: string) => value || 'N/A';

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Environment</TableHead>
          <TableHead>Provider</TableHead>
          <TableHead>Health</TableHead>
          <TableHead>Nodes</TableHead>
          <TableHead>Platform Version</TableHead>
          <TableHead>Resource Usage</TableHead>
          <TableHead>Monitoring</TableHead>
          <TableHead>Console</TableHead>
          <TableHead>Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {clusters.map((cluster) => (
          <TableRow key={cluster.name}>
            <TableCell className="font-medium">{cluster.name}</TableCell>
            <TableCell>
              <Badge variant="outline">{cluster.env}</Badge>
            </TableCell>
            <TableCell>
              <Badge variant="outline" className="capitalize">
                {cluster.provider.toLowerCase()}
              </Badge>
            </TableCell>
            <TableCell>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <Badge className={getClusterStatusColor(cluster.status)}>
                      {cluster.status}
                    </Badge>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Cluster Status: {cluster.status}</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </TableCell>
            <TableCell>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <span className="font-mono">
                      {cluster.nodes?.ready || 0}/{cluster.nodes?.total || 0}
                    </span>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>{cluster.nodes?.ready || 0} ready nodes out of {cluster.nodes?.total || 0} total nodes</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </TableCell>
            <TableCell>
              <span className="font-mono text-sm">
                {cluster.version?.platform || 'N/A'}
              </span>
            </TableCell>
            <TableCell>
              <div className="space-y-1">
                <div className="flex justify-between text-sm">
                  <span>CPU:</span>
                  <span className="font-mono">{formatResourceValue(cluster.resources?.cpu)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span>Memory:</span>
                  <span className="font-mono">{formatResourceValue(cluster.resources?.memory)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span>Storage:</span>
                  <span className="font-mono">{formatResourceValue(cluster.resources?.storage)}</span>
                </div>
              </div>
            </TableCell>
            <TableCell>
              <div className="flex space-x-2">
                {cluster.monitoring?.prometheus && (
                  <Button variant="ghost" size="sm" asChild>
                    <a
                      href={cluster.monitoring.urls?.prometheus}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center space-x-1"
                    >
                      <ExternalLink className="h-4 w-4" />
                      <span className="sr-only">Open Prometheus</span>
                    </a>
                  </Button>
                )}
              </div>
            </TableCell>
            <TableCell>
              {cluster.consoleUrl ? (
                <Button variant="ghost" size="sm" asChild>
                  <a
                    href={cluster.consoleUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center space-x-1"
                  >
                    <ExternalLink className="h-4 w-4" />
                    <span className="sr-only">Open Console</span>
                  </a>
                </Button>
              ) : (
                <span className="text-muted-foreground text-sm">N/A</span>
              )}
            </TableCell>
            <TableCell>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onResourceQuota(cluster)}
                className="hover:bg-blue-50 dark:hover:bg-blue-900/30"
              >
                Resource Quota
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}