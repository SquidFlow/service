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
      <div className="space-y-6">
        <Breadcrumb>
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
        <div className="w-full max-w-4xl mx-auto">
          <DeployForm
            onCancel={() => router.push('/dashboard/deploy/application')}
          />
        </div>
      </div>
    </PageContainer>
  );
}