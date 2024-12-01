import { useState } from "react";
import { ApplicationList } from './ApplicationList';
import { ApplicationDetail } from './ApplicationDetail';
import { DeployForm } from "../deploy";
import { useGetApplicationDetail } from '@/app/api';
import { ApplicationTemplate } from '@/types/application';

interface ApplicationProps {
  onSelectApp?: (appName: string) => void;
}

export function Application({ onSelectApp }: ApplicationProps) {
  const [isCreating, setIsCreating] = useState(false);
  const [selectedAppDetails, setSelectedAppDetails] = useState<ApplicationTemplate | null>(null);
  const { triggerGetApplicationDetail } = useGetApplicationDetail();

  const handleSelectApp = async (app: ApplicationTemplate) => {
    try {
      const data = await triggerGetApplicationDetail(app.name);
      setSelectedAppDetails(data);
      onSelectApp?.(data.name);
    } catch (error) {}
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
    return <ApplicationDetail app={selectedAppDetails} onBack={() => setSelectedAppDetails(null)} />;
  }

  return <ApplicationList onSelectApp={handleSelectApp} onCreateNew={() => setIsCreating(true)} />;
}