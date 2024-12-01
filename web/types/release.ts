export type EnvironmentType = 'SIT' | 'UAT' | 'PRD';

export interface ReleaseHistory {
  commitHash: string;
  commitLog: string;
  commitAuthor: string;
  operator: string;
  releaseDate: string;
  status: 'success' | 'failed';
  isCurrent: boolean;
  commitUrl?: string;
}

export type ReleaseHistories = {
  [K in EnvironmentType]: ReleaseHistory[];
};