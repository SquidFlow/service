import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";

interface TenantInfoProps {
  tenantInfo: {
    name: string;
    id: string;
    type: string;
    createdAt: string;
  };
}

export function TenantInfo({ tenantInfo }: TenantInfoProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Tenant Information</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <Label>Tenant Name</Label>
            <span className="text-sm font-medium">{tenantInfo.name}</span>
          </div>
          <div className="flex items-center justify-between">
            <Label>Tenant ID</Label>
            <span className="text-sm font-mono">{tenantInfo.id}</span>
          </div>
          <div className="flex items-center justify-between">
            <Label>Type</Label>
            <span className="text-sm font-medium">{tenantInfo.type}</span>
          </div>
          <div className="flex items-center justify-between">
            <Label>Created At</Label>
            <span className="text-sm">{new Date(tenantInfo.createdAt).toLocaleString()}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}