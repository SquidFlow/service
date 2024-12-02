import { useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import type { ApplicationTemplate } from '@/types';
import type { EnvironmentType, ReleaseHistory as ReleaseHistoryType } from '@/types/release';
import { ReleaseHistoryRow } from './ReleaseHistoryRow';
import { useToast } from "@/components/ui/use-toast";
import { useApplicationStore } from '@/store';

interface ReleaseHistoryProps {
  app: ApplicationTemplate;
}

export function ReleaseHistory({ app }: ReleaseHistoryProps) {
  const { releaseHistories, getReleaseHistories } = useApplicationStore();
  const { toast } = useToast();

  useEffect(() => {
    const fetchReleaseHistories = async () => {
      try {
        await getReleaseHistories();
      } catch (error) {
        toast({
          variant: "destructive",
          title: "Error",
          description: "Failed to fetch release histories",
        });
      }
    };

    fetchReleaseHistories();
  }, [getReleaseHistories, toast]);

  const environments: EnvironmentType[] = ['SIT', 'UAT', 'PRD'];

  return (
    <Card className="col-span-3">
      <CardHeader>
        <CardTitle>Release History</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="SIT">
          <TabsList>
            {environments.map((env) => (
              <TabsTrigger key={env} value={env}>
                {env}
              </TabsTrigger>
            ))}
          </TabsList>
          {environments.map((env) => (
            <TabsContent key={env} value={env}>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Status</TableHead>
                    <TableHead>Commit</TableHead>
                    <TableHead>Author</TableHead>
                    <TableHead>Operator</TableHead>
                    <TableHead>Date</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {releaseHistories[env]?.map((release: ReleaseHistoryType, index: number) => (
                    <ReleaseHistoryRow key={index} release={release} />
                  ))}
                </TableBody>
              </Table>
            </TabsContent>
          ))}
        </Tabs>
      </CardContent>
    </Card>
  );
}