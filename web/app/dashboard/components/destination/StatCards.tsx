import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ClusterInfo } from '@/types/cluster';

interface StatCardsProps {
  clusters: ClusterInfo[];
}

export function StatCards({ clusters }: StatCardsProps) {
  const nodeStats = clusters.reduce((acc, cluster) => ({
    total: acc.total + cluster.nodes.total,
    ready: acc.ready + cluster.nodes.ready,
  }), { total: 0, ready: 0 });

  const healthyClusters = clusters.filter(cluster => cluster.health.status === 'Healthy').length;
  const totalClusters = clusters.length;

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