"use client";

import { ClusterManager } from '@/app/dashboard/components/cluster';
import { PageContainer } from '@/app/dashboard/components/PageContainer';

export default function ClusterPage() {
  return (
    <PageContainer title="Clusters">
      <ClusterManager />
    </PageContainer>
  );
}