import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BarChart } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ApplicationTemplate } from '@/types/application';
import { renderResourceValue } from './utils';

interface ResourceMetricsProps {
  app: ApplicationTemplate;
}

export function ResourceMetrics({ app }: ResourceMetricsProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <BarChart className="h-5 w-5 text-green-500" />
          <span>Resource Metrics</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="SIT">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="SIT">SIT</TabsTrigger>
            <TabsTrigger value="UAT">UAT</TabsTrigger>
            <TabsTrigger value="PRD">PRD</TabsTrigger>
          </TabsList>
          {['SIT', 'UAT', 'PRD'].map((env) => (
            <TabsContent key={env} value={env}>
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-500">CPU Usage</span>
                  <span className="font-medium">{renderResourceValue(app, env, 'cpu')}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-500">Memory Usage</span>
                  <span className="font-medium">{renderResourceValue(app, env, 'memory')}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-500">Storage Usage</span>
                  <span className="font-medium">{renderResourceValue(app, env, 'storage')}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-500">Pod Count</span>
                  <span className="font-medium">{renderResourceValue(app, env, 'pods')}</span>
                </div>
              </div>
            </TabsContent>
          ))}
        </Tabs>
      </CardContent>
    </Card>
  );
}