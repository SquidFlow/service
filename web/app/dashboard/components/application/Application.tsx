"use client";

import { useState, useEffect } from "react";
import { useRouter, usePathname } from 'next/navigation';
import { ApplicationList } from './ApplicationList';
import { ApplicationDetail } from './ApplicationDetail';
import { DeployForm } from "../deploy";
import { useApplicationStore } from '@/store';
import type { ApplicationTemplate } from '@/types';

export function Application() {
  const [isCreating, setIsCreating] = useState(false);
  const router = useRouter();
  const pathname = usePathname();
  const { fetch: fetchApplications } = useApplicationStore();

  // 从 URL 中提取应用名称
  const appName = pathname.startsWith('/dashboard/deploy/application/')
    ? pathname.split('/').pop()
    : null;

  useEffect(() => {
    fetchApplications();
  }, [fetchApplications]);

  if (isCreating) {
    return (
      <div className="w-full max-w-4xl mx-auto space-y-4">
        <DeployForm onCancel={() => setIsCreating(false)} />
      </div>
    );
  }

  // 如果 URL 包含应用名称，显示详情页
  if (appName && appName !== 'application') {
    return <ApplicationDetail name={appName} />;
  }

  // 否则显示列表页
  return (
    <ApplicationList
      onSelectApp={(app: ApplicationTemplate) => {
        router.push(`/dashboard/deploy/application/${app.name}`);
      }}
      onCreateNew={() => setIsCreating(true)}
    />
  );
}