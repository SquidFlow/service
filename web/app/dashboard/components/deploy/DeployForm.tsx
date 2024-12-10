import { useRef, useState } from "react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { DryRun } from "./DryRun";
import { SourceSection } from "./SourceSection";
import { ApplicationSection } from "./ApplicationSection";
import { CreateApplicationPayload } from '@/types/application';
import { DeployFormProvider, useDeployForm } from './DeployFormContext';
import { useToast } from "@/components/ui/use-toast";
import { AlertCircle } from "lucide-react";
import { useApplicationStore } from '@/store';
import { ClusterSelector } from './ClusterSelector';
import { useRouter } from 'next/navigation';

interface DeployFormProps {
  onCancel: () => void;
}

interface DryRunEnvironment {
  environment: string;
  manifest: string;
  is_valid: boolean;
}

function DeployFormContent({ onCancel }: DeployFormProps) {
  const { source, selectedClusters } = useDeployForm();
  const [isDryRunOpen, setIsDryRunOpen] = useState(false);
  const [dryRunYaml, setDryRunYaml] = useState<DryRunEnvironment[]>([]);
  const formRef = useRef<HTMLDivElement>(null);
  const { toast } = useToast();
  const router = useRouter();

  const {
    dryRun,
    isLoading,
    createApplication,
    deploymentStatus
  } = useApplicationStore();

  const handleSubmit = async () => {
    try {
      const payload = {
        application_source: {
          repo: source.url,
          target_revision: source.targetRevision,
          path: source.path,
          submodules: true,
          application_specifier: source.application_specifier
        },
        application_instantiation: {
          application_name: source.name,
          tenant_name: source.tenant || '',
          appcode: source.appCode || '',
          description: source.description || ''
        },
        application_target: selectedClusters.map(cluster => ({
          cluster: cluster,
          namespace: source.namespace || 'default'
        })),
        is_dryrun: false
      };

      await createApplication(payload);

      toast({
        title: "Success",
        description: "Application created successfully",
      });

      router.push('/dashboard/deploy/application');
    } catch (error) {
      console.error('Failed to create application:', error);
      toast({
        variant: "destructive",
        title: "Failed to Create Application",
        description: error instanceof Error ? error.message : "An error occurred while creating the application",
      });
    }
  };

  const handleDryRun = async () => {
    try {
      const payload = {
        application_source: {
          repo: source.url,
          target_revision: source.targetRevision,
          path: source.path,
          submodules: true,
          application_specifier: source.application_specifier
        },
        application_instantiation: {
          application_name: source.name,
          tenant_name: source.tenant || '',
          appcode: source.appCode || '',
          description: source.description || ''
        },
        application_target: selectedClusters.map(cluster => ({
          cluster: cluster,
          namespace: source.namespace || 'default'
        })),
        is_dryrun: true
      };

      const result = await dryRun(payload);

      const yamlResults = result.environments.map(env => ({
        environment: env.environment,
        manifest: env.manifest,
        is_valid: env.is_valid
      }));

      setDryRunYaml(yamlResults);
      setIsDryRunOpen(true);
    } catch (error) {
      console.error('Dry run failed:', error);
      toast({
        variant: "destructive",
        title: "Dry Run Failed",
        description: error instanceof Error ? error.message : "Failed to perform dry run",
      });
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
          <ClusterSelector />
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