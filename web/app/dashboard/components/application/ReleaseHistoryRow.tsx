import { TableCell, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { GitCommit } from "lucide-react";
import type { ReleaseHistory } from '@/types/release';

interface ReleaseHistoryRowProps {
  release: ReleaseHistory;
}

export function ReleaseHistoryRow({ release }: ReleaseHistoryRowProps) {
  const formatDate = (dateString: string) => {
    try {
      return new Date(dateString).toLocaleString();
    } catch (error) {
      console.error('Error formatting date:', error);
      return dateString;
    }
  };

  return (
    <TableRow>
      <TableCell>
        <Badge
          variant={release.status === 'success' ? 'default' : 'destructive'}
          className="capitalize"
        >
          {release.status}
        </Badge>
      </TableCell>
      <TableCell>
        <div className="flex items-center space-x-2">
          <GitCommit className="h-4 w-4 text-gray-500" />
          <a
            href={release.commitUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm font-mono text-blue-500 hover:text-blue-600"
          >
            {release.commitHash.substring(0, 7)}
          </a>
          <span className="text-sm text-gray-500">{release.commitLog}</span>
        </div>
      </TableCell>
      <TableCell>
        <span className="text-sm">{release.commitAuthor}</span>
      </TableCell>
      <TableCell>
        <span className="text-sm">{release.operator}</span>
      </TableCell>
      <TableCell>
        <span className="text-sm text-gray-500">
          {formatDate(release.releaseDate)}
        </span>
      </TableCell>
    </TableRow>
  );
}