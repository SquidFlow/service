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

  if (appName && appName !== 'application') {
    return <ApplicationDetail name={appName} />;
  }

  return (
    <ApplicationList
      onSelectApp={(app: ApplicationTemplate) => {
        router.push(`/dashboard/deploy/application/${app.application_instantiation.application_name}`);
      }}
      onCreateNew={() => setIsCreating(true)}
    />
  );
}