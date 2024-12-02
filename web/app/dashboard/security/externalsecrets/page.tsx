"use client";

import { ExternalSecrets } from '@/app/dashboard/components/security/ExternalSecrets';
import { PageContainer } from '@/app/dashboard/components/PageContainer';

export default function ExternalSecretsPage() {
  return (
    <PageContainer title="External Secrets">
      <ExternalSecrets />
    </PageContainer>
  );
}