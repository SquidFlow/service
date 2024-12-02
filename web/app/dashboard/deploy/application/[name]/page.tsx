import { ApplicationDetail } from '@/app/dashboard/components/application/ApplicationDetail';

export default function ApplicationDetailPage({ params }: { params: { name: string } }) {
  return <ApplicationDetail name={params.name} />;
}