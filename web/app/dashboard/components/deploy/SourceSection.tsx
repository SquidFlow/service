import { Card, CardHeader, CardContent, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { GitBranch, CheckCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useDeployForm } from './DeployFormContext';
import { useState, useEffect } from "react";

export function SourceSection() {
  const { source, setSource } = useDeployForm();
  const [isValid, setIsValid] = useState(false);

  useEffect(() => {
    const hasAllFields = Boolean(
      source.url &&
      source.path &&
      source.targetRevision
    );
    setIsValid(hasAllFields);
  }, [source.url, source.path, source.targetRevision]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-3">
          <GitBranch className="h-6 w-6 text-blue-500" />
          <span>Application Source</span>
        </CardTitle>
        <p className="text-sm text-muted-foreground">
          Specify the Git repository details where your application manifests are stored
        </p>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <Label>Git Repository URL</Label>
          <Input
            value={source.url}
            onChange={(e) => setSource(prev => ({ ...prev, url: e.target.value }))}
            placeholder="e.g., https://github.com/username/repo"
          />
        </div>
        <div>
          <Label>Path</Label>
          <Input
            value={source.path}
            onChange={(e) => setSource(prev => ({ ...prev, path: e.target.value }))}
            placeholder="Path to the application manifests"
          />
        </div>
        <div>
          <Label>Target Revision</Label>
          <Input
            value={source.targetRevision}
            onChange={(e) => setSource(prev => ({ ...prev, targetRevision: e.target.value }))}
            placeholder="e.g., main, HEAD, v1.0.0"
          />
        </div>
        <div className="flex justify-end pt-4">
          <Button
            variant={isValid ? "default" : "secondary"}
            className={`flex items-center space-x-2 transition-colors ${
              isValid
                ? "bg-green-500 hover:bg-green-600 text-white"
                : "bg-gray-200 text-gray-500 cursor-not-allowed"
            }`}
            disabled={!isValid}
          >
            <CheckCircle className="h-4 w-4" />
            <span>Validate App Source</span>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}