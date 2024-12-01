import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ApplicationTemplate } from '@/types/application';
import { getStatusIcon } from './utils';

interface DeploymentStatusProps {
  app: ApplicationTemplate;
}

export function DeploymentStatus({ app }: DeploymentStatusProps) {
  return (
    <Card className="col-span-3 bg-card">
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <svg className="h-5 w-5 text-blue-500" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
          </svg>
          <span>ArgoCD Deployment Status</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-4 gap-6">
          <div>
            <h3 className="text-sm font-medium text-muted-foreground mb-2">Sync Status</h3>
            <div className="flex items-center space-x-2">
              {getStatusIcon(app.runtime_status.status)}
              <span>{app.runtime_status.status}</span>
            </div>
          </div>
          <div>
            <h3 className="text-sm font-medium text-muted-foreground mb-2">Health Status</h3>
            <div className="flex items-center space-x-2">
              <div className="p-0.5 bg-green-500 rounded-full">
                <svg className="h-3 w-3 text-white" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M20 6 9 17l-5-5"/>
                </svg>
              </div>
              <span>{app.runtime_status.health}</span>
            </div>
          </div>
          <div>
            <h3 className="text-sm font-medium text-muted-foreground mb-2">Deployed Environments</h3>
            <div className="flex items-center space-x-2">
              {app.deployed_environments?.map((env) => (
                <span key={env} className={`px-2 py-1 text-xs rounded-full ${
                  env === 'SIT' ? 'bg-blue-100 text-blue-700' :
                  env === 'UAT' ? 'bg-green-100 text-green-700' :
                  'bg-purple-100 text-purple-700'
                }`}>
                  {env}
                </span>
              ))}
            </div>
          </div>
          <div>
            <h3 className="text-sm font-medium text-muted-foreground mb-2">ArgoCD Console</h3>
            <a
              href={app.argocdUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center space-x-1 text-blue-500 hover:text-blue-600"
            >
              <span>View in ArgoCD</span>
              <svg className="h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
                <polyline points="15 3 21 3 21 9"/>
                <line x1="10" y1="14" x2="21" y2="3"/>
              </svg>
            </a>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}