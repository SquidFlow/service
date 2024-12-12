import { useEffect, useState } from 'react';
import { Card, CardHeader, CardContent, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Layout, HelpCircle, Plus, X, AlertCircle } from "lucide-react";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Button } from "@/components/ui/button";
import { useDeployForm } from './DeployFormContext';
import { Switch } from "@/components/ui/switch";
import { useTenantStore, useSecretStore } from '@/store';

interface IngressRule {
  name: string;
  service: string;
  port: string;
}

interface ValidationState {
  name: {
    isValid: boolean;
    message?: string;
  };
  namespace: {
    isValid: boolean;
    message?: string;
  };
}

export function ApplicationSection() {
  const {
    source,
    setSource,
    availableServices
  } = useDeployForm();

  const {
    data: tenants,
    appCodes,
    fetch: fetchTenants,
    fetchAppCodes,
    isLoading,
    error
  } = useTenantStore();

  const { data: secretStoreList } = useSecretStore();

  const [validation, setValidation] = useState<ValidationState>({
    name: { isValid: true },
    namespace: { isValid: true }
  });

  useEffect(() => {
    fetchTenants();
    fetchAppCodes();
  }, [fetchTenants, fetchAppCodes]);

  useEffect(() => {
    console.log('Current app codes:', appCodes);
  }, [appCodes]);

  useEffect(() => {
    if (!source.tenant && tenants.length > 0) {
      setSource(prev => ({ ...prev, tenant: tenants[0].name }));
    }
    if (!source.appCode && appCodes.length > 0) {
      setSource(prev => ({ ...prev, appCode: appCodes[0] }));
    }
  }, [tenants, appCodes]);

  const addIngressRule = () => {
    setSource(prev => ({
      ...prev,
      ingress: [
        ...(prev.ingress || []),
        { name: '', service: '', port: '' }
      ]
    }));
  };

  const removeIngressRule = (index: number) => {
    setSource(prev => ({
      ...prev,
      ingress: prev.ingress?.filter((_, i) => i !== index)
    }));
  };

  const updateIngressRule = (index: number, field: keyof IngressRule, value: string) => {
    setSource(prev => ({
      ...prev,
      ingress: prev.ingress?.map((rule, i) =>
        i === index ? { ...rule, [field]: value } : rule
      )
    }));
  };

  const handleServiceChange = (index: number, value: string) => {
    setSource(prev => ({
      ...prev,
      ingress: prev.ingress?.map((rule, i) =>
        i === index ? { ...rule, service: value } : rule
      )
    }));
  };

  // 验证应用名称
  const validateName = (name: string) => {
    if (!name) {
      return { isValid: false, message: "Application name is required" };
    }
    if (name.toLowerCase() === 'new') {
      return { isValid: false, message: "Application name cannot be 'new'" };
    }
    if (name.startsWith('_') || name.startsWith('-')) {
      return { isValid: false, message: "Application name cannot start with '_' or '-'" };
    }
    // 可以添加更多验证规则，如：只允许字母数字和中横线
    if (!/^[a-zA-Z0-9][-a-zA-Z0-9]*$/.test(name)) {
      return { isValid: false, message: "Application name must start with a letter or number and can only contain letters, numbers, and hyphens" };
    }
    return { isValid: true };
  };

  // 验证命名空间
  const validateNamespace = (namespace: string) => {
    if (!namespace) {
      return { isValid: false, message: "Namespace is required" };
    }
    if (namespace.toLowerCase().startsWith('kube')) {
      return { isValid: false, message: "Namespace cannot start with 'kube'" };
    }
    // 可以添加更多验证规则，如：只允许字母数字和中横线
    if (!/^[a-z0-9][a-z0-9-]*$/.test(namespace)) {
      return { isValid: false, message: "Namespace must start with a letter or number and can only contain lowercase letters, numbers, and hyphens" };
    }
    return { isValid: true };
  };

  const handleNameChange = (value: string) => {
    const nameValidation = validateName(value);
    setValidation(prev => ({
      ...prev,
      name: nameValidation
    }));
    setSource(prev => ({ ...prev, name: value }));
  };

  const handleNamespaceChange = (value: string) => {
    const namespaceValidation = validateNamespace(value);
    setValidation(prev => ({
      ...prev,
      namespace: namespaceValidation
    }));
    setSource(prev => ({ ...prev, namespace: value }));
  };

  return (
    <Card>
      <CardHeader className="pb-6">
        <CardTitle className="flex items-center space-x-3 text-xl">
          <Layout className="h-6 w-6 text-blue-500" />
          <span>ArgoCD Application Instantiation</span>
        </CardTitle>
        <p className="text-sm text-muted-foreground mt-2">
          Configure the ArgoCD Application parameters to instantiate your selected template
        </p>
      </CardHeader>
      <CardContent className="space-y-8">
        <div className="space-y-3">
          <div className="flex items-center space-x-2">
            <Label className="text-sm font-medium">
              ArgoCD Application Name
              <span className="text-red-500 ml-1">*</span>
            </Label>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <HelpCircle className="h-4 w-4 text-gray-400" />
                </TooltipTrigger>
                <TooltipContent>
                  <p className="max-w-xs">The name of your ArgoCD application</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          <div className="relative">
            <Input
              required
              value={source.name}
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder="Enter ArgoCD application name"
              className={`${
                !validation.name.isValid && source.name
                  ? 'border-red-500 focus:ring-red-500'
                  : validation.name.isValid && source.name
                  ? 'border-green-500 focus:ring-green-500'
                  : ''
              }`}
            />
            {!validation.name.isValid && source.name && (
              <div className="flex items-center mt-1 text-sm text-red-500">
                <AlertCircle className="h-4 w-4 mr-1" />
                <span>{validation.name.message}</span>
              </div>
            )}
            {validation.name.isValid && source.name && (
              <div className="absolute right-3 top-2.5 h-5 w-5 text-green-500">
                <svg viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              </div>
            )}
          </div>
        </div>

        <div className="grid grid-cols-2 gap-8">
          <div>
            <div className="flex items-center space-x-2">
              <Label>
                Tenant Name
                <span className="text-red-500 ml-1">*</span>
              </Label>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <HelpCircle className="h-4 w-4 text-gray-400" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p className="max-w-xs">Select your tenant</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <Select
              required
              value={source.tenant || ''}
              onValueChange={(value) => setSource(prev => ({ ...prev, tenant: value }))}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select tenant" />
              </SelectTrigger>
              <SelectContent>
                {tenants.map((tenant) => (
                  <SelectItem key={tenant.name} value={tenant.name}>
                    {tenant.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div>
            <div className="flex items-center space-x-2">
              <Label>
                App Code
                <span className="text-red-500 ml-1">*</span>
              </Label>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <HelpCircle className="h-4 w-4 text-gray-400" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p className="max-w-xs">Select your application code</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <Select
              required
              value={source.appCode || ''}
              onValueChange={(value) => {
                console.log('Selected app code:', value);
                setSource(prev => ({ ...prev, appCode: value }));
              }}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Select app code" />
              </SelectTrigger>
              <SelectContent>
                {appCodes.map((code) => (
                  <SelectItem key={code} value={code}>
                    {code}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <div>
          <div className="flex items-center space-x-2">
            <Label>
              Namespace
              <span className="text-red-500 ml-1">*</span>
            </Label>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <HelpCircle className="h-4 w-4 text-gray-400" />
                </TooltipTrigger>
                <TooltipContent>
                  <p className="max-w-xs">Enter the Kubernetes namespace</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          <div className="relative">
            <Input
              required
              value={source.namespace}
              onChange={(e) => handleNamespaceChange(e.target.value)}
              placeholder="Enter namespace"
              className={`${
                !validation.namespace.isValid && source.namespace
                  ? 'border-red-500 focus:ring-red-500'
                  : validation.namespace.isValid && source.namespace
                  ? 'border-green-500 focus:ring-green-500'
                  : ''
              }`}
            />
            {!validation.namespace.isValid && source.namespace && (
              <div className="flex items-center mt-1 text-sm text-red-500">
                <AlertCircle className="h-4 w-4 mr-1" />
                <span>{validation.namespace.message}</span>
              </div>
            )}
            {validation.namespace.isValid && source.namespace && (
              <div className="absolute right-3 top-2.5 h-5 w-5 text-green-500">
                <svg viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              </div>
            )}
          </div>
        </div>

        <div>
          <div className="flex items-center space-x-2">
            <Label>Description</Label>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <HelpCircle className="h-4 w-4 text-gray-400" />
                </TooltipTrigger>
                <TooltipContent>
                  <p className="max-w-xs">Enter a description for your application</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          <Textarea
            value={source.description || ''}
            onChange={(e) => setSource(prev => ({ ...prev, description: e.target.value }))}
            placeholder="Optional: Enter description for your application"
            className="h-24"
          />
        </div>

        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium">External Secrets Integration</h3>
            <Switch
              checked={source.externalSecrets?.enabled || false}
              onCheckedChange={(enabled) => setSource(prev => ({
                ...prev,
                externalSecrets: {
                  enabled,
                }
              }))}
            />
          </div>

          {source.externalSecrets?.enabled && (
            <div>
              <Label>Secret Store</Label>
              <Select
                value={source.externalSecrets?.secretStore || ''}
                onValueChange={(value) => setSource(prev => ({
                  ...prev,
                  externalSecrets: {
                    enabled: true,
                    secretStore: value,
                    ...prev.externalSecrets,
                  }
                }))}
              >
                <SelectTrigger className="w-full bg-background">
                  <div className="flex items-center justify-between">
                    <SelectValue placeholder="Select Secret Store" />
                    <span className="text-xs text-muted-foreground font-mono">
                      {secretStoreList.find(store => store.id === source.externalSecrets?.secretStore)?.path}
                    </span>
                  </div>
                </SelectTrigger>
                <SelectContent>
                  {secretStoreList.map((store) => (
                    <SelectItem
                      key={store.id}
                      value={store.id || 'default'}
                      className="relative flex w-full cursor-default select-none items-center rounded-sm py-2 pl-3 pr-8 text-sm outline-none hover:bg-accent hover:text-accent-foreground"
                    >
                      <div className="flex items-center justify-between w-full">
                        <span className="font-medium">{store.name}</span>
                        <div className="flex items-center space-x-2">
                          <span className={`px-2 py-0.5 text-xs rounded-md ${
                            store.provider === 'aws' ? 'bg-yellow-100 text-yellow-800' :
                            store.provider === 'vault' ? 'bg-blue-100 text-blue-800' :
                            'bg-purple-100 text-purple-800'
                          }`}>
                            {store.provider.toUpperCase()}
                          </span>
                          <span className="px-2 py-0.5 text-xs rounded-md bg-gray-100 text-gray-800">
                            {store.type}
                          </span>
                        </div>
                        <div className="text-xs text-gray-500 font-mono mt-1">
                          Path: {store.path}
                        </div>
                      </div>
                      <span className="absolute right-2 flex h-3.5 w-3.5 items-center justify-center">
                        {store.id === source.externalSecrets?.secretStore && (
                          <span aria-hidden="true">
                            <svg width="15" height="15" viewBox="0 0 15 15" fill="none" xmlns="http://www.w3.org/2000/svg" className="h-4 w-4">
                              <path d="M11.4669 3.72684C11.7558 3.91574 11.8369 4.30308 11.648 4.59198L7.39799 11.092C7.29783 11.2452 7.13556 11.3467 6.95402 11.3699C6.77247 11.3931 6.58989 11.3355 6.45446 11.2124L3.70446 8.71241C3.44905 8.48022 3.43023 8.08494 3.66242 7.82953C3.89461 7.57412 4.28989 7.55529 4.5453 7.78749L6.75292 9.79441L10.6018 3.90792C10.7907 3.61902 11.178 3.53795 11.4669 3.72684Z" fill="currentColor" fillRule="evenodd" clipRule="evenodd" />
                            </svg>
                          </span>
                        )}
                      </span>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}
        </div>

        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Label>Ingress</Label>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <HelpCircle className="h-4 w-4 text-gray-400" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p className="max-w-xs">Configure ingress rules for your application</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={addIngressRule}
              className="flex items-center space-x-2"
            >
              <Plus className="h-4 w-4" />
              <span>Add Ingress</span>
            </Button>
          </div>

          {source.ingress?.map((rule, index) => (
            <div key={index} className="grid grid-cols-3 gap-4 items-start">
              <div>
                <Label>Name</Label>
                <Input
                  value={rule.name}
                  onChange={(e) => updateIngressRule(index, 'name', e.target.value)}
                  placeholder="Enter name"
                />
              </div>
              <div>
                <Label>Service</Label>
                {availableServices.length > 0 ? (
                  <Select
                    value={rule.service}
                    onValueChange={(value) => handleServiceChange(index, value)}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select service" />
                    </SelectTrigger>
                    <SelectContent>
                      {availableServices.map((service) => (
                        <SelectItem key={service} value={service}>
                          {service}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                ) : (
                  <Input
                    value={rule.service}
                    onChange={(e) => updateIngressRule(index, 'service', e.target.value)}
                    placeholder="Enter service"
                    className="text-muted-foreground"
                  />
                )}
              </div>
              <div className="flex items-end space-x-2">
                <div className="flex-1">
                  <Label>Port</Label>
                  <Input
                    value={rule.port}
                    onChange={(e) => updateIngressRule(index, 'port', e.target.value)}
                    placeholder="Enter port"
                  />
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => removeIngressRule(index)}
                  className="mb-[2px] hover:bg-destructive/10 hover:text-destructive"
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}