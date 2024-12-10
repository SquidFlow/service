import { useState } from "react";
import { Button } from "@/components/ui/button";
import { CheckCircle } from "lucide-react";
import { useToast } from "@/components/ui/use-toast";
import { AlertCircle, CopyIcon, CheckIcon } from "lucide-react";
import { useApplicationStore } from '@/store';

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
}

interface ErrorResponse {
  success: boolean;
  message: string;
  type: string;
  suiteable_env?: Array<{
    environments: string;
    valid: boolean;
    error: string;
  }>;
}

export function ValidateSourceButton({ isValid, source }: ValidateSourceButtonProps) {
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

      const result = await validateApplication(payload);

      if (result.success) {
        toast({
          title: "Validation Successful",
          description: (
            <div className="space-y-2">
              <p>{result.message}</p>
              <p className="text-sm text-muted-foreground">Type: {result.type}</p>
            </div>
          ),
          duration: 3000,
        });
      } else {
        const fullErrorMessage = `
Application Type: ${result.type}
Status: ${result.message}

Validation Details:
${result.suiteable_env?.map(env => `
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

              {result.suiteable_env?.map((env, index) => (
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
      // 处理 HTTP 错误和其他错误
      let errorData: ErrorResponse | null = null;

      if (error instanceof Error) {
        const axiosError = error as any;
        if (axiosError.response?.data) {
          errorData = axiosError.response.data as ErrorResponse;
        }
      }

      if (errorData) {
        // 如果有错误响应数据，使用相同的错误展示格式
        const fullErrorMessage = `
Application Type: ${errorData.type}
Status: ${errorData.message}

Validation Details:
${errorData.suiteable_env?.map(env => `
Environment: ${env.environments}
Status: ${env.valid ? 'Valid' : 'Invalid'}
Error: ${env.error}
`).join('\n')}`;

        toast({
          variant: "destructive",
          title: "Validation Failed",
          description: (
            <div className="mt-2 space-y-4">
              {/* 错误概述 */}
              <div className="flex justify-between items-center">
                <div className="font-medium text-red-700 dark:text-red-400">
                  {errorData.message}
                </div>
                <div className="text-sm text-red-600 bg-red-100 dark:bg-red-900/30 px-2 py-1 rounded">
                  Type: {errorData.type}
                </div>
              </div>

              {/* 详细错误信息 */}
              {errorData.suiteable_env?.map((env, index) => (
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

              {/* 复制按钮 */}
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
        // 如果没有错误响应数据，显示基本错误信息
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
    <Button
      variant={isValid ? "default" : "secondary"}
      className={`flex items-center space-x-2 transition-colors ${
        isValid
          ? "bg-green-500 hover:bg-green-600 text-white"
          : "bg-gray-200 text-gray-500 cursor-not-allowed"
      }`}
      disabled={!isValid || isValidating}
      onClick={handleValidate}
    >
      <CheckCircle className="h-4 w-4" />
      <span>{isValidating ? "Validating..." : "Validate App Source"}</span>
    </Button>
  );
}