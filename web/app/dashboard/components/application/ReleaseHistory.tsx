import { useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { History } from "lucide-react";
import { Table, TableBody, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ApplicationTemplate } from '@/types/application';
import { useGetReleaseHistories } from '@/app/api';
import { EnvironmentType, ReleaseHistory as ReleaseHistoryType } from '@/types/release';
import { ReleaseHistoryRow } from './ReleaseHistoryRow';
import { useToast } from "@/components/ui/use-toast";

interface ReleaseHistoryProps {
  app: ApplicationTemplate;
}

export function ReleaseHistory({ app }: ReleaseHistoryProps) {
  const { releaseHistories } = useGetReleaseHistories();
  const { toast } = useToast();
  const [currentCommits, setCurrentCommits] = useState<Record<EnvironmentType, string>>({
    SIT: releaseHistories.SIT[0]?.commitHash || "",
    UAT: releaseHistories.UAT[0]?.commitHash || "",
    PRD: releaseHistories.PRD[0]?.commitHash || "",
  });

  const handleRollback = useCallback((env: EnvironmentType, commitHash: string) => {
    setCurrentCommits((prev) => ({
      ...prev,
      [env]: commitHash,
    }));
    toast({
      title: "Rollback Initiated",
      description: `Rolling back ${env} to commit ${commitHash.substring(0, 7)}`,
    });
  }, [toast]);

  const handlePromote = useCallback((fromEnv: EnvironmentType, toEnv: EnvironmentType, commitHash: string) => {
    toast({
      title: "Promotion Initiated",
      description: `Promoting from ${fromEnv} to ${toEnv}`,
    });
  }, [toast]);

  const handleRedeploy = useCallback((env: EnvironmentType, commitHash: string) => {
    toast({
      title: "Redeploy Initiated",
      description: `Redeploying commit ${commitHash.substring(0, 7)} in ${env}`,
    });
  }, [toast]);

  return (
    <Card className="col-span-3">
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <History className="h-5 w-5 text-green-500" />
          <span>Release History and Rollback</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow className="bg-gray-50 dark:bg-gray-800">
              <TableHead>Commit Log</TableHead>
              <TableHead>Commit Hash</TableHead>
              <TableHead>Commit Author</TableHead>
              <TableHead>Operator</TableHead>
              <TableHead>Release Date</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="w-[100px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {releaseHistories.SIT.map((release: ReleaseHistoryType, index: number) => (
              <ReleaseHistoryRow
                key={index}
                release={release}
                environments={releaseHistories}
                onRollback={handleRollback}
                onPromote={handlePromote}
                onRedeploy={handleRedeploy}
              />
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}