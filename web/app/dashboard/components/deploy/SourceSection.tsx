import { Card, CardHeader, CardContent, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { GitBranch } from "lucide-react";
import { useDeployForm } from './DeployFormContext';
import { useState, useEffect } from "react";
import { ValidateSourceButton } from './ValidateSourceButton';
import { Switch } from "@/components/ui/switch";
import { AlertCircle } from "lucide-react";

interface ValidationState {
  url: {
    isValid: boolean;
    message?: string;
  };
  path: {
    isValid: boolean;
    message?: string;
  };
  targetRevision: {
    isValid: boolean;
    message?: string;
  };
}

export function SourceSection() {
  const { source, setSource, setAvailableServices } = useDeployForm();
  const [validation, setValidation] = useState<ValidationState>({
    url: { isValid: true },
    path: { isValid: true },
    targetRevision: { isValid: true }
  });
  const [enableHelmManifest, setEnableHelmManifest] = useState(false);

  // Git Repository URL 验证
  const validateGitUrl = (url: string) => {
    if (!url) {
      return { isValid: false, message: "Git repository URL is required" };
    }

    // 支持 HTTPS 和 Git SSH 协议
    const httpsPattern = /^https:\/\/[a-zA-Z0-9\-_.]+\/[a-zA-Z0-9\-_./]+$/;
    const gitPattern = /^git@[a-zA-Z0-9\-_.]+:[a-zA-Z0-9\-_./]+$/;

    if (!httpsPattern.test(url) && !gitPattern.test(url)) {
      return {
        isValid: false,
        message: "Invalid Git URL format. Must be HTTPS or Git SSH protocol"
      };
    }

    return { isValid: true };
  };

  // Path 验证
  const validatePath = (path: string) => {
    if (!path) {
      return { isValid: false, message: "Path is required" };
    }

    // 支持相对路径和绝对路径
    const pathPattern = /^[a-zA-Z0-9\-_./]+$/;
    if (!pathPattern.test(path)) {
      return {
        isValid: false,
        message: "Invalid path format. Only alphanumeric characters, dash, underscore, dot, and forward slash are allowed"
      };
    }

    return { isValid: true };
  };

  // Target Revision 验证
  const validateTargetRevision = (revision: string) => {
    if (!revision) {
      return { isValid: false, message: "Target revision is required" };
    }

    // 支持分支名称、commit hash 和 tag
    const branchPattern = /^[a-zA-Z0-9\-_./]+$/;  // 分支名称
    const hashPattern = /^[0-9a-f]{7,40}$/;       // commit hash
    const tagPattern = /^v?\d+\.\d+\.\d+(?:-[a-zA-Z0-9]+)?$/;  // tag (如 v1.0.0 或 1.0.0-beta)

    if (!branchPattern.test(revision) && !hashPattern.test(revision) && !tagPattern.test(revision)) {
      return {
        isValid: false,
        message: "Invalid revision format. Must be a branch name, commit hash, or version tag"
      };
    }

    return { isValid: true };
  };

  // 处理输入变化
  const handleUrlChange = (value: string) => {
    const urlValidation = validateGitUrl(value);
    setValidation(prev => ({
      ...prev,
      url: urlValidation
    }));
    setSource(prev => ({ ...prev, url: value }));
  };

  const handlePathChange = (value: string) => {
    const pathValidation = validatePath(value);
    setValidation(prev => ({
      ...prev,
      path: pathValidation
    }));
    setSource(prev => ({ ...prev, path: value }));
  };

  const handleRevisionChange = (value: string) => {
    const revisionValidation = validateTargetRevision(value);
    setValidation(prev => ({
      ...prev,
      targetRevision: revisionValidation
    }));
    setSource(prev => ({ ...prev, targetRevision: value }));
  };

  const handleHelmManifestToggle = (checked: boolean) => {
    setEnableHelmManifest(checked);
    if (!checked) {
      // clear helm_manifest_path if disabled
      setSource(prev => ({
        ...prev,
        application_specifier: undefined
      }));
    }
  };

  const handleHelmManifestPathChange = (path: string) => {
    setSource(prev => ({
      ...prev,
      application_specifier: {
        helm_manifest_path: path
      }
    }));
  };

  return (
    <Card>
      <CardHeader className="pb-6">
        <CardTitle className="flex items-center space-x-3 text-xl">
          <GitBranch className="h-6 w-6 text-blue-500" />
          <span>Application Source</span>
        </CardTitle>
        <p className="text-sm text-muted-foreground mt-2">
          Specify the Git repository details where your application manifests are stored
        </p>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-3">
          <Label className="text-sm font-medium">
            Git Repository URL
            <span className="text-red-500 ml-1">*</span>
          </Label>
          <Input
            required
            value={source.url}
            onChange={(e) => handleUrlChange(e.target.value)}
            placeholder="e.g., https://github.com/username/repo.git"
            className={`${
              !validation.url.isValid && source.url
                ? 'border-red-500 focus:ring-red-500'
                : validation.url.isValid && source.url
                ? 'border-green-500 focus:ring-green-500'
                : ''
            }`}
          />
          {!validation.url.isValid && source.url && (
            <div className="flex items-center mt-1 text-sm text-red-500">
              <AlertCircle className="h-4 w-4 mr-1" />
              <span>{validation.url.message}</span>
            </div>
          )}
        </div>

        <div className="space-y-3">
          <Label className="text-sm font-medium">Path</Label>
          <Input
            value={source.path}
            onChange={(e) => handlePathChange(e.target.value)}
            placeholder="Path to the application manifests (default: /)"
            className={`${
              !validation.path.isValid && source.path
                ? 'border-red-500 focus:ring-red-500'
                : validation.path.isValid && source.path
                ? 'border-green-500 focus:ring-green-500'
                : ''
            }`}
          />
          {!validation.path.isValid && source.path && (
            <div className="flex items-center mt-1 text-sm text-red-500">
              <AlertCircle className="h-4 w-4 mr-1" />
              <span>{validation.path.message}</span>
            </div>
          )}
        </div>

        <div className="space-y-3">
          <Label className="text-sm font-medium">Target Revision</Label>
          <Input
            value={source.targetRevision}
            onChange={(e) => handleRevisionChange(e.target.value)}
            placeholder="Branch, tag, or commit (default: master)"
            className={`${
              !validation.targetRevision.isValid && source.targetRevision
                ? 'border-red-500 focus:ring-red-500'
                : validation.targetRevision.isValid && source.targetRevision
                ? 'border-green-500 focus:ring-green-500'
                : ''
            }`}
          />
          {!validation.targetRevision.isValid && source.targetRevision && (
            <div className="flex items-center mt-1 text-sm text-red-500">
              <AlertCircle className="h-4 w-4 mr-1" />
              <span>{validation.targetRevision.message}</span>
            </div>
          )}
        </div>

        {/* Helm Manifest Path Configuration */}
        <div className="space-y-3 pt-4 border-t">
          <div className="flex items-center justify-between">
            <Label className="text-sm font-medium">Enable Helm Manifest Path</Label>
            <Switch
              checked={enableHelmManifest}
              onCheckedChange={handleHelmManifestToggle}
            />
          </div>
          {enableHelmManifest && (
            <div className="space-y-3">
              <Label className="text-sm font-medium">Helm Manifest Path</Label>
              <Input
                value={source.application_specifier?.helm_manifest_path || ''}
                onChange={(e) => handleHelmManifestPathChange(e.target.value)}
                placeholder="e.g., manifests/4.0.0"
                className="mt-1.5"
              />
              <p className="text-sm text-muted-foreground">
                Specify the path to your Helm manifests within the repository
              </p>
            </div>
          )}
        </div>

        <div className="flex justify-end pt-6">
          <ValidateSourceButton
            isValid={validation.url.isValid && validation.path.isValid && validation.targetRevision.isValid}
            source={source}
            onServiceListUpdate={setAvailableServices}
          />
        </div>
      </CardContent>
    </Card>
  );
}