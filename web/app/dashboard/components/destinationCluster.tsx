import { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Plus,
  Search,
  Settings2,
  ExternalLink,
  Layout,
  CheckCircle,
  LucideIcon,
  Server,
  AlertTriangle,
  XCircle,
  MoreVertical,
  Cpu,
  HelpCircle,
  MemoryStick,
  HardDrive,
  Network,
  Box,
} from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogClose,
  DialogFooter,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  ClusterInfo,
  clusters,
  ResourceQuota,
  resourceDescriptions,
  providerColorMap,
  healthStatusMap,
  monitoringTypeStyles,
  IconType,
} from "./destinationClusterMock";
import { useGetClusterList } from "@/app/api";

const icons: Record<IconType, LucideIcon> = {
  CheckCircle,
  AlertTriangle,
  XCircle,
  Settings2,
  ExternalLink,
  Layout,
  Server,
  Cpu,
  MemoryStick,
  HardDrive,
  Network,
  Box,
} as const;

const TableHeadWithTooltip = ({
  children,
  tooltip,
}: {
  children: React.ReactNode;
  tooltip: string;
}) => (
  <TooltipProvider>
    <TableHead>
      <div className="flex items-center gap-1">
        {children}
        <Tooltip>
          <TooltipTrigger>
            <HelpCircle className="h-4 w-4 text-gray-400" />
          </TooltipTrigger>
          <TooltipContent>
            <p className="max-w-xs">{tooltip}</p>
          </TooltipContent>
        </Tooltip>
      </div>
    </TableHead>
  </TooltipProvider>
);

const ResourceQuotaCell = ({
  quota,
  disabled,
}: {
  quota: ResourceQuota;
  disabled?: boolean;
}) => (
  <div
    className={`space-y-2 transition-all duration-200 ${
      disabled ? "opacity-75 grayscale" : ""
    }`}
  >
    {Object.entries(resourceDescriptions).map(([key, desc]) => (
      <div key={key} className="flex items-center justify-between text-sm">
        <TooltipProvider>
          <div className="flex items-center space-x-2">
            {key === "cpu" && (
              <Cpu
                className={`h-4 w-4 ${disabled ? "text-gray-400" : "text-gray-500"}`}
              />
            )}
            {key === "memory" && (
              <MemoryStick
                className={`h-4 w-4 ${disabled ? "text-gray-400" : "text-gray-500"}`}
              />
            )}
            {key === "storage" && (
              <HardDrive
                className={`h-4 w-4 ${disabled ? "text-gray-400" : "text-gray-500"}`}
              />
            )}
            {key === "pvcs" && (
              <Network
                className={`h-4 w-4 ${disabled ? "text-gray-400" : "text-gray-500"}`}
              />
            )}
            {key === "nodeports" && (
              <Box
                className={`h-4 w-4 ${disabled ? "text-gray-400" : "text-gray-500"}`}
              />
            )}
            <span className={disabled ? "text-gray-400" : ""}>
              {desc.label}:
            </span>
          </div>
          <Tooltip>
            <TooltipTrigger>
              <span
                className={`inline-flex px-2.5 py-0.5 rounded-md font-mono text-sm font-medium transition-colors duration-200 ${
                  disabled
                    ? "bg-gray-200 text-gray-500 dark:bg-gray-800 dark:text-gray-400"
                    : "bg-gray-100 dark:bg-gray-800"
                }`}
              >
                {quota[key as keyof ResourceQuota]}
              </span>
            </TooltipTrigger>
            <TooltipContent>
              <p className="max-w-xs">{desc.tooltip}</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
    ))}
  </div>
);

const ProviderBadge = ({ provider }: { provider: ClusterInfo["provider"] }) => (
  <span
    className={`px-2 py-1 rounded-full text-xs font-medium ${providerColorMap[provider] || providerColorMap.default}`}
  >
    {provider}
  </span>
);

const HealthBadge = ({ health }: { health: ClusterInfo["health"] }) => {
  const style = healthStatusMap[health.status] || healthStatusMap.default;
  const Icon = style.icon ? icons[style.icon] : null;

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger>
          <div
            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${style.bg} ${style.text}`}
          >
            {Icon && <Icon className="h-3 w-3 mr-1" />}
            {health.status}
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p className="max-w-xs">{health.message}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
};

const ConsoleLink = ({ url }: { url?: string }) => {
  if (!url) return null;

  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="text-blue-500 hover:text-blue-700 flex items-center gap-1"
    >
      <ExternalLink className="h-4 w-4" />
      <span>Console</span>
    </a>
  );
};

const BuildinBadge = () => (
  <span className="px-2 py-1 rounded-full text-xs font-medium bg-purple-100 text-purple-800 ml-2">
    Builtin
  </span>
);

const MonitoringLink = ({
  type,
  url,
}: {
  type: keyof typeof monitoringTypeStyles;
  url?: string;
}) => {
  if (!url) return null;

  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className={`px-2 py-1 text-xs rounded-full transition-colors ${monitoringTypeStyles[type]} hover:underline`}
    >
      {type.charAt(0).toUpperCase() + type.slice(1)}
    </a>
  );
};

const StatsCard = ({
  title,
  value,
  icon: Icon,
  description,
}: {
  title: string;
  value: string | number | React.ReactNode;
  icon: LucideIcon;
  description?: string;
}) => (
  <Card className="bg-gradient-to-br from-white to-gray-50 dark:from-gray-800 dark:to-gray-900 border-gray-200/50 dark:border-gray-700/50 transition-all duration-200 hover:shadow-md">
    <CardContent className="p-6">
      <div className="flex items-center space-x-4">
        <div className="p-3 bg-blue-50 dark:bg-blue-900/20 rounded-xl">
          <Icon className="h-6 w-6 text-blue-600 dark:text-blue-400" />
        </div>
        <div>
          <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
            {title}
          </p>
          <h3 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {value}
          </h3>
          {description && (
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {description}
            </p>
          )}
        </div>
      </div>
    </CardContent>
  </Card>
);

// 修改 ClusterNameCell 组件，移除节点信息
const ClusterNameCell = ({ cluster }: { cluster: ClusterInfo }) => (
  <div
    className={`flex items-center space-x-3 ${
      cluster.status === "disabled" ? "opacity-80 grayscale" : ""
    }`}
  >
    {/* 图标部分 */}
    <div
      className={`p-2 rounded-lg transition-colors duration-200 ${
        cluster.status === "disabled"
          ? "bg-gray-200 text-gray-400 dark:bg-gray-800 dark:text-gray-500"
          : cluster.provider === "GKE"
            ? "bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-400"
            : cluster.provider === "OCP"
              ? "bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-400"
              : cluster.provider === "AKS"
                ? "bg-purple-50 text-purple-600 dark:bg-purple-900/20 dark:text-purple-400"
                : "bg-yellow-50 text-yellow-600 dark:bg-yellow-900/20 dark:text-yellow-400"
      }`}
    >
      <Server className="h-5 w-5" />
    </div>

    <div className="space-y-1">
      <div className="flex items-center space-x-2">
        <span
          className={`font-mono text-base font-bold tracking-wider px-2 py-0.5 rounded transition-colors duration-200 ${
            cluster.status === "disabled"
              ? "bg-gray-200 text-gray-500 dark:bg-gray-800 dark:text-gray-400"
              : "bg-gray-100 dark:bg-gray-800"
          }`}
        >
          {cluster.name}
        </span>
        {cluster.builtin && <BuildinBadge />}
        {cluster.status === "disabled" && (
          <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-600 dark:bg-red-900/20 dark:text-red-400 animate-pulse">
            Disabled
          </span>
        )}
      </div>
      <div className="flex items-center space-x-2 text-xs text-gray-500">
        <span>{cluster.region}</span>
      </div>
    </div>
  </div>
);

export function DestinationCluster() {
  const { clusterList } = useGetClusterList();
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedEnvironment, setSelectedEnvironment] = useState<string>("all");
  const [selectedProvider, setSelectedProvider] = useState<string>("all");
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [selectedCluster, setSelectedCluster] = useState<ClusterInfo | null>(
    null
  );

  const filteredClusters = clusterList.filter((cluster: ClusterInfo) => {
    const matchesSearch =
      cluster.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      cluster.region.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesEnvironment =
      selectedEnvironment === "all" ||
      cluster.environment === selectedEnvironment;
    const matchesProvider =
      selectedProvider === "all" || cluster.provider === selectedProvider;

    return matchesSearch && matchesEnvironment && matchesProvider;
  });

  const handleUpdateResourceQuota = (cluster: ClusterInfo) => {
    setSelectedCluster(cluster);
    setIsDialogOpen(true);
  };

  const stats = {
    totalClusters: clusterList.length,
    activeNodes: clusterList.reduce(
      (acc: number, cluster: ClusterInfo) => acc + cluster.nodes.ready,
      0
    ),
    totalNodes: clusterList.reduce(
      (acc: number, cluster: ClusterInfo) => acc + cluster.nodes.total,
      0
    ),
    healthyClusters: clusterList.filter(
      (c: ClusterInfo) => c.health.status === "Healthy"
    ).length,
  };

  const handleClusterDialog = () => {};

  return (
    <div className="space-y-8">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-gray-100 dark:to-gray-400 bg-clip-text text-transparent">
            Target Clusters
          </h2>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            Manage and monitor your Kubernetes clusters across environments
          </p>
        </div>
        <Button
          onClick={handleClusterDialog}
          className="bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 transition-all duration-200"
        >
          <Plus className="h-4 w-4 mr-2" />
          Add Cluster
        </Button>
      </div>

      <div className="grid grid-cols-4 gap-6">
        <StatsCard
          title="Total Clusters"
          value={stats.totalClusters}
          icon={Server}
          description="Across all environments"
        />
        <StatsCard
          title="Active Nodes"
          value={
            <div className="flex items-baseline space-x-1">
              <span className="text-2xl font-bold">{stats.activeNodes}</span>
              <span className="text-base text-gray-500 dark:text-gray-400">
                /
              </span>
              <span className="text-lg text-gray-600 dark:text-gray-400">
                {stats.totalNodes}
              </span>
              <span className="text-sm text-gray-500 dark:text-gray-400 ml-1">
                nodes
              </span>
            </div>
          }
          icon={Cpu}
          description="Ready nodes vs total nodes"
        />
        <StatsCard
          title="Healthy Clusters"
          value={`${stats.healthyClusters}/${stats.totalClusters}`}
          icon={CheckCircle}
          description="Clusters in healthy state"
        />
        <StatsCard
          title="Environments"
          value={new Set(clusterList.map((c) => c.environment)).size}
          icon={Layout}
          description="Distinct deployment environments"
        />
      </div>

      <Card className="bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm border-gray-200/50 dark:border-gray-700/50">
        <CardContent className="p-6">
          <div className="flex gap-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  placeholder="Search clusters by name or region..."
                  className="pl-10 bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-700"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
            </div>
            <Select
              value={selectedEnvironment}
              onValueChange={setSelectedEnvironment}
            >
              <SelectTrigger className="w-[180px] bg-white dark:bg-gray-900">
                <SelectValue placeholder="Environment" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Environments</SelectItem>
                <SelectItem value="SIT">SIT</SelectItem>
                <SelectItem value="UAT">UAT</SelectItem>
                <SelectItem value="PRD">PRD</SelectItem>
              </SelectContent>
            </Select>
            <Select
              value={selectedProvider}
              onValueChange={setSelectedProvider}
            >
              <SelectTrigger className="w-[180px] bg-white dark:bg-gray-900">
                <SelectValue placeholder="Provider" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Providers</SelectItem>
                <SelectItem value="GKE">GKE</SelectItem>
                <SelectItem value="OCP">OpenShift</SelectItem>
                <SelectItem value="AKS">AKS</SelectItem>
                <SelectItem value="EKS">EKS</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* 集群列表 */}
      <Card className="bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm border-gray-200/50 dark:border-gray-700/50">
        <CardContent className="p-6">
          <div className="rounded-lg overflow-hidden border border-gray-200 dark:border-gray-700">
            <Table>
              <TableHeader>
                <TableRow className="bg-gray-50 dark:bg-gray-900/50">
                  <TableHeadWithTooltip tooltip="Cluster name, region and node information">
                    Name
                  </TableHeadWithTooltip>
                  <TableHeadWithTooltip tooltip="Deployment environment (SIT/UAT/PRD)">
                    Environment
                  </TableHeadWithTooltip>
                  <TableHeadWithTooltip tooltip="Kubernetes platform provider (GKE/OCP/AKS/EKS)">
                    Provider
                  </TableHeadWithTooltip>
                  <TableHeadWithTooltip tooltip="All the core kubernetes components are healthy">
                    Health
                  </TableHeadWithTooltip>
                  <TableHeadWithTooltip tooltip="Number of ready nodes vs total nodes in the cluster">
                    Nodes
                  </TableHeadWithTooltip>
                  <TableHeadWithTooltip tooltip="Current platform version including distribution specific version">
                    Platform Version
                  </TableHeadWithTooltip>
                  <TableHeadWithTooltip tooltip="Resource quotas allocated to your tenant in this cluster. These limits apply to all your tenant's workloads within the cluster">
                    Tenant Resource Quota
                  </TableHeadWithTooltip>
                  <TableHeadWithTooltip tooltip="Available monitoring tools (Prometheus/Grafana)">
                    Monitoring
                  </TableHeadWithTooltip>
                  <TableHeadWithTooltip tooltip="Link to cluster's web console">
                    Console
                  </TableHeadWithTooltip>
                  <TableHeadWithTooltip tooltip="Cluster management actions">
                    Actions
                  </TableHeadWithTooltip>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredClusters.map((cluster) => (
                  <TableRow
                    key={cluster.name}
                    className={`
                      transition-all duration-200
                      ${
                        cluster.status === "disabled"
                          ? "bg-gray-50/80 dark:bg-gray-900/40 hover:bg-gray-100/80 dark:hover:bg-gray-900/60 grayscale"
                          : "hover:bg-gray-50 dark:hover:bg-gray-900/50"
                      }
                    `}
                  >
                    <TableCell>
                      <ClusterNameCell cluster={cluster} />
                    </TableCell>
                    <TableCell>{cluster.environment}</TableCell>
                    <TableCell>
                      <ProviderBadge provider={cluster.provider} />
                    </TableCell>
                    <TableCell>
                      <HealthBadge health={cluster.health} />
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center space-x-2">
                        <div
                          className={`flex items-baseline space-x-1 px-2 py-1 rounded-md ${
                            cluster.status === "disabled"
                              ? "bg-gray-100 dark:bg-gray-800"
                              : cluster.nodes.ready === cluster.nodes.total
                                ? "bg-green-50 dark:bg-green-900/20"
                                : cluster.nodes.ready < cluster.nodes.total / 2
                                  ? "bg-red-50 dark:bg-red-900/20"
                                  : "bg-yellow-50 dark:bg-yellow-900/20"
                          }`}
                        >
                          <span
                            className={`text-sm font-medium ${
                              cluster.status === "disabled"
                                ? "text-gray-500 dark:text-gray-400"
                                : cluster.nodes.ready === cluster.nodes.total
                                  ? "text-green-700 dark:text-green-400"
                                  : cluster.nodes.ready <
                                      cluster.nodes.total / 2
                                    ? "text-red-700 dark:text-red-400"
                                    : "text-yellow-700 dark:text-yellow-400"
                            }`}
                          >
                            {cluster.nodes.ready}
                          </span>
                          <span className="text-gray-400 dark:text-gray-500">
                            /
                          </span>
                          <span className="text-gray-600 dark:text-gray-400">
                            {cluster.nodes.total}
                          </span>
                        </div>
                        <span className="text-xs text-gray-500 dark:text-gray-400">
                          ready
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>{cluster.version.platform}</TableCell>
                    <TableCell>
                      <ResourceQuotaCell
                        quota={cluster.resourceQuota}
                        disabled={cluster.status === "disabled"}
                      />
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        {cluster.monitoring.prometheus &&
                          cluster.monitoring.urls?.prometheus && (
                            <MonitoringLink
                              type="prometheus"
                              url={cluster.monitoring.urls.prometheus}
                            />
                          )}
                        {cluster.monitoring.grafana &&
                          cluster.monitoring.urls?.grafana && (
                            <MonitoringLink
                              type="grafana"
                              url={cluster.monitoring.urls.grafana}
                            />
                          )}
                        {cluster.monitoring.alertmanager &&
                          cluster.monitoring.urls?.alertmanager && (
                            <MonitoringLink
                              type="alertmanager"
                              url={cluster.monitoring.urls.alertmanager}
                            />
                          )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <ConsoleLink url={cluster.consoleUrl} />
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="hover:bg-gray-100 dark:hover:bg-gray-800"
                          >
                            <MoreVertical className="h-4 w-4 text-gray-600 dark:text-gray-400" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-56">
                          <DropdownMenuItem
                            onSelect={() => handleUpdateResourceQuota(cluster)}
                            className={`flex items-center space-x-2 text-sm cursor-pointer ${
                              cluster.status === "disabled"
                                ? "opacity-50 cursor-not-allowed"
                                : ""
                            }`}
                            disabled={cluster.status === "disabled"}
                          >
                            <Settings2 className="h-4 w-4 text-gray-500" />
                            <span>Update Resource Quota</span>
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onSelect={() => {
                              const newStatus =
                                cluster.status === "active"
                                  ? "disabled"
                                  : "active";
                              // 这里添加实际的状态更新逻辑
                              console.log(
                                `Setting cluster ${cluster.name} status to ${newStatus}`
                              );
                            }}
                            className="flex items-center space-x-2 text-sm cursor-pointer"
                          >
                            {cluster.status === "active" ? (
                              <>
                                <XCircle className="h-4 w-4 text-red-500" />
                                <span className="text-red-600">
                                  Disable Cluster
                                </span>
                              </>
                            ) : (
                              <>
                                <CheckCircle className="h-4 w-4 text-green-500" />
                                <span className="text-green-600">
                                  Enable Cluster
                                </span>
                              </>
                            )}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {filteredClusters.length === 0 && (
        <div className="text-center py-12">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-100 dark:bg-gray-800 mb-4">
            <Search className="h-8 w-8 text-gray-400" />
          </div>
          <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
            No clusters found
          </h3>
          <p className="text-gray-500 dark:text-gray-400">
            Try adjusting your search or filter criteria
          </p>
        </div>
      )}

      {/* Update Resource Quota Dialog */}
      {selectedCluster && (
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogContent className="sm:max-w-[425px]">
            <DialogHeader>
              <DialogTitle className="flex items-center space-x-2">
                <Settings2 className="h-5 w-5 text-gray-500" />
                <span>Update Resource Quota</span>
              </DialogTitle>
              <DialogDescription className="text-sm text-gray-500">
                Update resource quotas for cluster{" "}
                <span className="font-mono font-medium">
                  {selectedCluster.name}
                </span>
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              {Object.entries(resourceDescriptions).map(([key, desc]) => {
                const value =
                  selectedCluster.resourceQuota[key as keyof ResourceQuota];
                const numericValue = value.match(/^\d+/)?.[0] || "";
                const unit = value.match(/[A-Za-z]+$/)?.[0] || "";

                return (
                  <div
                    key={key}
                    className="grid grid-cols-4 items-center gap-4"
                  >
                    <Label className="text-right col-span-1">
                      {desc.label}
                    </Label>
                    <div className="col-span-3 flex items-center space-x-2">
                      <Input
                        type="number"
                        defaultValue={numericValue}
                        className="w-24 font-mono"
                      />
                      <span className="text-sm text-gray-500 font-mono">
                        {unit}
                      </span>
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger>
                            <HelpCircle className="h-4 w-4 text-gray-400" />
                          </TooltipTrigger>
                          <TooltipContent>
                            <p className="max-w-xs">{desc.tooltip}</p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                  </div>
                );
              })}
            </div>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsDialogOpen(false)}
                className="mr-2"
              >
                Cancel
              </Button>
              <Button
                onClick={() => {
                  console.log("Resource Quota Updated");
                  setIsDialogOpen(false);
                }}
                className="bg-blue-600 hover:bg-blue-700"
              >
                Save Changes
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
