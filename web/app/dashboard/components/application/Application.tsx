import { useState, useEffect } from "react";
import { ApplicationList } from './ApplicationList';
import { ApplicationDetail } from './ApplicationDetail';
import { DeployForm } from "../deploy";
import { useApplicationStore } from '@/store';
import { ApplicationTemplate } from '@/types/application';

interface ApplicationProps {
  onSelectApp?: (appName: string) => void;
}

export function Application({ onSelectApp }: ApplicationProps) {
  const [isCreating, setIsCreating] = useState(false);
  const [selectedAppDetails, setSelectedAppDetails] = useState<ApplicationTemplate | null>(null);
  const {
    data: applications,
    isLoading,
    error,
    fetch: fetchApplications,
    getApplicationDetail
  } = useApplicationStore();

  useEffect(() => {
    fetchApplications();
  }, [fetchApplications]);

  const handleSelectApp = async (app: ApplicationTemplate) => {
    try {
      if (!app?.name) {
        console.error('Application name is missing');
        return;
      }

      const data = await getApplicationDetail(app.name);
      if (!data) {
        console.error('Failed to get application details: No data returned');
        return;
      }

      setSelectedAppDetails(data);
      onSelectApp?.(app.name);
    } catch (error) {
      console.error('Failed to get application details:', error);
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
    return <ApplicationDetail app={selectedAppDetails} onBack={() => setSelectedAppDetails(null)} />;
  }

  return <ApplicationList onSelectApp={handleSelectApp} onCreateNew={() => setIsCreating(true)} />;
}