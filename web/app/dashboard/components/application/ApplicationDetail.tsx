"use client";

import { useEffect, useState } from 'react';
import { Button } from "@/components/ui/button";
import { ChevronLeft } from "lucide-react";
import { useRouter } from 'next/navigation';
import { DeploymentStatus } from './DeploymentStatus';
import { GeneralInfo } from './GeneralInfo';
import { ResourceMetrics } from './ResourceMetrics';
import { useApplicationStore } from '@/store';
import type { ApplicationTemplate } from '@/types';
import { useToast } from "@/components/ui/use-toast";

interface ApplicationDetailProps {
  name: string;  // 从路由参数获取
}

export function ApplicationDetail({ name }: ApplicationDetailProps) {
  const router = useRouter();
  const [app, setApp] = useState<ApplicationTemplate | null>(null);
  const { getApplicationDetail } = useApplicationStore();
  const { toast } = useToast();

  useEffect(() => {
    const fetchAppDetail = async () => {
      try {
        const data = await getApplicationDetail(name);
        setApp(data);
      } catch (error) {
        console.error('Failed to fetch application details:', error);
        toast({
          variant: "destructive",
          title: "Error",
          description: "Failed to load application details",
        });
      }
    };

    if (name) {
      fetchAppDetail();
    }
  }, [name, getApplicationDetail, toast]);

  if (!app) {
    return (
      <div className="flex items-center justify-center h-[50vh]">
        <div className="text-center space-y-4">
          <div className="text-lg font-medium">Loading application details...</div>
          <Button variant="outline" onClick={() => router.push('/dashboard/deploy/application')}>
            <ChevronLeft className="mr-2 h-4 w-4" />
            Back to List
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-4">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">
          {app.application_instantiation.application_name}
        </h1>
        <Button variant="outline" onClick={() => router.push('/dashboard/deploy/application')}>
          <ChevronLeft className="mr-2 h-4 w-4" />
          Back to List
        </Button>
      </div>

      <DeploymentStatus app={app} />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-6">
        <GeneralInfo app={app} />
        <ResourceMetrics app={app} />
      </div>
    </div>
  );
}