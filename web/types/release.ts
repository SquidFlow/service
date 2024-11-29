export type EnvironmentType = "SIT" | "UAT" | "PRD";

export interface ReleaseHistory {
  commitLog: string;
  commitHash: string;
  commitAuthor: string;
  operator: string;
  releaseDate: string;
  isCurrent: boolean;
  status: "success" | "failed" | "in-progress";
  deploymentDetails?: {
    duration: string;
    podReplicas: string;
    configChanges: string[];
  };
}

export type ReleaseHistories = {
  [K in EnvironmentType]: ReleaseHistory[];
};