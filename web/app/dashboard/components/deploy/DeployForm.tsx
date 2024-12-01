import { useRef, useState } from "react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { DryRun } from "./DryRun";
import { SourceSection } from "./SourceSection";
import { ApplicationSection } from "./ApplicationSection";
import { EnvironmentSection } from "./EnvironmentSection";
import { usePostCreateApplication } from '@/app/api';
import { CreateApplicationPayload } from '@/types/application';
import { DeployFormProvider, useDeployForm } from './DeployFormContext';
import { useToast } from "@/components/ui/use-toast";
import { AlertCircle } from "lucide-react";

interface DeployFormProps {
  onCancel: () => void;
}

function DeployFormContent({ onCancel }: DeployFormProps) {
  const { source, selectedClusters } = useDeployForm();
  const { triggerPostCreateApplication } = usePostCreateApplication();
  const [isDryRunOpen, setIsDryRunOpen] = useState(false);
  const [dryRunYaml, setDryRunYaml] = useState<{ cluster: string; content: string }[]>([]);
  const formRef = useRef<HTMLDivElement>(null);
  const { toast } = useToast();

  const handleSubmit = async () => {
    if (!source.url || !source.path || !source.targetRevision || !source.name || selectedClusters.length === 0) {
      toast({
        variant: "destructive",
        title: "Validation Error",
        description: "Please fill in all required fields",
        duration: 3000,
      });
      return;
    }

    const data: CreateApplicationPayload = {
      name: source.name,
      description: source.description || '',
      tenant_name: source.tenant || '',
      appcode: source.appCode || '',
      source: {
        url: source.url,
        path: source.path,
        targetRevision: source.targetRevision,
        appType: "kustomize",
      },
      destination: {
        clusters: selectedClusters,
      },
      is_dryrun: false,
    };

    try {
      await triggerPostCreateApplication(data);
      onCancel();
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Failed to Create Application",
        description: error instanceof Error ? error.message : "An error occurred",
        duration: 5000,
        action: <AlertCircle className="h-4 w-4" />,
      });
    }
  };

  const handleDryRun = async () => {
    if (!source.url || !source.path || !source.targetRevision || !source.name || selectedClusters.length === 0) {
      toast({
        variant: "destructive",
        title: "Validation Error",
        description: "Please fill in all required fields",
        duration: 3000,
      });
      return;
    }

    const data: CreateApplicationPayload = {
      name: source.name,
      description: source.description || '',
      tenant_name: source.tenant || '',
      appcode: source.appCode || '',
      source: {
        url: source.url,
        path: source.path,
        targetRevision: source.targetRevision,
        appType: "kustomize",
      },
      destination: {
        clusters: selectedClusters,
      },
      is_dryrun: true,
    };

    try {
      const dryRunResult = await triggerPostCreateApplication(data);
      setDryRunYaml(selectedClusters.map(cluster => ({
        cluster,
        content: dryRunResult[cluster] || ''
      })));
      setIsDryRunOpen(true);
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Dry Run Failed",
        description: error instanceof Error ? error.message : "Failed to perform dry run",
        duration: 5000,
        action: <AlertCircle className="h-4 w-4" />,
      });
      console.error("Failed to perform dry run:", error);
    }
  };

  return (
    <div className="flex relative w-full min-h-screen p-6">
      <div
        ref={formRef}
        className="flex space-x-6 transition-transform duration-300 ease-in-out mx-auto"
        style={{ width: isDryRunOpen ? "80%" : "100%" }}
      >
        <div className="flex flex-col space-y-6 w-full max-w-[2400px] mx-auto">
          <SourceSection />
          <ApplicationSection />
          <EnvironmentSection />
          {/* Action Buttons */}
          <Card className="bg-gradient-to-br from-white to-gray-50 dark:from-gray-800 dark:to-gray-900">
            <div className="p-6">
              <div className="flex items-center justify-end space-x-4">
                <Button
                  variant="outline"
                  onClick={onCancel}
                  className="hover:bg-gray-100 dark:hover:bg-gray-700"
                >
                  Cancel
                </Button>
                <Button
                  variant="outline"
                  onClick={handleDryRun}
                  className="hover:bg-blue-50 dark:hover:bg-blue-900/30 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-800"
                  disabled={!selectedClusters.length}
                >
                  Dry Run
                </Button>
                <Button
                  onClick={handleSubmit}
                  className="bg-green-500 hover:bg-green-600 text-white"
                  disabled={!selectedClusters.length}
                >
                  Submit
                </Button>
              </div>
            </div>
          </Card>
        </div>
      </div>

      <DryRun
        isOpen={isDryRunOpen}
        yamls={dryRunYaml}
        onClose={() => setIsDryRunOpen(false)}
      />
    </div>
  );
}

export function DeployForm(props: DeployFormProps) {
  return (
    <DeployFormProvider>
      <DeployFormContent {...props} />
    </DeployFormProvider>
  );
}