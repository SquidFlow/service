'use client';

import React, { useState, useEffect, useRef } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import {
  XCircle,
  PlusCircle,
  Cpu,
  MemoryStick,
  HardDrive,
  Network,
  Box,
  X,
  CheckCircle,
  Layout,
  Settings2,
  Check,
  RefreshCw,
  HelpCircle,
} from 'lucide-react';
import { Separator } from '@/components/ui/separator';
import { Tooltip } from '@/components/ui/tooltip';
import {
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Light as SyntaxHighlighter } from 'react-syntax-highlighter';
import yaml from 'react-syntax-highlighter/dist/esm/languages/hljs/yaml';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import {
  clusterDefaults,
  // tenants,
  fieldDescriptions,
  mockYamlTemplate,
  // type TenantInfo,
  type Ingress,
  type TemplateSource,
} from './mockData';
import { DryRun } from './dryrun';
import {
  clusters,
  type ClusterInfo,
} from '@/app/dashboard/components/destinationClusterMock';
import {
  // getAvailableSecretStores,
  getSecretStoreDetails,
} from '@/app/dashboard/components/security';
// import type { SecretStore } from '@/app/dashboard/components/securityMock';
import {
  useKustomizationsData,
  useGetAvailableTenants,
  useGetClusterList,
  useGetSecretStore,
} from '@/app/api';
import { ApplicationTemplate } from '@/app/api';  // Import the type

SyntaxHighlighter.registerLanguage('yaml', yaml);

interface DeployFormProps {
  onCancel: () => void;
}

export function DeployForm({ onCancel }: DeployFormProps) {
  const { kustomizationsData } = useKustomizationsData();
  const { availableTenants } = useGetAvailableTenants();
  const { clusterList } = useGetClusterList();
  const { secretStoreList } = useGetSecretStore();
  const [ingresses, setIngresses] = useState<Ingress[]>([
    { name: '', service: '', port: '' },
  ]);
  const [templateSource, setTemplateSource] = useState<TemplateSource>({
    type: 'builtin',
    value: '',
    instanceName: '',
  });

  const [selectedClusters, setSelectedClusters] = useState<string[]>([]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    console.log('Form submitted');
    onCancel();
  };

  const addIngress = () => {
    if (ingresses.length < 2) {
      setIngresses([...ingresses, { name: '', service: '', port: '' }]);
    }
  };

  const removeIngress = (index: number) => {
    setIngresses(ingresses.filter((_, i) => i !== index));
  };

  const updateIngress = (
    index: number,
    field: keyof Ingress,
    value: string
  ) => {
    const newIngresses = [...ingresses];
    newIngresses[index][field] = value;
    setIngresses(newIngresses);
  };

  useEffect(() => {
    const calculateProgress = () => {
      let totalFields = 0;
      let completedFields = 0;

      // Template Selection
      totalFields += 1;
      if (templateSource.value) completedFields += 1;

      // Basic Information
      totalFields += 3; // tenant name, app code, namespace
      const basicInfoInputs = document.querySelectorAll(
        '#tenantName, #appCode, #namespace'
      );
      basicInfoInputs.forEach((input) => {
        if ((input as HTMLInputElement).value) completedFields += 1;
      });

      // Ingresses
      ingresses.forEach((ingress) => {
        totalFields += 3; // name, service, port
        if (ingress.name) completedFields += 1;
        if (ingress.service) completedFields += 1;
        if (ingress.port) completedFields += 1;
      });

      // Environment
      totalFields += 1;
      const envSelect = document.querySelector(
        '[placeholder="Select environment"]'
      );
      if ((envSelect as HTMLSelectElement)?.value) completedFields += 1;

      // Resource Quotas (5 fields per environment * 3 environments)
      totalFields += 15;
      const quotaInputs = document.querySelectorAll(
        '[placeholder*="vCPU"], [placeholder*="RAM"], [placeholder*="Storage"], [placeholder*="PVCs"], [placeholder*="Nodeports"]'
      );
      quotaInputs.forEach((input) => {
        if ((input as HTMLInputElement).value) completedFields += 1;
      });

      return Math.round((completedFields / totalFields) * 100);
    };

    const updateProgress = () => {
      setProgress(calculateProgress());
    };

    updateProgress();

    const form = document.querySelector('form');
    const observer = new MutationObserver(updateProgress);

    if (form) {
      observer.observe(form, {
        subtree: true,
        childList: true,
        characterData: true,
        attributes: true,
      });
    }

    return () => observer.disconnect();
  }, [ingresses, templateSource.value]);

  const [progress, setProgress] = useState(0);

  const renderEnvironmentSection = () => (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
          Destination Cluster
        </h3>
        <span className="text-sm text-gray-500">
          {selectedClusters.length} cluster(s) selected
        </span>
      </div>
      <div className="flex flex-wrap gap-2">
        {clusterList.map((cluster) => (
          <Button
            key={cluster.name}
            className="flex items-center gap-2 group relative"
            variant={
              selectedClusters.includes(cluster.name) ? 'default' : 'outline'
            }
            onClick={() => {
              setSelectedClusters((prev) =>
                prev.includes(cluster.name)
                  ? prev.filter((c) => c !== cluster.name)
                  : [...prev, cluster.name]
              );
            }}
          >
            <div className="flex items-center gap-2">
              <span className="font-mono font-medium">{cluster.name}</span>
              <span
                className={`text-xs px-2 py-0.5 rounded-full ${
                  cluster.environment === 'SIT'
                    ? 'bg-blue-100 text-blue-700'
                    : cluster.environment === 'UAT'
                      ? 'bg-green-100 text-green-700'
                      : 'bg-purple-100 text-purple-700'
                }`}
              >
                {cluster.environment}
              </span>
            </div>
            {selectedClusters.includes(cluster.name) && (
              <CheckCircle className="h-4 w-4 ml-1" />
            )}
            {cluster.builtin && (
              <span className="absolute -top-2 -right-2 px-1.5 py-0.5 text-[10px] font-medium bg-purple-100 text-purple-800 rounded-full">
                Builtin
              </span>
            )}
          </Button>
        ))}
      </div>
    </div>
  );

  const resourceDescriptions = {
    cpu: {
      label: 'CPU',
      tooltip:
        'Maximum CPU cores allocated. 1 CPU = 1000m (millicores). For example, 2000m = 2 CPU cores',
    },
    memory: {
      label: 'Memory',
      tooltip:
        'Maximum RAM allocated. Memory in GiB. For example, 4GiB = 4096MiB',
    },
    storage: {
      label: 'Storage',
      tooltip:
        'Maximum storage space for persistent volumes. Storage in GiB. Local and network storage combined',
    },
    pvcs: {
      label: 'PVCs',
      tooltip: 'Maximum number of persistent volumes that can be created',
    },
    nodeports: {
      label: 'NodePorts',
      tooltip: 'Maximum number of NodePort services allowed',
    },
  };

  const renderResourceQuotas = () => (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
        Tenant Resource Quotas
      </h3>
      <div className="grid grid-cols-4 gap-6">
        {selectedClusters.map((clusterName) => {
          const clusterInfo = clusters.find(
            (c: ClusterInfo) => c.name === clusterName
          );
          if (!clusterInfo) return null;

          return (
            <Card
              key={clusterName}
              className="shadow-sm bg-gradient-to-br from-gray-50 to-white dark:from-gray-900 dark:to-gray-800 border border-gray-200/50 dark:border-gray-700/50"
            >
              <CardHeader className="pb-2">
                <CardTitle className="text-base font-medium">
                  {clusterName}
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-6">
                <TooltipProvider>
                  {Object.entries(resourceDescriptions).map(([key, desc]) => (
                    <div key={key} className="space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-2">
                          {key === 'cpu' && (
                            <Cpu className="h-4 w-4 text-gray-500" />
                          )}
                          {key === 'memory' && (
                            <MemoryStick className="h-4 w-4 text-gray-500" />
                          )}
                          {key === 'storage' && (
                            <HardDrive className="h-4 w-4 text-gray-500" />
                          )}
                          {key === 'pvcs' && (
                            <Network className="h-4 w-4 text-gray-500" />
                          )}
                          {key === 'nodeports' && (
                            <Box className="h-4 w-4 text-gray-500" />
                          )}
                          <Label className="font-medium">{desc.label}</Label>
                        </div>
                        <Tooltip>
                          <TooltipTrigger>
                            <HelpCircle className="h-4 w-4 text-gray-400 hover:text-gray-500" />
                          </TooltipTrigger>
                          <TooltipContent>
                            <p className="max-w-xs">{desc.tooltip}</p>
                          </TooltipContent>
                        </Tooltip>
                      </div>
                      <div className="w-full px-3 py-2 rounded-md bg-gray-100 dark:bg-gray-800 font-mono text-sm">
                        {
                          clusterInfo.resourceQuota[
                            key as keyof typeof clusterInfo.resourceQuota
                          ]
                        }
                      </div>
                    </div>
                  ))}
                </TooltipProvider>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );

  // dryrun 相关状态
  const [isDryRunOpen, setIsDryRunOpen] = useState(false);
  const [dryRunYaml, setDryRunYaml] = useState<
    { cluster: string; content: string }[]
  >([]);
  const formRef = useRef<HTMLDivElement>(null);

  const handleDryRun = async () => {
    const namespaceElement = document.getElementById(
      'namespace'
    ) as HTMLInputElement;
    const namespace = namespaceElement?.value;

    const yamls = selectedClusters.map((cluster) => ({
      cluster,
      content: generateYAML(namespace, cluster),
    }));

    setDryRunYaml(yamls);
    setIsDryRunOpen(true);
  };

  const closeDryRun = () => {
    setIsDryRunOpen(false);
    if (formRef.current) {
      formRef.current.style.transform = 'translateX(0)';
    }
  };

  const [enableExternalSecret, setEnableExternalSecret] = useState(false);
  const [selectedTenant, setSelectedTenant] = useState<string>('');

  interface ValidationResult {
    environment: string;
    isValid: boolean;
    message?: string;
  }

  const [validationResults, setValidationResults] = useState<
    ValidationResult[]
  >([]);
  const [isValidating, setIsValidating] = useState(false);

  const validateExternalTemplate = async () => {
    if (!templateSource.value || !templateSource.targetRevision) {
      return;
    }

    setIsValidating(true);
    setValidationResults([]);

    try {
      // 这里模拟API调用，实际实现时替换为真实的API调用
      await new Promise((resolve) => setTimeout(resolve, 1500));

      // 模拟验证结果
      const results: ValidationResult[] = Object.keys(clusterDefaults).map(
        (env) => ({
          environment: env,
          isValid: Math.random() > 0.3,
          message:
            Math.random() > 0.3
              ? 'Template structure validated successfully'
              : 'Invalid template structure or missing required files',
        })
      );

      setValidationResults(results);
    } catch (error) {
      console.error('Validation failed:', error);
    } finally {
      setIsValidating(false);
    }
  };

  // 在 DeployForm 组件中添加用户租户状态
  // const [availableTenants, setAvailableTenants] = useState<TenantInfo[]>([]);
  const [selectedTenantId, setSelectedTenantId] = useState<string>('');

  // 添加获取用户租户的 effect
  // useEffect(() => {
  //   // 这里模拟从 API 获取当前用户可用的租户列表
  //   const fetchUserTenants = async () => {
  //     // 模拟 API 调用
  //     const userTenants = tenants.filter(tenant =>
  //       // 这里可以添加实际的权限检查逻辑
  //       true
  //     );
  //     setAvailableTenants(userTenants);
  //   };

  //   fetchUserTenants();
  // }, []);

  // interface SyncOptions {
  //   respectIgnoreDifferences: boolean;
  //   createNamespace: boolean;
  //   applyOutOfSyncOnly: boolean;
  //   pruneLast: boolean;
  //   serverSideApply: boolean;
  // }

  // const [syncOptions, setSyncOptions] = useState<SyncOptions>({
  //   respectIgnoreDifferences: true,
  //   createNamespace: false,
  //   applyOutOfSyncOnly: true,
  //   pruneLast: true,
  //   serverSideApply: true,
  // });

  const [selectedSecretStore, setSelectedSecretStore] = useState<string>('');
  // const [availableSecretStores, setAvailableSecretStores] = useState<
  //   SecretStore[]
  // >([]);

  // useEffect(() => {
  //   const stores = getAvailableSecretStores();
  //   setAvailableSecretStores(stores);
  // }, []);

  const generateYAML = (namespace: string, cluster: string) => {
    let yaml = mockYamlTemplate(namespace, cluster, clusterDefaults);

    if (enableExternalSecret && selectedSecretStore) {
      const secretStore = getSecretStoreDetails(selectedSecretStore);
      if (secretStore) {
        yaml += `\n---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: ${templateSource.instanceName}-external-secret
  namespace: ${namespace}
spec:
  refreshInterval: "1h"
  secretStoreRef:
    name: ${secretStore.name}
    kind: ${secretStore.type}
  target:
    name: ${templateSource.instanceName}-secret
  data:
  - secretKey: example-key
    remoteRef:
      key: ${secretStore.path}/${templateSource.instanceName}
      property: value`;
      }
    }
    return yaml;
  };

  // 更新 ExternalSecret 配置部分
  const renderExternalSecretConfig = () => (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="space-y-0.5">
          <Label>External Secrets Integration</Label>
        </div>
        <Switch
          checked={enableExternalSecret}
          onCheckedChange={setEnableExternalSecret}
        />
      </div>

      {enableExternalSecret && (
        <div className="space-y-4">
          <div className="space-y-2">
            <Label>Secret Store</Label>
            <Select
              value={selectedSecretStore}
              onValueChange={setSelectedSecretStore}
            >
              <SelectTrigger>
                {selectedSecretStore ? (
                  <div className="flex items-center justify-between w-full">
                    <span className="font-medium">{selectedSecretStore}</span>
                    <span className="text-xs text-gray-500 font-mono">
                      {getSecretStoreDetails(selectedSecretStore)?.path}
                    </span>
                  </div>
                ) : (
                  <SelectValue placeholder="Select a Secret Store" />
                )}
              </SelectTrigger>
              <SelectContent>
                {secretStoreList.map((store) => (
                  <SelectItem key={store.name} value={store.name}>
                    <div className="space-y-1">
                      <div className="flex items-center space-x-2">
                        <span className="font-medium">{store.name}</span>
                        <Badge
                          variant="outline"
                          className={getProviderBadgeColor(store.provider)}
                        >
                          {store.provider}
                        </Badge>
                        <Badge variant="outline">{store.type}</Badge>
                      </div>
                      <div className="text-xs text-gray-500 font-mono">
                        Path: {store.path}
                      </div>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      )}
    </div>
  );

  // 添加 provider badge 颜色函数
  const getProviderBadgeColor = (provider: string) => {
    const colors = {
      AWS: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400',
      GCP: 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400',
      Azure:
        'bg-purple-100 text-purple-800 dark:bg-purple-900/20 dark:text-purple-400',
      Vault:
        'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/20 dark:text-indigo-400',
      CyberArk:
        'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400',
    };
    return (
      colors[provider as keyof typeof colors] || 'bg-gray-100 text-gray-800'
    );
  };

  return (
    <div className="flex relative w-full min-h-screen p-6">
      <div
        ref={formRef}
        className="flex space-x-6 transition-transform duration-300 ease-in-out mx-auto"
        style={{ width: isDryRunOpen ? '80%' : '100%' }}
      >
        <div className="flex flex-col space-y-6 w-full max-w-[2400px] mx-auto">
          <Card className="bg-white dark:bg-gray-800 shadow-lg rounded-lg overflow-hidden border border-gray-100 dark:border-gray-700">
            <CardHeader className="border-b border-gray-100 dark:border-gray-700">
              <CardTitle className="text-xl font-semibold flex items-center space-x-3">
                <Layout className="h-6 w-6 text-blue-500 dark:text-blue-400" />
                <span className="bg-gradient-to-r from-gray-800 to-gray-600 dark:from-gray-200 dark:to-gray-400 text-transparent bg-clip-text">
                  Application Template Selection
                </span>
                {templateSource.value && (
                  <CheckCircle className="h-5 w-5 text-emerald-500" />
                )}
              </CardTitle>
              <p className="text-sm text-gray-500 dark:text-gray-400 ml-9">
                Choose a template to define your application deployment
                configuration
              </p>
            </CardHeader>
            <CardContent className="p-6">
              <Tabs
                defaultValue="builtin"
                onValueChange={(value) =>
                  setTemplateSource({
                    type: value as 'builtin' | 'external',
                    value: '',
                  })
                }
              >
                <TabsList className="grid w-full grid-cols-2 mb-6 p-1 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  <TabsTrigger
                    value="builtin"
                    className="flex items-center space-x-2 data-[state=active]:bg-white dark:data-[state=active]:bg-gray-700 data-[state=active]:shadow-sm transition-all duration-200"
                  >
                    <div className="flex items-center space-x-2">
                      <span>Built-in Templates</span>
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger>
                            <HelpCircle className="h-4 w-4 text-gray-400" />
                          </TooltipTrigger>
                          <TooltipContent>
                            <p className="max-w-xs">
                              {fieldDescriptions.builtinTemplate.tooltip}
                            </p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                  </TabsTrigger>
                  <TabsTrigger
                    value="external"
                    className="flex items-center space-x-2"
                  >
                    <span>External Template</span>
                    {templateSource.type === 'external' &&
                      templateSource.value && (
                        <CheckCircle className="h-4 w-4 text-green-500" />
                      )}
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="builtin" className="space-y-4">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-base font-medium">
                        Select Template
                      </Label>
                      {templateSource.value && (
                        <span className="text-sm text-green-500 flex items-center space-x-1">
                          <CheckCircle className="h-4 w-4" />
                          <span>Template selected</span>
                        </span>
                      )}
                    </div>
                    <Select
                      value={templateSource.value}
                      onValueChange={(value) =>
                        setTemplateSource({ ...templateSource, value })
                      }
                    >
                      <SelectTrigger id="builtinTemplate" className="w-full">
                        <SelectValue placeholder="Choose a built-in template" />
                      </SelectTrigger>
                      <SelectContent>
                        {kustomizationsData.map((kustomization: ApplicationTemplate) => (
                          <SelectItem
                            key={kustomization.id}
                            value={kustomization.name}
                          >
                            <div className="space-y-1">
                              <div className="font-medium">
                                {kustomization.name}
                              </div>
                              <div className="text-sm text-gray-500">
                                Path: {kustomization.path}
                                <div className="flex flex-wrap gap-1 mt-1">
                                  {kustomization.environments.map((env) => (
                                    <span
                                      key={env}
                                      className="px-1.5 py-0.5 text-xs rounded-full bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300"
                                    >
                                      {env}
                                    </span>
                                  ))}
                                </div>
                              </div>
                            </div>
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  {templateSource.value && (
                    <div className="mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                      <h4 className="font-medium mb-2">Template Details</h4>
                      <p className="text-sm text-gray-600 dark:text-gray-300">
                        This template provides a standardized configuration for
                        deploying your application. It includes recommended
                        settings for resources, scaling, and monitoring.
                      </p>
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="external" className="space-y-4">
                  <div className="space-y-4">
                    {/* Repository URL Input */}
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <Label className="text-base font-medium">
                          External Template URL
                        </Label>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger>
                              <HelpCircle className="h-4 w-4 text-gray-400" />
                            </TooltipTrigger>
                            <TooltipContent>
                              <div className="space-y-2 max-w-xs">
                                <p>Supported URL formats:</p>
                                <ul className="list-disc pl-4 text-sm">
                                  <li>
                                    HTTP(S): http(s)://github.com/user/repo.git
                                  </li>
                                  <li>SSH: git@github.com:user/repo.git</li>
                                </ul>
                              </div>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      </div>
                      <Input
                        id="externalTemplate"
                        placeholder="Enter template URL (e.g., https://github.com/user/repo.git)"
                        value={templateSource.value}
                        onChange={(e) =>
                          setTemplateSource({
                            ...templateSource,
                            value: e.target.value,
                          })
                        }
                        className="w-full"
                      />
                    </div>

                    {/* Target Revision Input */}
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <Label className="text-base font-medium">
                          Target Revision
                        </Label>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger>
                              <HelpCircle className="h-4 w-4 text-gray-400" />
                            </TooltipTrigger>
                            <TooltipContent>
                              <div className="space-y-2 max-w-xs">
                                <p>Supported revision formats:</p>
                                <ul className="list-disc pl-4 text-sm">
                                  <li>Branch name (e.g., main, develop)</li>
                                  <li>Commit hash (e.g., 1a2b3c4)</li>
                                  <li>Tag (e.g., v1.0.0)</li>
                                </ul>
                              </div>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      </div>
                      <Input
                        id="targetRevision"
                        placeholder="Enter target revision (branch, commit hash, or tag)"
                        value={templateSource.targetRevision}
                        onChange={(e) =>
                          setTemplateSource({
                            ...templateSource,
                            targetRevision: e.target.value,
                          })
                        }
                        className="w-full"
                      />
                      {templateSource.targetRevision && (
                        <p className="text-sm text-gray-500">
                          Your template will be synced from:{' '}
                          {templateSource.targetRevision}
                        </p>
                      )}
                    </div>

                    <div className="flex justify-between items-center">
                      <Button
                        onClick={validateExternalTemplate}
                        disabled={
                          !templateSource.value ||
                          !templateSource.targetRevision ||
                          isValidating
                        }
                        className={`${
                          templateSource.value && templateSource.targetRevision
                            ? 'bg-green-500 hover:bg-green-600'
                            : 'bg-gray-300'
                        } text-white transition-colors duration-200`}
                      >
                        {isValidating ? (
                          <>
                            <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                            Validating...
                          </>
                        ) : (
                          <>
                            <Check className="h-4 w-4 mr-2" />
                            Validate Template
                          </>
                        )}
                      </Button>
                      {validationResults.length > 0 && (
                        <span className="text-sm text-gray-500">
                          Validation completed for {validationResults.length}{' '}
                          environments
                        </span>
                      )}
                    </div>

                    {/* 验证结果显示 */}
                    {validationResults.length > 0 && (
                      <div className="mt-4 space-y-3 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                        <h4 className="font-medium text-sm text-gray-700 dark:text-gray-300">
                          Validation Results
                        </h4>
                        <div className="grid grid-cols-2 gap-3">
                          {validationResults.map((result) => (
                            <div
                              key={result.environment}
                              className={`p-3 rounded-lg border ${
                                result.isValid
                                  ? 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-900/20'
                                  : 'border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-900/20'
                              }`}
                            >
                              <div className="flex items-center justify-between">
                                <span className="font-medium">
                                  {result.environment}
                                </span>
                                {result.isValid ? (
                                  <CheckCircle className="h-4 w-4 text-green-500" />
                                ) : (
                                  <XCircle className="h-4 w-4 text-red-500" />
                                )}
                              </div>
                              <p
                                className={`text-sm mt-1 ${
                                  result.isValid
                                    ? 'text-green-600'
                                    : 'text-red-600'
                                }`}
                              >
                                {result.message}
                              </p>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </TabsContent>
              </Tabs>
            </CardContent>
          </Card>

          <Separator className="my-6" />

          {/* Template Configuration Block - Update to ArgoCD Application Block */}
          <Card className="bg-white dark:bg-gray-800 shadow-lg rounded-lg overflow-hidden border border-gray-100 dark:border-gray-700">
            <CardHeader className="border-b border-gray-100 dark:border-gray-700">
              <CardTitle className="text-xl font-semibold flex items-center space-x-3">
                <Settings2 className="h-6 w-6 text-purple-500 dark:text-purple-400" />
                <span className="bg-gradient-to-r from-gray-800 to-gray-600 dark:from-gray-200 dark:to-gray-400 text-transparent bg-clip-text">
                  ArgoCD Application Instantiation
                </span>
              </CardTitle>
              <p className="text-sm text-gray-500 dark:text-gray-400 ml-9">
                Configure the ArgoCD Application parameters to instantiate your
                selected template
              </p>
            </CardHeader>
            <CardContent className="p-6">
              <div className="space-y-8">
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label
                      htmlFor="instanceName"
                      className="text-base font-medium"
                    >
                      ArgoCD Application Name
                    </Label>
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger>
                          <HelpCircle className="h-4 w-4 text-gray-400" />
                        </TooltipTrigger>
                        <TooltipContent>
                          <p className="max-w-xs">
                            Provide a unique name for this template instance.
                            This name will be used to identify your deployment.
                          </p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                  <Input
                    id="instanceName"
                    placeholder="Enter ArgoCD application name"
                    value={templateSource.instanceName}
                    onChange={(e) =>
                      setTemplateSource({
                        ...templateSource,
                        instanceName: e.target.value,
                      })
                    }
                    className="w-full"
                  />
                  {templateSource.instanceName && (
                    <p className="text-sm text-gray-500">
                      Your template will be instantiated as:{' '}
                      {templateSource.instanceName}
                    </p>
                  )}
                </div>

                {/* Basic Information */}
                <div className="space-y-6">
                  <div className="grid grid-cols-2 gap-6">
                    {/* Tenant Name - 更新为 Select 组件 */}
                    <div className="space-y-2">
                      <div className="flex items-center space-x-2">
                        <Label
                          htmlFor="tenantName"
                          className="text-sm font-medium"
                        >
                          {fieldDescriptions.tenantName.label}
                        </Label>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger>
                              <HelpCircle className="h-4 w-4 text-gray-400" />
                            </TooltipTrigger>
                            <TooltipContent>
                              <p className="max-w-xs">
                                {fieldDescriptions.tenantName.tooltip}
                              </p>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      </div>
                      <Select
                        value={selectedTenantId}
                        onValueChange={setSelectedTenantId}
                      >
                        <SelectTrigger id="tenantName" className="w-full">
                          <SelectValue placeholder="Select tenant"></SelectValue>
                        </SelectTrigger>
                        <SelectContent>
                          {availableTenants.map((tenant) => (
                            <SelectItem key={tenant.id} value={tenant.id}>
                              {tenant.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      {/* {selectedTenantId && (
                        <p className="text-sm text-gray-500">
                          Selected tenant:{' '}
                          {
                            availableTenants.find(
                              (t) => t.id === selectedTenantId
                            )?.name
                          }
                        </p>
                      )} */}
                    </div>

                    {/* App Code */}
                    <div className="space-y-2">
                      <div className="flex items-center space-x-2">
                        <Label
                          htmlFor="appCode"
                          className="text-sm font-medium"
                        >
                          {fieldDescriptions.appCode.label}
                        </Label>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger>
                              <HelpCircle className="h-4 w-4 text-gray-400" />
                            </TooltipTrigger>
                            <TooltipContent>
                              <p className="max-w-xs">
                                {fieldDescriptions.appCode.tooltip}
                              </p>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      </div>
                      <Select>
                        <SelectTrigger id="appCode">
                          <SelectValue placeholder="Select app code" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="code1">Code 1</SelectItem>
                          <SelectItem value="code2">Code 2</SelectItem>
                          <SelectItem value="code3">Code 3</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Label
                          htmlFor="namespace"
                          className="text-sm font-medium"
                        >
                          {fieldDescriptions.namespace.label}
                        </Label>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger>
                              <HelpCircle className="h-4 w-4 text-gray-400" />
                            </TooltipTrigger>
                            <TooltipContent>
                              <p className="max-w-xs">
                                {fieldDescriptions.namespace.tooltip}
                              </p>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      </div>
                      <Input
                        id="namespace"
                        placeholder="Enter namespace"
                        className="w-full"
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Label
                          htmlFor="description"
                          className="text-sm font-medium"
                        >
                          {fieldDescriptions.description.label}
                        </Label>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger>
                              <HelpCircle className="h-4 w-4 text-gray-400" />
                            </TooltipTrigger>
                            <TooltipContent>
                              <p className="max-w-xs">
                                {fieldDescriptions.description.tooltip}
                              </p>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      </div>
                      <Textarea
                        id="description"
                        placeholder="Enter description"
                        className="w-full"
                      />
                    </div>
                  </div>
                </div>

                <Separator />

                {/* Ingresses */}
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
                    {fieldDescriptions.ingress.label}
                  </h3>
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <HelpCircle className="h-4 w-4 text-gray-400" />
                      </TooltipTrigger>
                      <TooltipContent>
                        <div className="space-y-2">
                          <p className="max-w-xs">
                            {fieldDescriptions.ingress.tooltip}
                          </p>
                          <ul className="text-sm list-disc pl-4">
                            <li>Name: {fieldDescriptions.ingress.name}</li>
                            <li>
                              Service: {fieldDescriptions.ingress.service}
                            </li>
                            <li>Port: {fieldDescriptions.ingress.port}</li>
                          </ul>
                        </div>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                  {ingresses.map((ingress, index) => (
                    <div key={index} className="flex items-center space-x-4">
                      <Input
                        placeholder="Name"
                        value={ingress.name}
                        onChange={(e) =>
                          updateIngress(index, 'name', e.target.value)
                        }
                        className="flex-1"
                      />
                      <Input
                        placeholder="Service"
                        value={ingress.service}
                        onChange={(e) =>
                          updateIngress(index, 'service', e.target.value)
                        }
                        className="flex-1"
                      />
                      <Input
                        placeholder="Port"
                        value={ingress.port}
                        onChange={(e) =>
                          updateIngress(index, 'port', e.target.value)
                        }
                        className="flex-1"
                      />
                      <Button
                        variant="outline"
                        size="icon"
                        onClick={() => removeIngress(index)}
                        className="flex-shrink-0"
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                  <Button
                    variant="outline"
                    onClick={addIngress}
                    className="mt-2 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 hover:from-blue-100 hover:to-indigo-100 dark:hover:from-blue-900/30 dark:hover:to-indigo-900/30 transition-all duration-200"
                  >
                    <PlusCircle className="h-4 w-4 mr-2" />
                    Add Ingress
                  </Button>
                </div>

                <Separator />

                {/* Environment */}
                {renderEnvironmentSection()}

                <Separator />

                {/* Resource Quotas */}
                {selectedClusters.length > 0 && renderResourceQuotas()}

                <Separator />

                {/* ExternalSecret Integration Section */}
                {renderExternalSecretConfig()}

                <Separator className="my-6" />
              </div>
            </CardContent>
          </Card>

          <Separator className="my-6" />

          {/* Action Buttons Block */}
          <Card className="bg-gradient-to-br from-white to-gray-50 dark:from-gray-800 dark:to-gray-900 shadow-lg rounded-lg overflow-hidden border border-gray-200/50 dark:border-gray-700/50">
            <CardContent className="p-6">
              <div className="flex justify-between space-x-6">
                <Button
                  variant="outline"
                  onClick={onCancel}
                  className="w-full bg-gradient-to-r from-gray-50 to-gray-100 dark:from-gray-800 dark:to-gray-700 hover:from-gray-100 hover:to-gray-200 dark:hover:from-gray-700 dark:hover:to-gray-600 transition-all duration-200"
                >
                  Cancel
                </Button>
                <Button
                  variant="outline"
                  onClick={handleDryRun}
                  className="w-full bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 hover:from-blue-100 hover:to-indigo-100 dark:hover:from-blue-900/30 dark:hover:to-indigo-900/30 transition-all duration-200"
                >
                  Dry Run
                </Button>
                <Button
                  onClick={handleSubmit}
                  className="w-full bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 transition-all duration-200"
                >
                  Submit
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
      <DryRun isOpen={isDryRunOpen} yamls={dryRunYaml} onClose={closeDryRun} />
      {/* Progress Bar - adjust its position when dry run is open */}
      <div
        className={`fixed transition-all duration-300 ease-in-out ${
          isDryRunOpen ? 'right-[47%]' : 'right-6'
        } top-1/2 -translate-y-1/2 h-64 flex flex-col items-center space-y-2`}
      >
        <div className="relative h-full w-4 bg-gradient-to-t from-gray-200 to-gray-100 dark:from-gray-700 dark:to-gray-600 rounded-full overflow-hidden">
          <div
            className={`absolute bottom-0 w-full rounded-full transition-all duration-500 ${
              progress === 100
                ? 'bg-gradient-to-t from-green-500 to-green-400'
                : 'bg-gradient-to-t from-blue-600 to-blue-500'
            }`}
            style={{
              height: `${progress}%`,
              transition: 'height 0.5s ease-in-out',
            }}
          />
        </div>
        <span
          className={`text-sm font-medium ${
            progress === 100 ? 'text-green-500' : 'text-blue-500'
          }`}
        >
          {progress}%
        </span>
      </div>
    </div>
  );
}
