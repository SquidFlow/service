"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

export function RenderSettings() {
  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <p className="text-muted-foreground">
          Configure global settings and preferences
        </p>
      </div>

      <div className="grid gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Notifications</CardTitle>
            <CardDescription>
              Configure how you want to receive notifications
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-6">
            <div className="flex items-center justify-between space-x-2">
              <Label htmlFor="deployment-notifications">Deployment Notifications</Label>
              <Switch id="deployment-notifications" />
            </div>
            <div className="flex items-center justify-between space-x-2">
              <Label htmlFor="sync-notifications">Sync Status Notifications</Label>
              <Switch id="sync-notifications" />
            </div>
            <div className="flex items-center justify-between space-x-2">
              <Label htmlFor="health-notifications">Health Status Notifications</Label>
              <Switch id="health-notifications" />
            </div>
          </CardContent>
        </Card>

        {/* 其他设置卡片 */}
      </div>
    </div>
  );
}