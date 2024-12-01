import { Button } from "@/components/ui/button";
import { ChevronLeft } from "lucide-react";
import { DeploymentStatus } from './DeploymentStatus';
import { GeneralInfo } from './GeneralInfo';
import { ResourceMetrics } from './ResourceMetrics';
import { ReleaseHistory } from './ReleaseHistory';
import { ApplicationTemplate } from '@/types/application';

interface ApplicationDetailProps {
  app: ApplicationTemplate;
  onBack: () => void;
}

export function ApplicationDetail({ app, onBack }: ApplicationDetailProps) {
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-gray-100 dark:to-gray-400 bg-clip-text text-transparent">
          {app.name}
        </h1>
        <Button variant="outline" onClick={onBack}>
          <ChevronLeft className="mr-2 h-4 w-4" />
          Back to List
        </Button>
      </div>

      <DeploymentStatus app={app} />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <GeneralInfo app={app} />
        <ResourceMetrics app={app} />
        <ReleaseHistory app={app} />
      </div>
    </div>
  );
}