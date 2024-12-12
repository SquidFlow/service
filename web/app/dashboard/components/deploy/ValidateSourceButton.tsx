import { useState } from "react";
import { Button } from "@/components/ui/button";
import { CheckCircle, FileText } from "lucide-react";
import { useToast } from "@/components/ui/use-toast";
import { AlertCircle, CopyIcon, CheckIcon } from "lucide-react";
import { useApplicationStore } from '@/store';

interface ErrorResponse {
  success: boolean;
  message: string;
  type: string;
  suiteable_env?: Array<{
    environments: string;
    valid: boolean;
    error: string;
    manifest?: string;
  }>;
}

interface ValidationResult {
  success: boolean;
  message: string;
  type: string;
  suiteable_env?: Array<{
    environments: string;
    valid: boolean;
    error?: string;
    manifest?: string;
  }>;
}

interface ValidateSourceButtonProps {
  isValid: boolean;
  source: {
    url: string;
    path: string;
    targetRevision: string;
    application_specifier?: {
      helm_manifest_path?: string;
    };
  };
  onServiceListUpdate?: (services: string[]) => void;
}

interface SuiteableEnv {
  environments: string;
  valid: boolean;
  error?: string;
  manifest?: string;
}

interface ResourceCount {
  deployments: number;
  services: number;
  configmaps: number;
  secrets: number;
  others: number;
}

export function ValidateSourceButton({ isValid, source, onServiceListUpdate }: ValidateSourceButtonProps) {
  const [isValidating, setIsValidating] = useState(false);
  const [isCopied, setIsCopied] = useState(false);
  const { validateApplication } = useApplicationStore();
  const { toast } = useToast();

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setIsCopied(true);
      setTimeout(() => setIsCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy text:', err);
    }
  };

  const calculateResourceCount = (manifest: string): ResourceCount => {
    const resources = manifest.split('---').filter(Boolean);
    const count: ResourceCount = {
      deployments: 0,
      services: 0,
      configmaps: 0,
      secrets: 0,
      others: 0
    };

    resources.forEach(resource => {
      try {
        const doc = JSON.parse(resource);
        switch (doc.kind?.toLowerCase()) {
          case 'deployment':
            count.deployments++;
            break;
          case 'service':
            count.services++;
            break;
          case 'configmap':
            count.configmaps++;
            break;
          case 'secret':
            count.secrets++;
            break;
          default:
            count.others++;
        }
      } catch (e) {
        const kindMatch = resource.match(/kind:\s*(\w+)/i);
        if (kindMatch) {
          switch (kindMatch[1].toLowerCase()) {
            case 'deployment':
              count.deployments++;
              break;
            case 'service':
              count.services++;
              break;
            case 'configmap':
              count.configmaps++;
              break;
            case 'secret':
              count.secrets++;
              break;
            default:
              count.others++;
          }
        }
      }
    });

    return count;
  };

  const extractServices = (manifest: string): string[] => {
    const services: string[] = [];
    const resources = manifest.split('---').filter(Boolean);

    resources.forEach(resource => {
      try {
        const doc = JSON.parse(resource);
        if (doc.kind?.toLowerCase() === 'service') {
          services.push(doc.metadata?.name);
        }
      } catch (e) {
        const serviceMatch = resource.match(/kind:\s*Service[\s\S]*?name:\s*(\S+)/im);
        if (serviceMatch) {
          services.push(serviceMatch[1]);
        }
      }
    });

    return services.filter(Boolean);
  };

  const handleValidate = async () => {
    if (!isValid) return;

    setIsValidating(true);
    try {
      const payload = {
        repo: source.url,
        target_revision: source.targetRevision,
        path: source.path,
        submodules: true,
        application_specifier: source.application_specifier || {
          helm_manifest_path: ""
        }
      };

      const result = await validateApplication(payload) as ValidationResult;

      if (result.success) {
        const manifest = result.suiteable_env?.[0]?.manifest;
        if (manifest) {
          const resourceCount = calculateResourceCount(manifest);
          const services = extractServices(manifest);
          onServiceListUpdate?.(services);

          toast({
            title: "Validation Successful",
            description: (
              <div className="space-y-4">
                <div className="space-y-2">
                  <p>{result.message}</p>
                  <p className="text-sm text-muted-foreground">Type: {result.type}</p>
                </div>
                <div className="bg-muted p-4 rounded-lg space-y-2">
                  <h4 className="font-medium">Resource Summary:</h4>
                  <ul className="space-y-1 text-sm">
                    <li>Deployments: {resourceCount.deployments}</li>
                    <li>Services: {resourceCount.services}</li>
                    <li>ConfigMaps: {resourceCount.configmaps}</li>
                    <li>Secrets: {resourceCount.secrets}</li>
                    <li>Others: {resourceCount.others}</li>
                  </ul>
                </div>
              </div>
            ),
            duration: 5000,
          });
        }
      } else {
        const fullErrorMessage = `
Application Type: ${result.type}
Status: ${result.message}

Validation Details:
${result.suiteable_env?.map((env: SuiteableEnv) => `
Environment: ${env.environments}
Status: ${env.valid ? 'Valid' : 'Invalid'}
Error: ${env.error}
`).join('\n')}`;

        toast({
          variant: "destructive",
          title: "Validation Failed",
          description: (
            <div className="mt-2 space-y-4">
              <div className="flex justify-between items-center">
                <div className="font-medium text-red-700 dark:text-red-400">
                  {result.message}
                </div>
                <div className="text-sm text-red-600 bg-red-100 dark:bg-red-900/30 px-2 py-1 rounded">
                  Type: {result.type}
                </div>
              </div>

              {result.suiteable_env?.map((env: SuiteableEnv, index: number) => (
                <div key={index} className="relative bg-red-50 dark:bg-red-900/20 rounded-lg p-4">
                  <div className="flex justify-between items-center mb-2">
                    <span className="font-medium text-red-800 dark:text-red-200">
                      Environment: {env.environments}
                    </span>
                    <span className={`text-sm px-2 py-0.5 rounded ${
                      env.valid
                        ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-200'
                        : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-200'
                    }`}>
                      {env.valid ? 'Valid' : 'Invalid'}
                    </span>
                  </div>
                  <div className="font-mono text-sm">
                    <pre className="whitespace-pre-wrap break-words text-red-800 dark:text-red-200">
                      {env.error}
                    </pre>
                  </div>
                </div>
              ))}

              <div className="flex justify-end">
                <button
                  onClick={() => copyToClipboard(fullErrorMessage)}
                  className="flex items-center space-x-2 px-3 py-2 text-sm bg-red-100 hover:bg-red-200 dark:bg-red-900/30 dark:hover:bg-red-800/50 text-red-700 dark:text-red-200 rounded-md transition-colors"
                >
                  {isCopied ? (
                    <>
                      <CheckIcon className="h-4 w-4" />
                      <span>Copied!</span>
                    </>
                  ) : (
                    <>
                      <CopyIcon className="h-4 w-4" />
                      <span>Copy Full Details</span>
                    </>
                  )}
                </button>
              </div>
            </div>
          ),
          duration: 20000,
        });
      }
    } catch (error) {
      let errorData: ErrorResponse | null = null;

      if (error instanceof Error) {
        const axiosError = error as any;
        if (axiosError.response?.data) {
          errorData = axiosError.response.data as ErrorResponse;
        }
      }

      if (errorData) {
        const fullErrorMessage = `
Application Type: ${errorData.type}
Status: ${errorData.message}

Validation Details:
${errorData.suiteable_env?.map((env: SuiteableEnv) => `
Environment: ${env.environments}
Status: ${env.valid ? 'Valid' : 'Invalid'}
Error: ${env.error}
`).join('\n')}`;

        toast({
          variant: "destructive",
          title: "Validation Failed",
          description: (
            <div className="mt-2 space-y-4">
              <div className="flex justify-between items-center">
                <div className="font-medium text-red-700 dark:text-red-400">
                  {errorData.message}
                </div>
                <div className="text-sm text-red-600 bg-red-100 dark:bg-red-900/30 px-2 py-1 rounded">
                  Type: {errorData.type}
                </div>
              </div>

              {errorData.suiteable_env?.map((env: SuiteableEnv, index: number) => (
                <div key={index} className="relative bg-red-50 dark:bg-red-900/20 rounded-lg p-4">
                  <div className="flex justify-between items-center mb-2">
                    <span className="font-medium text-red-800 dark:text-red-200">
                      Environment: {env.environments}
                    </span>
                    <span className={`text-sm px-2 py-1 rounded ${
                      env.valid
                        ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-200'
                        : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-200'
                    }`}>
                      {env.valid ? 'Valid' : 'Invalid'}
                    </span>
                  </div>
                  <div className="font-mono text-sm overflow-x-auto">
                    <pre className="whitespace-pre-wrap break-words text-red-800 dark:text-red-200">
                      {env.error}
                    </pre>
                  </div>
                </div>
              ))}

              <div className="flex justify-end">
                <button
                  onClick={() => copyToClipboard(fullErrorMessage)}
                  className="flex items-center space-x-2 px-3 py-2 text-sm bg-red-100 hover:bg-red-200 dark:bg-red-900/30 dark:hover:bg-red-800/50 text-red-700 dark:text-red-200 rounded-md transition-colors"
                >
                  {isCopied ? (
                    <>
                      <CheckIcon className="h-4 w-4" />
                      <span>Copied!</span>
                    </>
                  ) : (
                    <>
                      <CopyIcon className="h-4 w-4" />
                      <span>Copy Full Details</span>
                    </>
                  )}
                </button>
              </div>
            </div>
          ),
          duration: 20000,
        });
      } else {
        toast({
          variant: "destructive",
          title: "Validation Failed",
          description: (
            <div className="mt-2 space-y-4">
              <div className="font-medium text-red-700 dark:text-red-400">
                Failed to validate application source
              </div>
              <div className="bg-red-50 dark:bg-red-900/20 rounded-lg p-4 font-mono text-sm">
                <pre className="whitespace-pre-wrap break-words text-red-800 dark:text-red-200">
                  {error instanceof Error ? error.message : String(error)}
                </pre>
              </div>
            </div>
          ),
          duration: 10000,
          action: <AlertCircle className="h-4 w-4" />,
        });
      }
    } finally {
      setIsValidating(false);
    }
  };

  return (
    <div className="flex space-x-2">
      <Button
        variant="outline"
        className="bg-blue-500 hover:bg-blue-600 text-white"
        disabled={!isValid || isValidating}
        onClick={handleValidate}
      >
        <FileText className="h-4 w-4 mr-2" />
        <span>Render & Validate</span>
      </Button>
    </div>
  );
}