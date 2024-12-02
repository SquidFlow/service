import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { ApplicationTemplate } from '@/types';
import { getStatusIcon } from './utils';
import { ExternalLink } from "lucide-react";

interface DeploymentStatusProps {
  app: ApplicationTemplate;
}

export function DeploymentStatus({ app }: DeploymentStatusProps) {
  return (
    <div className="grid gap-4 md:grid-cols-3">
      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium">Sync Status</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center space-x-2">
            {getStatusIcon(app.runtime_status.status)}
            <span className="text-2xl font-bold">
              {app.runtime_status.status}
            </span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium">Health Status</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center space-x-2">
            {getStatusIcon(app.runtime_status.health)}
            <span className="text-2xl font-bold">
              {app.runtime_status.health}
            </span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium">Resource Usage</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">CPU</span>
              <span className="font-mono">
                {app.runtime_status.resource_metrics.cpu_cores}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Memory</span>
              <span className="font-mono">
                {app.runtime_status.resource_metrics.memory_usage}
              </span>
            </div>
          </div>
          <div className="mt-4">
            <h3 className="text-sm font-medium text-muted-foreground mb-2">ArgoCD Console</h3>
            <a
              href={app.argocd_url}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center space-x-1 text-blue-500 hover:text-blue-600"
            >
              <ExternalLink className="h-4 w-4" />
              <span>Open in ArgoCD</span>
            </a>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}