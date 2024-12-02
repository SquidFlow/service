interface PageContainerProps {
  title: string;
  children: React.ReactNode;
}

export function PageContainer({ title, children }: PageContainerProps) {
  return (
    <div className="flex-1 space-y-4 p-8">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">{title}</h2>
      </div>
      {children}
    </div>
  );
}