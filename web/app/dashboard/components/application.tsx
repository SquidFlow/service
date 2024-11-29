"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import {
  AlertCircle,
  Clock,
  CheckCircle,
  XCircle,
  ChevronLeft,
  FileText,
  BarChart,
  History,
  ExternalLink,
  Pause,
} from "lucide-react";
import { Trash2, RefreshCw, Plus } from "lucide-react";
import { DeployForm } from "@/app/dashboard/components/deploy";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { ChevronRight } from "lucide-react";
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
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { MoreHorizontal } from "lucide-react";
import { useGetApplicationDetail, useGetApplicationList, useDeleteApplications, useGetReleaseHistories } from "@/app/api";
import { useToast } from "@/components/ui/use-toast";
import { ApplicationTemplate } from "@/types";
import { EnvironmentType, ReleaseHistory } from "@/types/release";

const getStatusIcon = (status: string) => {
  switch (status) {
    case "Succeeded":
    case "Synced":
      return <CheckCircle className="h-4 w-4 text-green-500" />;
    case "OutOfSync":
      return <AlertCircle className="h-4 w-4 text-yellow-500" />;
    case "Progressing":
      return <Clock className="h-4 w-4 text-blue-500" />;
    case "Degraded":
      return <XCircle className="h-4 w-4 text-red-500" />;
    default:
      return <AlertCircle className="h-4 w-4 text-gray-500" />;
  }
};

const getHealthIcon = (health: string) => {
  switch (health) {
    case "Healthy":
      return <CheckCircle className="h-4 w-4 text-green-500" />;
    case "Degraded":
      return <XCircle className="h-4 w-4 text-red-500" />;
    case "Progressing":
      return <Clock className="h-4 w-4 text-blue-500" />;
    case "Suspended":
      return <Pause className="h-4 w-4 text-yellow-500" />;
    default:
      return <AlertCircle className="h-4 w-4 text-gray-500" />;
  }
};

const ArgoLink = ({ url }: { url: string }) => (
  <a
    href={url}
    target="_blank"
    rel="noopener noreferrer"
    className="inline-flex items-center px-3 py-2 rounded-md text-base font-medium
    bg-blue-50 text-blue-600 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400
    dark:hover:bg-blue-900/30 transition-colors duration-200 group"
  >
    <ExternalLink className="h-5 w-5 mr-2 group-hover:translate-x-0.5 transition-transform duration-200" />
    ArgoCD Console
  </a>
);

interface ArgoApplicationProps {
  onSelectApp?: (appName: string) => void;
}

export function Application({ onSelectApp }: ArgoApplicationProps) {
  const { applications, mutate: refreshApplications } = useGetApplicationList();
  const [selectedApps, setSelectedApps] = useState<string[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const { triggerGetApplicationDetail } = useGetApplicationDetail();
  const [selectedAppDetails, setSelectedAppDetails] =
    useState<ApplicationTemplate | null>(null);
  const { toast } = useToast();
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [deleteConfirmationInput, setDeleteConfirmationInput] = useState("");
  const { deleteApplications, isLoading: isDeleting } = useDeleteApplications();
  const { releaseHistories: releaseHistoriesData } = useGetReleaseHistories();

  const [currentCommits, setCurrentCommits] = useState<Record<EnvironmentType, string>>({
    SIT: releaseHistoriesData.SIT[0]?.commitHash || "",
    UAT: releaseHistoriesData.UAT[0]?.commitHash || "",
    PRD: releaseHistoriesData.PRD[0]?.commitHash || "",
  });

  const handleRollback = (env: EnvironmentType, commitHash: string) => {
    setCurrentCommits((prev) => ({
      ...prev,
      [env]: commitHash,
    }));

    const updatedHistories = { ...releaseHistoriesData };
    updatedHistories[env] = releaseHistoriesData[env].map((release: ReleaseHistory) => ({
      ...release,
      isCurrent: release.commitHash === commitHash,
    }));

    console.log(`Rolling back to ${commitHash} in ${env}`);
  };

  const renderApplicationDetail = (app: ApplicationTemplate) => {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-gray-100 dark:to-gray-400 bg-clip-text text-transparent">
            {app.name}
          </h1>
          <Button variant="outline" onClick={() => setSelectedAppDetails(null)}>
            <ChevronLeft className="mr-2 h-4 w-4" />
            Back to List
          </Button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Card className="col-span-3 bg-white dark:bg-gray-800 shadow-sm hover:shadow-md transition-shadow duration-200">
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <FileText className="h-5 w-5 text-blue-500" />
                <span>ArgoCD Deployment Status</span>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                <div className="space-y-2">
                  <p className="text-sm font-medium text-gray-500">
                    Sync Status
                  </p>
                  <div className="flex items-center space-x-2">
                    {getStatusIcon(app.runtime_status.status)}
                    <span className="font-medium">
                      {app.runtime_status.status}
                    </span>
                  </div>
                </div>

                {/* Health Status */}
                <div className="space-y-2">
                  <p className="text-sm font-medium text-gray-500">
                    Health Status
                  </p>
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <div className="flex items-center space-x-2">
                          {getHealthIcon(app.runtime_status.health)}
                          <span className="font-medium">
                            {app.runtime_status.health}
                          </span>
                        </div>
                      </TooltipTrigger>
                      {/* <TooltipContent>
                        <p>{app.health.message}</p>
                      </TooltipContent> */}
                    </Tooltip>
                  </TooltipProvider>
                </div>

                {/* Deployed Environments */}
                <div className="space-y-2">
                  <p className="text-sm font-medium text-gray-500">
                    Deployed Environments
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {app.deployed_environments?.map((env) => (
                      <span
                        key={env}
                        className="px-2 py-1 text-sm rounded-full bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300"
                      >
                        {env}
                      </span>
                    ))}
                  </div>
                </div>

                {/* ArgoCD Link */}
                <div className="space-y-2">
                  <p className="text-sm font-medium text-gray-500">
                    ArgoCD Console
                  </p>
                  <a
                    href={app.argocdUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center space-x-2 text-blue-500 hover:text-blue-600 hover:underline"
                  >
                    <ExternalLink className="h-4 w-4" />
                    <span>View in ArgoCD</span>
                  </a>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* General Information Card - 保持原有的仓库信息 */}
          <Card className="col-span-2 bg-white dark:bg-gray-800 shadow-sm hover:shadow-md transition-shadow duration-200">
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <FileText className="h-5 w-5 text-blue-500" />
                <span>General Information</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Basic Info */}
              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-4">
                  <div>
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Owner
                    </p>
                    <div className="flex items-center space-x-2 mt-1">
                      {/* <Avatar className="h-6 w-6">
                        <AvatarFallback className="bg-blue-100 text-blue-600">
                          {app.owner
                            .split(" ")
                            .map((n) => n[0])
                            .join("")}
                        </AvatarFallback>
                      </Avatar> */}
                      <span className="font-medium">{app.owner}</span>
                    </div>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Last Update
                    </p>
                    <p className="mt-1">
                      {new Date(app.lastUpdate).toLocaleString()}
                    </p>
                  </div>
                </div>
              </div>

              {/* Remote Repository Section */}
              <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
                <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                  Remote Repository
                </h4>
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-500">
                      Repository URL
                    </span>
                    <a
                      href={app.remoteRepo?.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-blue-500 hover:text-blue-600 font-mono"
                    >
                      {app.remoteRepo?.url}
                    </a>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-500">Branch</span>
                    <span className="text-sm font-medium">
                      {app.remoteRepo?.branch}
                    </span>
                  </div>
                  <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-3 space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-gray-500">Latest Commit</span>
                      <code className="px-2 py-1 bg-gray-100 dark:bg-gray-900 rounded text-xs">
                        {app.remoteRepo?.latestCommit.id}
                      </code>
                    </div>
                    <p className="text-sm">
                      {app.remoteRepo?.latestCommit.message}
                    </p>
                    <div className="flex items-center justify-between text-xs text-gray-500">
                      <span>{app.remoteRepo?.latestCommit.author}</span>
                      <time>
                        {new Date(
                          app.remoteRepo?.latestCommit.timestamp
                        ).toLocaleString()}
                      </time>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Resource Metrics Card */}
          <Card className="bg-white dark:bg-gray-800 shadow-sm hover:shadow-md transition-shadow duration-200">
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <BarChart className="h-5 w-5 text-purple-500" />
                <span>Resource Metrics</span>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Tabs defaultValue={app.deployed_environments?.[0]}>
                <TabsList className="grid w-full grid-cols-3 mb-6">
                  {app.deployed_environments?.map((env) => (
                    <TabsTrigger key={env} value={env}>
                      {env}
                    </TabsTrigger>
                  ))}
                </TabsList>

                {app.deployed_environments?.map((env) => (
                  <TabsContent key={env} value={env} className="space-y-6">
                    <div className="space-y-2">
                      <div className="flex justify-between items-center text-sm">
                        <span className="text-gray-500 dark:text-gray-400">
                          CPU Usage
                        </span>
                        <span className="font-medium">
                          {app.resources[env].cpu}
                        </span>
                      </div>
                      <div className="h-2 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-purple-500 rounded-full"
                          style={{ width: "60%" }}
                        />
                      </div>
                    </div>
                    <div className="space-y-2">
                      <div className="flex justify-between items-center text-sm">
                        <span className="text-gray-500 dark:text-gray-400">
                          Memory Usage
                        </span>
                        <span className="font-medium">
                          {app.resources[env].memory}
                        </span>
                      </div>
                      <div className="h-2 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-blue-500 rounded-full"
                          style={{ width: "45%" }}
                        />
                      </div>
                    </div>
                    <div className="space-y-2">
                      <div className="flex justify-between items-center text-sm">
                        <span className="text-gray-500 dark:text-gray-400">
                          Storage Usage
                        </span>
                        <span className="font-medium">
                          {app.resources[env].storage}
                        </span>
                      </div>
                      <div className="h-2 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-green-500 rounded-full"
                          style={{ width: "30%" }}
                        />
                      </div>
                    </div>
                    <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
                      <div className="grid grid-cols-2 gap-4">
                        <div className="text-center p-3 bg-gray-50 dark:bg-gray-900 rounded-lg">
                          <p className="text-2xl font-bold text-blue-500">
                            {app.resources[env].pods}
                          </p>
                          <p className="text-sm text-gray-500">Active Pods</p>
                        </div>
                        <div className="text-center p-3 bg-gray-50 dark:bg-gray-900 rounded-lg">
                          <p className="text-2xl font-bold text-green-500">
                            {app.secretCount}
                          </p>
                          <p className="text-sm text-gray-500">Secrets</p>
                        </div>
                      </div>
                    </div>
                  </TabsContent>
                ))}
              </Tabs>
            </CardContent>
          </Card>

          {/* Release History Card */}
          <Card className="col-span-3 bg-white dark:bg-gray-800 shadow-sm hover:shadow-md transition-shadow duration-200">
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <History className="h-5 w-5 text-green-500" />
                <span>Release History and Rollback</span>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow className="bg-gray-50 dark:bg-gray-800">
                    <TableHead>Commit Log</TableHead>
                    <TableHead>Commit Hash</TableHead>
                    <TableHead>Commit Author</TableHead>
                    <TableHead>Operator</TableHead>
                    <TableHead>Release Date</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="w-[100px]">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {releaseHistoriesData.SIT.map((release: ReleaseHistory, index: number) => (
                    <TableRow
                      key={index}
                      className="hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
                    >
                      <TableCell>
                        <div className="space-y-2">
                          <div className="font-medium">{release.commitLog}</div>
                          <div className="flex gap-2">
                            {releaseHistoriesData.SIT.find(
                              (r: ReleaseHistory) => r.commitHash === release.commitHash
                            )?.isCurrent && (
                              <Badge
                                variant="outline"
                                className="bg-blue-50 text-blue-600 dark:bg-blue-900/20 font-medium text-xs"
                              >
                                @SIT0
                              </Badge>
                            )}
                            {/* UAT 环境状态 */}
                            {releaseHistoriesData.UAT.find(
                              (r: ReleaseHistory) => r.commitHash === release.commitHash
                            )?.isCurrent && (
                              <Badge
                                variant="outline"
                                className="bg-purple-50 text-purple-600 dark:bg-purple-900/20 font-medium text-xs"
                              >
                                @UAT
                              </Badge>
                            )}
                            {/* PRD 环境状态 */}
                            {releaseHistoriesData.PRD.find(
                              (r: ReleaseHistory) => r.commitHash === release.commitHash
                            )?.isCurrent && (
                              <Badge
                                variant="outline"
                                className="bg-green-50 text-green-600 dark:bg-green-900/20 font-medium text-xs"
                              >
                                @PDC
                              </Badge>
                            )}
                            {releaseHistoriesData.PRD.find(
                              (r: ReleaseHistory) => r.commitHash === release.commitHash
                            )?.isCurrent && (
                              <Badge
                                variant="outline"
                                className="bg-green-50 text-green-600 dark:bg-green-900/20 font-medium text-xs"
                              >
                                @DDC
                              </Badge>
                            )}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <a
                          href={`${app.remoteRepo?.baseCommitUrl}/${release.commitHash}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="group flex items-center space-x-1 text-sm"
                        >
                          <code className="px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded font-mono text-xs">
                            {release.commitHash.substring(0, 7)}
                          </code>
                          <ExternalLink className="h-3 w-3 text-gray-400 group-hover:text-blue-600 opacity-0 group-hover:opacity-100 transition-all" />
                        </a>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center space-x-2">
                          <Avatar className="h-6 w-6">
                            <AvatarFallback className="bg-blue-100 text-blue-600 text-xs">
                              {release.commitAuthor
                                .split(" ")
                                .map((n) => n[0])
                                .join("")}
                            </AvatarFallback>
                          </Avatar>
                          <span>{release.commitAuthor}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center space-x-2">
                          <Avatar className="h-6 w-6">
                            <AvatarFallback className="bg-purple-100 text-purple-600 text-xs">
                              {release.operator
                                .split(" ")
                                .map((n) => n[0])
                                .join("")}
                            </AvatarFallback>
                          </Avatar>
                          <span>{release.operator}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <time className="text-gray-500">
                          {new Date(release.releaseDate).toLocaleString()}
                        </time>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center space-x-2">
                          {release.status === "success" ? (
                            <CheckCircle className="h-4 w-4 text-green-500" />
                          ) : release.status === "failed" ? (
                            <XCircle className="h-4 w-4 text-red-500" />
                          ) : (
                            <Clock className="h-4 w-4 text-blue-500" />
                          )}
                          <span
                            className={
                              release.status === "success"
                                ? "text-green-600"
                                : release.status === "failed"
                                  ? "text-red-600"
                                  : "text-blue-600"
                            }
                          >
                            {release.status.charAt(0).toUpperCase() +
                              release.status.slice(1)}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="w-8 h-8 p-0"
                            >
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() => {
                                console.log(`Redeploying in SIT`);
                              }}
                            >
                              <RefreshCw className="h-4 w-4 mr-2" />
                              Redeploy
                            </DropdownMenuItem>

                            {releaseHistoriesData.SIT.includes(release) && (
                              <DropdownMenuItem
                                onClick={() => {
                                  console.log("Promoting to UAT");
                                }}
                              >
                                <ChevronRight className="h-4 w-4 mr-2" />
                                Promote to UAT
                              </DropdownMenuItem>
                            )}

                            {releaseHistoriesData.UAT.includes(release) && (
                              <DropdownMenuItem
                                onClick={() => {
                                  console.log("Promoting to PRD");
                                }}
                              >
                                <ChevronRight className="h-4 w-4 mr-2" />
                                Promote to PRD
                              </DropdownMenuItem>
                            )}

                            {!release.isCurrent && (
                              <DropdownMenuItem
                                onClick={() => {
                                  if (
                                    release.commitHash !== currentCommits["SIT"]
                                  ) {
                                    handleRollback("SIT", release.commitHash);
                                  }
                                }}
                              >
                                <History className="h-4 w-4 mr-2" />
                                Rollback to this version
                              </DropdownMenuItem>
                            )}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  };
  const filteredApps = applications.filter(
    (app) =>
      app.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      app.owner.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleSelectApp = async (app: ApplicationTemplate) => {
    try {
      const data = await triggerGetApplicationDetail(app.name);
      setSelectedAppDetails(data);
      onSelectApp?.(data.name);
    } catch (error) {}
  };

  const handleDeleteApplications = async () => {
    try {
      if (deleteConfirmationInput !== selectedApps.join(", ")) {
        toast({
          title: "Error",
          description: "Please enter the correct application name(s) to confirm deletion.",
          variant: "destructive",
        });
        return;
      }

      await deleteApplications(selectedApps);

      toast({
        title: "Success",
        description: `Successfully deleted ${selectedApps.length} application(s)`,
      });

      // Clear selection and refresh the list
      setSelectedApps([]);
      setIsDeleteDialogOpen(false);
      setDeleteConfirmationInput("");
      refreshApplications();
    } catch (error) {
      console.error('Error deleting applications:', error);
      toast({
        title: "Error",
        description: "Failed to delete applications. Please try again.",
        variant: "destructive",
      });
    }
  };

  if (isCreating) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen bg-background">
        <div className="w-full max-w-4xl p-6 space-y-4">
          <DeployForm onCancel={() => setIsCreating(false)} />
        </div>
      </div>
    );
  }

  if (selectedAppDetails) {
    return renderApplicationDetail(selectedAppDetails);
  }

  const mainListView = (
    <div className="space-y-6">
      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Application(s)</DialogTitle>
            <DialogDescription>
              This action cannot be undone. To confirm deletion, please type the name(s) of the application(s):
              <span className="font-mono text-sm mt-2 block bg-muted p-2 rounded">
                {selectedApps.join(", ")}
              </span>
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Input
              placeholder="Enter application name(s) to confirm"
              value={deleteConfirmationInput}
              onChange={(e) => setDeleteConfirmationInput(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setIsDeleteDialogOpen(false);
                setDeleteConfirmationInput("");
              }}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteApplications}
              disabled={!deleteConfirmationInput || isDeleting}
            >
              {isDeleting ? (
                <>
                  <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                "Delete"
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 顶部操作栏 */}
      <div className="flex justify-between items-center">
        <div className="flex items-center space-x-4 flex-1">
          <Input
            placeholder="Search applications..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="max-w-sm text-sm font-medium"
          />
          <Button
            variant="outline"
            size="icon"
            onClick={() => console.log("Refresh")}
          >
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
        <div className="flex items-center space-x-4">
          {/* Add these new buttons */}
          {selectedApps.length > 0 && (
            <>
              <Button
                variant="outline"
                className="text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200"
                onClick={() => setIsDeleteDialogOpen(true)}
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Delete Selected
              </Button>
              <Button
                variant="outline"
                className="text-blue-600 hover:text-blue-700 hover:bg-blue-50 border-blue-200"
                onClick={() => {
                  console.log("Syncing apps:", selectedApps);
                  // Add your sync logic here
                }}
              >
                <RefreshCw className="h-4 w-4 mr-2" />
                Sync Selected
              </Button>
            </>
          )}
          <Button onClick={() => setIsCreating(true)}>
            <Plus className="h-4 w-4 mr-2" />
            New Application
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base font-medium text-gray-900 dark:text-gray-100">
            Applications
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow className="bg-muted/50">
                <TableHead className="w-12">
                  {filteredApps.length > 0 && (
                    <Checkbox
                      checked={selectedApps.length === filteredApps.length && filteredApps.length > 0}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          setSelectedApps(filteredApps.map((app) => app.name));
                        } else {
                          setSelectedApps([]);
                        }
                      }}
                    />
                  )}
                </TableHead>
                <TableHead className="text-base font-semibold">Name</TableHead>
                <TableHead className="text-base font-semibold">
                  Status
                </TableHead>
                <TableHead className="text-base font-semibold">
                  Health
                </TableHead>
                <TableHead className="text-base font-semibold">
                  Repository
                </TableHead>
                <TableHead className="text-base font-semibold">
                  Environments
                </TableHead>
                <TableHead className="text-base font-semibold">
                  Last Updated
                </TableHead>
                <TableHead className="text-base font-semibold text-right">
                  ArgoCD
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredApps.map((app) => (
                <TableRow
                  key={app.id}
                  className="hover:bg-muted/50 transition-colors duration-200"
                >
                  <TableCell className="py-4">
                    <Checkbox
                      checked={selectedApps.includes(app.name)}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          setSelectedApps((prev) => [...prev, app.name]);
                        } else {
                          setSelectedApps((prev) => prev.filter((name) => name !== app.name));
                        }
                      }}
                    />
                  </TableCell>
                  <TableCell className="py-4">
                    <Button
                      variant="link"
                      className="p-0 h-auto text-sm font-semibold hover:text-blue-600"
                      onClick={() => handleSelectApp(app)}
                    >
                      <span>{app.name}</span>
                    </Button>
                  </TableCell>
                  <TableCell className="py-4">
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(app.runtime_status.status)}
                      <span className="text-sm">
                        {app.runtime_status.status}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell className="py-4">
                    <Badge
                      variant={
                        app.runtime_status.health === "Healthy"
                          ? "default"
                          : app.runtime_status.health === "Degraded"
                            ? "destructive"
                            : app.runtime_status.health === "Progressing"
                              ? "secondary"
                              : "outline"
                      }
                      className="text-sm px-3 py-1"
                    >
                      <div className="flex items-center space-x-2">
                        {getHealthIcon(app.runtime_status.health)}
                        <span>{app.runtime_status.health}</span>
                      </div>
                    </Badge>
                  </TableCell>
                  <TableCell className="py-4">
                    <div className="flex items-center space-x-2">
                      <code className="px-3 py-1.5 bg-muted rounded text-sm">
                        {app.template?.source?.url}
                      </code>
                    </div>
                  </TableCell>
                  <TableCell className="py-4">
                    <div className="flex flex-wrap gap-1.5">
                      {app.destination_clusters?.clusters.map((env: string) => (
                        <Badge
                          key={env}
                          variant="outline"
                          className="text-sm px-3 py-1"
                        >
                          {env}
                        </Badge>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell className="py-4">
                    <time className="text-sm text-muted-foreground">
                      {new Date().toLocaleDateString()}
                    </time>
                  </TableCell>
                  <TableCell className="text-right py-4">
                    <ArgoLink url={app.argocdUrl} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );

  return mainListView;
}
