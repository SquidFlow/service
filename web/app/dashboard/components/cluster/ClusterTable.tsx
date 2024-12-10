"use client";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Server,
  CircleHelp,
  ExternalLink,
  CircleCheck,
  MoreVertical,
  Cpu,
  MemoryStick,
  HardDrive,
  Network,
  Box
} from "lucide-react";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import type { ClusterInfo } from '@/types';
import { Checkbox } from "@/components/ui/checkbox";

interface TableColumnProps {
  children: React.ReactNode;
  tooltip?: string;
}

function TableColumn({ children, tooltip }: TableColumnProps) {
  return (
    <div className="flex items-center gap-1">
      {children}
      {tooltip && (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger>
              <CircleHelp className="h-4 w-4 text-gray-400" />
            </TooltipTrigger>
            <TooltipContent>{tooltip}</TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )}
    </div>
  );
}

function getProviderIconColor(provider: string) {
  switch (provider.toLowerCase()) {
    case 'gke':
      return 'bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-400';
    case 'ocp':
      return 'bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-400';
    default:
      return 'bg-gray-50 text-gray-600 dark:bg-gray-900/20 dark:text-gray-400';
  }
}

function getProviderBadgeColor(provider: string) {
  switch (provider.toLowerCase()) {
    case 'gke':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400';
    case 'ocp':
      return 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-400';
  }
}

interface ClusterTableProps {
  clusters: ClusterInfo[];
  onResourceQuota: (cluster: ClusterInfo) => void;
  selectedClusters: string[];
  onSelectedChange: (selected: string[]) => void;
}

interface MonitoringLinkProps {
  href?: string;
  label: string;
  className: string;
}

function MonitoringLink({ href, label, className }: MonitoringLinkProps) {
  if (!href) return null;

  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className={`px-2 py-1 text-xs rounded-full transition-colors hover:underline ${className}`}
    >
      {label}
    </a>
  );
}

export function ClusterTable({
  clusters,
  onResourceQuota,
  selectedClusters,
  onSelectedChange
}: ClusterTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted bg-gray-50 dark:bg-gray-900/50">
          <TableHead className="w-12">
            <Checkbox
              checked={selectedClusters.length === clusters.length && clusters.length > 0}
              onCheckedChange={(checked) => {
                if (checked) {
                  onSelectedChange(clusters.map(c => c.name));
                } else {
                  onSelectedChange([]);
                }
              }}
            />
          </TableHead>
          <TableHead><TableColumn tooltip="Cluster name and region">Name</TableColumn></TableHead>
          <TableHead><TableColumn tooltip="Environment type">Environment</TableColumn></TableHead>
          <TableHead><TableColumn tooltip="Kubernetes provider">Provider</TableColumn></TableHead>
          <TableHead><TableColumn tooltip="Cluster health status">Health</TableColumn></TableHead>
          <TableHead><TableColumn tooltip="Node status">Nodes</TableColumn></TableHead>
          <TableHead><TableColumn tooltip="Kubernetes version">Platform Version</TableColumn></TableHead>
          <TableHead><TableColumn tooltip="Resource quota limits">Tenant Resource Quota</TableColumn></TableHead>
          <TableHead><TableColumn tooltip="Monitoring tools">Monitoring</TableColumn></TableHead>
          <TableHead><TableColumn tooltip="Management console">Console</TableColumn></TableHead>
          <TableHead><TableColumn tooltip="Additional actions">Actions</TableColumn></TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {clusters.map((cluster) => (
          <TableRow
            key={cluster.name}
            className="border-b data-[state=selected]:bg-muted transition-all duration-200 hover:bg-gray-50 dark:hover:bg-gray-900/50"
          >
            <TableCell>
              <Checkbox
                checked={selectedClusters.includes(cluster.name)}
                onCheckedChange={(checked) => {
                  if (checked) {
                    onSelectedChange([...selectedClusters, cluster.name]);
                  } else {
                    onSelectedChange(selectedClusters.filter(name => name !== cluster.name));
                  }
                }}
              />
            </TableCell>
            <TableCell>
              <div className="flex items-center space-x-3">
                <div className={`p-2 rounded-lg transition-colors duration-200 ${getProviderIconColor(cluster.provider)}`}>
                  <Server className="h-5 w-5" />
                </div>
                <div className="space-y-1">
                  <div className="flex items-center space-x-2">
                    <span className="font-mono text-base font-bold tracking-wider px-2 py-0.5 rounded transition-colors duration-200 bg-gray-100 dark:bg-gray-800">
                      {cluster.name}
                    </span>
                    {cluster.builtin && (
                      <span className="px-2 py-1 rounded-full text-xs font-medium bg-purple-100 text-purple-800 ml-2">
                        Builtin
                      </span>
                    )}
                  </div>
                  <div className="flex items-center space-x-2 text-xs text-gray-500">
                    <span>{cluster.region}</span>
                  </div>
                </div>
              </div>
            </TableCell>
            <TableCell>{cluster.environment}</TableCell>
            <TableCell>
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${getProviderBadgeColor(cluster.provider)}`}>
                {cluster.provider}
              </span>
            </TableCell>
            <TableCell>
              <Button variant="ghost" className="p-0">
                <div className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400">
                  <CircleCheck className="h-3 w-3 mr-1" />
                  {cluster.health.status}
                </div>
              </Button>
            </TableCell>
            <TableCell>
              <div className="flex items-center space-x-2">
                <div className="flex items-baseline space-x-1 px-2 py-1 rounded-md bg-green-50 dark:bg-green-900/20">
                  <span className="text-sm font-medium text-green-700 dark:text-green-400">
                    {cluster.nodes.ready}
                  </span>
                  <span className="text-gray-400 dark:text-gray-500">/</span>
                  <span className="text-gray-600 dark:text-gray-400">
                    {cluster.nodes.total}
                  </span>
                </div>
                <span className="text-xs text-gray-500 dark:text-gray-400">ready</span>
              </div>
            </TableCell>
            <TableCell>{cluster.version.platform}</TableCell>
            <TableCell>
              <div className="space-y-2 transition-all duration-200">
                <ResourceQuotaItem
                  icon={<Cpu className="h-4 w-4 text-gray-500" />}
                  label="CPU"
                  value={cluster.resourceQuota.cpu}
                />
                <ResourceQuotaItem
                  icon={<MemoryStick className="h-4 w-4 text-gray-500" />}
                  label="Memory"
                  value={cluster.resourceQuota.memory}
                />
                <ResourceQuotaItem
                  icon={<HardDrive className="h-4 w-4 text-gray-500" />}
                  label="Storage"
                  value={cluster.resourceQuota.storage}
                />
                <ResourceQuotaItem
                  icon={<Network className="h-4 w-4 text-gray-500" />}
                  label="PVCs"
                  value={cluster.resourceQuota.pvcs}
                />
                <ResourceQuotaItem
                  icon={<Box className="h-4 w-4 text-gray-500" />}
                  label="NodePorts"
                  value={cluster.resourceQuota.nodeports}
                />
              </div>
            </TableCell>
            <TableCell>
              <div className="flex gap-1">
                {cluster.monitoring?.prometheus && (
                  <MonitoringLink
                    href={cluster.monitoring.urls?.prometheus}
                    label="Prometheus"
                    className="bg-blue-50 text-blue-700 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400"
                  />
                )}
                {cluster.monitoring?.grafana && (
                  <MonitoringLink
                    href={cluster.monitoring.urls?.grafana}
                    label="Grafana"
                    className="bg-orange-50 text-orange-700 hover:bg-orange-100 dark:bg-orange-900/20 dark:text-orange-400"
                  />
                )}
                {cluster.monitoring?.alertmanager && (
                  <MonitoringLink
                    href={cluster.monitoring.urls?.alertmanager}
                    label="Alertmanager"
                    className="bg-red-50 text-red-700 hover:bg-red-100 dark:bg-red-900/20 dark:text-red-400"
                  />
                )}
              </div>
            </TableCell>
            <TableCell>
              <a
                href={cluster.consoleUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-500 hover:text-blue-700 flex items-center gap-1"
              >
                <ExternalLink className="h-4 w-4" />
                <span>Console</span>
              </a>
            </TableCell>
            <TableCell>
              <Button
                variant="ghost"
                size="sm"
                className="hover:bg-gray-100 dark:hover:bg-gray-800"
              >
                <MoreVertical className="h-4 w-4 text-gray-600 dark:text-gray-400" />
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

function ResourceQuotaItem({ icon, label, value }: { icon: React.ReactNode; label: string; value: string | number }) {
  return (
    <div className="flex items-center justify-between text-sm">
      <div className="flex items-center space-x-2">
        {icon}
        <span className="">{label}:</span>
      </div>
      <span className="inline-flex px-2.5 py-0.5 rounded-md font-mono text-sm font-medium transition-colors duration-200 bg-gray-100 dark:bg-gray-800">
        {value}
      </span>
    </div>
  );
}