"use client";

import { DestinationCluster } from '@/app/dashboard/components/destination';
import { PageContainer } from '@/app/dashboard/components/PageContainer';

export default function DestinationClusterPage() {
  return (
    <PageContainer title="Destination Clusters">
      <DestinationCluster />
    </PageContainer>
  );
}