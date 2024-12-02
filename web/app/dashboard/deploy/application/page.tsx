"use client";

import { Application } from '@/app/dashboard/components/application';
import { PageContainer } from '@/app/dashboard/components/PageContainer';

export default function ApplicationListPage() {
  return (
    <PageContainer title="Applications">
      <Application />
    </PageContainer>
  );
}