import { useEffect } from 'react';
import { Card, CardHeader, CardContent, CardTitle } from "@/components/ui/card";
import { Server, CheckCircle2, AlertTriangle, XCircle } from "lucide-react";
import { useClusterStore } from '@/store';
import { useDeployForm } from './DeployFormContext';
import { Button } from "@/components/ui/button";
import { ClusterInfo } from '@/types/cluster';
import { useToast } from "@/components/ui/use-toast";
import { cn } from '@/lib/utils';

interface ClusterTagProps {
  cluster: ClusterInfo;
  isSelected: boolean;
  onClick: () => void;
}

export function ClusterTag({ cluster, isSelected, onClick }: ClusterTagProps) {
  const getEnvironmentStyle = (env: string) => {
    switch (env.toUpperCase()) {
      case 'SIT':
        return 'border-blue-500 bg-blue-50 hover:bg-blue-100';
      case 'UAT':
        return 'border-green-500 bg-green-50 hover:bg-green-100';
      case 'PRD':
        return 'border-purple-500 bg-purple-50 hover:bg-purple-100';
      default:
        return 'border-gray-500 bg-gray-50 hover:bg-gray-100';
    }
  };

  const getBuiltinBadge = () => {
    if (!cluster.builtin) return null;
    return (
      <span className="absolute -top-2 -right-2 px-1.5 py-0.5 text-[10px] font-medium bg-blue-500 text-white rounded-full">
        Builtin
      </span>
    );
  };

  return (
    <button
      onClick={onClick}
      className={cn(
        'relative px-4 py-2 rounded-md border-2 transition-all',
        getEnvironmentStyle(cluster.environment),
        isSelected && 'ring-2 ring-blue-500 ring-offset-2',
        'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
      )}
    >
      {getBuiltinBadge()}
      <div className="flex items-center space-x-2">
        <span className="font-medium">{cluster.name}</span>
        <span className="text-xs uppercase">
          {cluster.environment || 'Default'}
        </span>
      </div>
    </button>
  );
}

export function ClusterSelector() {
  const { data: clusters, getClusterList, isLoading, error } = useClusterStore();
  const { selectedClusters, setSelectedClusters, setClusterDetails } = useDeployForm();
  const { toast } = useToast();

  useEffect(() => {
    const fetchClusters = async () => {
      try {
        await getClusterList();
      } catch (err) {
        console.error('Failed to fetch clusters:', err);
        toast({
          variant: "destructive",
          title: "Error",
          description: "Failed to load clusters",
        });
      }
    };

    fetchClusters();
    console.log('ClusterSelector mounted');
  }, [getClusterList, toast]);

  useEffect(() => {
    console.log('Clusters data:', clusters);
  }, [clusters]);

  const handleClusterSelect = (cluster: ClusterInfo) => {
    setSelectedClusters((prev) => {
      const newSelected = prev.includes(cluster.name)
        ? prev.filter((name) => name !== cluster.name)
        : [...prev, cluster.name];

      setClusterDetails(
        clusters.filter(c => newSelected.includes(c.name))
      );

      return newSelected;
    });
  };

  const getEnvironmentStyle = (env: string) => {
    switch (env.toUpperCase()) {
      case "SIT":
        return "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300";
      case "UAT":
        return "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300";
      case "PRD":
        return "bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300";
      default:
        return "bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-300";
    }
  };

  const getHealthIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case "successful":
      case "healthy":
        return <CheckCircle2 className="h-4 w-4 text-green-500" />;
      case "degraded":
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      default:
        return <XCircle className="h-4 w-4 text-red-500" />;
    }
  };

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <Server className="h-6 w-6 text-blue-500" />
            <span>Destination Clusters</span>
          </div>
          <span className="text-sm text-muted-foreground">
            {selectedClusters.length} cluster(s) selected
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="text-center py-4 text-muted-foreground">
            Loading clusters...
          </div>
        ) : error ? (
          <div className="text-center py-4 text-red-500">
            Failed to load clusters: {error.message}
          </div>
        ) : clusters.length === 0 ? (
          <div className="text-center py-4 text-muted-foreground">
            No clusters available
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {clusters.map((cluster) => (
              <Button
                key={cluster.name}
                variant={selectedClusters.includes(cluster.name) ? "default" : "outline"}
                className="h-auto p-4 w-full flex flex-col items-start gap-2"
                onClick={() => handleClusterSelect(cluster)}
              >
                <div className="flex items-center justify-between w-full">
                  <div className="flex items-center gap-2">
                    <Server className="h-4 w-4" />
                    <span className="font-medium">{cluster.name}</span>
                    {cluster.builtin && (
                      <span className="px-1.5 py-0.5 text-[10px] font-medium bg-blue-500 text-white rounded-full">
                        Builtin
                      </span>
                    )}
                  </div>
                  {getHealthIcon(cluster.health.status)}
                </div>

                <div className="flex flex-wrap gap-2 w-full">
                  <span className={`text-xs px-2 py-0.5 rounded-full ${getEnvironmentStyle(cluster.environment)}`}>
                    {cluster.environment || 'Default'}
                  </span>
                </div>
              </Button>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}