"use client";

import { useRouter } from 'next/navigation';
import { DeployForm } from "@/app/dashboard/components/deploy/DeployForm";
import { PageContainer } from "@/app/dashboard/components/PageContainer";
import {
  Breadcrumb,
  BreadcrumbList,
  BreadcrumbItem,
  BreadcrumbLink
} from "@/components/ui/breadcrumb";

export default function NewApplicationPage() {
  const router = useRouter();

  return (
    <PageContainer title="New Application">
      <div className="max-w-6xl mx-auto space-y-8">
        <div className="rounded-lg border bg-card shadow-sm">
          <Breadcrumb className="p-4 border-b bg-muted/30">
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbLink onClick={() => router.push("/dashboard/deploy/application")}>
                  Applications
                </BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbItem>
                <BreadcrumbLink>
                  New Application
                </BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          <div className="p-6">
            <DeployForm
              onCancel={() => router.push('/dashboard/deploy/application')}
            />
          </div>
        </div>
      </div>
    </PageContainer>
  );
}