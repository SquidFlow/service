import { Card, CardHeader, CardContent, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { GitBranch } from "lucide-react";
import { useDeployForm } from './DeployFormContext';
import { useState, useEffect } from "react";
import { ValidateSourceButton } from './ValidateSourceButton';
import { Switch } from "@/components/ui/switch";

export function SourceSection() {
  const { source, setSource } = useDeployForm();
  const [isValid, setIsValid] = useState(false);
  const [enableHelmManifest, setEnableHelmManifest] = useState(false);

  useEffect(() => {
    const hasAllFields = Boolean(
      source.url &&
      source.path &&
      source.targetRevision
    );
    setIsValid(hasAllFields);
  }, [source.url, source.path, source.targetRevision]);

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
          <Label className="text-sm font-medium">Git Repository URL</Label>
          <Input
            value={source.url}
            onChange={(e) => setSource(prev => ({ ...prev, url: e.target.value }))}
            placeholder="e.g., git@github.com:username/repo.git"
            className="mt-1.5"
          />
        </div>
        <div className="space-y-3">
          <Label className="text-sm font-medium">Path</Label>
          <Input
            value={source.path}
            onChange={(e) => setSource(prev => ({ ...prev, path: e.target.value }))}
            placeholder="Path to the application manifests"
            className="mt-1.5"
          />
        </div>
        <div className="space-y-3">
          <Label className="text-sm font-medium">Target Revision</Label>
          <Input
            value={source.targetRevision}
            onChange={(e) => setSource(prev => ({ ...prev, targetRevision: e.target.value }))}
            placeholder="e.g., main, HEAD, v1.0.0"
            className="mt-1.5"
          />
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
          <ValidateSourceButton isValid={isValid} source={source} />
        </div>
      </CardContent>
    </Card>
  );
}