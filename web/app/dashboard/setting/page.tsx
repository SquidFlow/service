"use client";

import { RenderSettings } from '@/app/dashboard/components/setting';
import { PageContainer } from '@/app/dashboard/components/PageContainer';

export default function SettingPage() {
  return (
    <PageContainer title="Settings">
      <RenderSettings />
    </PageContainer>
  );
}