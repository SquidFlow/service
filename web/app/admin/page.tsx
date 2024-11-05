"use client"

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useRouter } from 'next/navigation';
import { CheckCircle, XCircle } from 'lucide-react';
import Header from "@/app/components/header";

interface ComponentStatus {
  name: string;
  isHealthy: boolean;
  lastHeartbeat: string;
  version: string;
}

export default function AdminPage() {
  const router = useRouter();
  const [components, setComponents] = useState<ComponentStatus[]>([
    { name: 'Kubernetes', isHealthy: true, lastHeartbeat: '2023-05-01T12:00:00Z', version: 'v1.22.0' },
    { name: 'ArgoCD', isHealthy: true, lastHeartbeat: '2023-05-01T12:01:00Z', version: 'v2.3.0' },
    { name: 'ArgoWorkflow', isHealthy: false, lastHeartbeat: '2023-05-01T11:55:00Z', version: 'v3.2.0' },
  ]);

  const handleLogout = () => {
    localStorage.removeItem('isLoggedIn');
    localStorage.removeItem('userRole');
    router.push('/login');
  };

  // 模拟定期检查组件状态
  useEffect(() => {
    const interval = setInterval(() => {
      setComponents(prevComponents =>
        prevComponents.map(component => ({
          ...component,
          isHealthy: Math.random() > 0.2, // 80% 概率健康
          lastHeartbeat: new Date().toISOString()
        }))
      );
    }, 5000); // 每5秒更新一次

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="flex flex-col min-h-screen">
      <Header isLoggedIn={true} onLogout={handleLogout}/>
      <main className="flex-grow p-6">
        <h1 className="text-2xl font-bold mb-6">System Status</h1>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {components.map((component) => (
            <Card key={component.name}>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  {component.name}
                  {component.isHealthy ? (
                    <CheckCircle className="text-green-500" />
                  ) : (
                    <XCircle className="text-red-500" />
                  )}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="mb-2">
                  <strong>Status:</strong> {component.isHealthy ? 'Healthy' : 'Unhealthy'}
                </p>
                <p className="mb-2">
                  <strong>Last Heartbeat:</strong> {new Date(component.lastHeartbeat).toLocaleString()}
                </p>
                <p>
                  <strong>Version:</strong> {component.version}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>
      </main>
    </div>
  );
}