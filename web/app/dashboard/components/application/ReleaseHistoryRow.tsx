import { TableRow, TableCell } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { History, ChevronRight, RefreshCw, MoreHorizontal } from "lucide-react";
import { ReleaseHistory, EnvironmentType } from '@/types/release';

interface ReleaseHistoryRowProps {
  release: ReleaseHistory;
  environments: {
    SIT: ReleaseHistory[];
    UAT: ReleaseHistory[];
    PRD: ReleaseHistory[];
  };
  onRollback: (env: EnvironmentType, commitHash: string) => void;
  onPromote: (fromEnv: EnvironmentType, toEnv: EnvironmentType, commitHash: string) => void;
  onRedeploy: (env: EnvironmentType, commitHash: string) => void;
}

export function ReleaseHistoryRow({
  release,
  environments,
  onRollback,
  onPromote,
  onRedeploy,
}: ReleaseHistoryRowProps) {
  const getEnvironmentBadge = (env: EnvironmentType) => {
    const colors = {
      SIT: "bg-blue-50 text-blue-600",
      UAT: "bg-purple-50 text-purple-600",
      PRD: "bg-green-50 text-green-600",
    };
    return environments[env].find(r => r.commitHash === release.commitHash)?.isCurrent && (
      <Badge variant="outline" className={colors[env]}>@{env}</Badge>
    );
  };

  return (
    <TableRow>
      <TableCell>
        <div className="space-y-2">
          <div className="font-medium">{release.commitLog}</div>
          <div className="flex gap-2">
            {getEnvironmentBadge("SIT")}
            {getEnvironmentBadge("UAT")}
            {getEnvironmentBadge("PRD")}
          </div>
        </div>
      </TableCell>
      <TableCell>
        <a
          href={release.commitUrl ?? '#'}
          target="_blank"
          rel="noopener noreferrer"
          className="group flex items-center space-x-1"
        >
          <code className="px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded font-mono text-xs">
            {release.commitHash.substring(0, 7)}
          </code>
        </a>
      </TableCell>
      <TableCell>
        <div className="flex items-center space-x-2">
          <span>{release.commitAuthor}</span>
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center space-x-2">
          <span>{release.operator}</span>
        </div>
      </TableCell>
      <TableCell>
        <time className="text-sm text-gray-500">
          {new Date(release.releaseDate).toLocaleString()}
        </time>
      </TableCell>
      <TableCell>
        <Badge
          variant={release.status === 'success' ? 'default' : 'destructive'}
          className="capitalize"
        >
          {release.status}
        </Badge>
      </TableCell>
      <TableCell>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              onClick={() => onRedeploy("SIT", release.commitHash)}
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Redeploy
            </DropdownMenuItem>
            {environments.SIT.includes(release) && (
              <DropdownMenuItem
                onClick={() => onPromote("SIT", "UAT", release.commitHash)}
              >
                <ChevronRight className="h-4 w-4 mr-2" />
                Promote to UAT
              </DropdownMenuItem>
            )}
            {environments.UAT.includes(release) && (
              <DropdownMenuItem
                onClick={() => onPromote("UAT", "PRD", release.commitHash)}
              >
                <ChevronRight className="h-4 w-4 mr-2" />
                Promote to PRD
              </DropdownMenuItem>
            )}
            {!release.isCurrent && (
              <DropdownMenuItem
                onClick={() => onRollback("SIT", release.commitHash)}
              >
                <History className="h-4 w-4 mr-2" />
                Rollback to this version
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  );
}