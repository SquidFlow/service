import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ClusterInfo } from '@/types/cluster';

interface StatCardsProps {
  clusters: ClusterInfo[];
}

export function StatCards({ clusters = [] }: StatCardsProps) {
  const safeClusters = Array.isArray(clusters) ? clusters : [];

  const safeGetNodeCount = (cluster: ClusterInfo | null | undefined, type: 'total' | 'ready'): number => {
    try {
      if (!cluster || !cluster.nodes) return 0;
      const value = type === 'total' ? cluster.nodes.total : cluster.nodes.ready;
      return typeof value === 'number' && !isNaN(value) ? value : 0;
    } catch (error) {
      console.error('Error getting node count:', error);
      return 0;
    }
  };

  const nodeStats = (() => {
    try {
      return safeClusters.reduce((acc, cluster) => ({
        total: acc.total + safeGetNodeCount(cluster, 'total'),
        ready: acc.ready + safeGetNodeCount(cluster, 'ready')
      }), { total: 0, ready: 0 });
    } catch (error) {
      console.error('Error calculating node stats:', error);
      return { total: 0, ready: 0 };
    }
  })();

  const healthyClusters = (() => {
    try {
      return safeClusters.filter(cluster =>
        cluster && cluster.status === 'active'
      ).length;
    } catch (error) {
      console.error('Error calculating healthy clusters:', error);
      return 0;
    }
  })();

  const totalClusters = safeClusters.length;

  const stats = [
    {
      title: "Total Clusters",
      value: totalClusters,
      description: "Number of registered clusters"
    },
    {
      title: "Healthy Clusters",
      value: healthyClusters,
      description: `${healthyClusters}/${totalClusters} clusters are healthy`
    },
    {
      title: "Node Status",
      value: `${nodeStats.ready}/${nodeStats.total}`,
      description: "Ready nodes across all clusters"
    }
  ];

  return (
    <div className="grid gap-4 md:grid-cols-3">
      {stats.map((stat) => (
        <Card key={stat.title}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {stat.title}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stat.value}</div>
            <p className="text-xs text-muted-foreground">
              {stat.description}
            </p>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}