import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useEffect, useState } from "react"

export const RenderSettings = () => {
  const [tenantInfo, setTenantInfo] = useState({
    name: '',
    id: '',
    type: '',
    createdAt: ''
  });

  useEffect(() => {
    // 从 localStorage 获取租户信息
    // 实际项目中，这里可能需要从 API 获取更详细的租户信息
    const username = localStorage.getItem('username') || '';
    const userRole = localStorage.getItem('userRole') || '';
    const mockTenantId = 'tenant-' + username.toLowerCase();

    setTenantInfo({
      name: username,
      id: mockTenantId,
      type: userRole,
      createdAt: '2024-01-01' // 这里可以替换为实际的创建时间
    });
  }, []);

  return (
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Settings</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <Label>Timezone</Label>
                <Select defaultValue="Asia/Shanghai">
                  <SelectTrigger className="w-[280px]">
                    <SelectValue placeholder="Select timezone" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="Asia/Shanghai">China Standard Time (UTC+8)</SelectItem>
                    <SelectItem value="Asia/Singapore">Singapore Time (UTC+8)</SelectItem>
                    <SelectItem value="Asia/Kolkata">India Standard Time (UTC+5:30)</SelectItem>
                    <SelectItem value="UTC">Coordinated Universal Time (UTC+0)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Tenant Information</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <Label>Tenant Name</Label>
                <span className="text-sm text-muted-foreground">{tenantInfo.name}</span>
              </div>
              <div className="flex items-center justify-between">
                <Label>Tenant ID</Label>
                <span className="text-sm text-muted-foreground">{tenantInfo.id}</span>
              </div>
              <div className="flex items-center justify-between">
                <Label>Account Type</Label>
                <span className="text-sm text-muted-foreground">{tenantInfo.type}</span>
              </div>
              <div className="flex items-center justify-between">
                <Label>Created At</Label>
                <span className="text-sm text-muted-foreground">{tenantInfo.createdAt}</span>
              </div>
            </div>
          </CardContent>
        </Card>
    </div>
  );
};
