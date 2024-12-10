"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from 'next/navigation';
import { ApplicationList } from './ApplicationList';
import { ApplicationDetail } from './ApplicationDetail';
import { useApplicationStore } from '@/store';
import type { ApplicationTemplate } from '@/types';

export function Application() {
  const router = useRouter();
  const pathname = usePathname();
  const { fetch: fetchApplications } = useApplicationStore();

  const appName = pathname.startsWith('/dashboard/deploy/application/')
    ? pathname.split('/').pop()
    : null;

  useEffect(() => {
    fetchApplications();
  }, [fetchApplications]);

  if (appName && appName !== 'application') {
    return <ApplicationDetail name={appName} />;
  }

  return (
    <ApplicationList
      onSelectApp={(app: ApplicationTemplate) => {
        router.push(`/dashboard/deploy/application/${app.application_instantiation.application_name}`);
      }}
    />
  );
}