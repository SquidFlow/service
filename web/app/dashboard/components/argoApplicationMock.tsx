import { Application } from "@/app/dashboard/interfaces";

export interface ExtendedApplication extends Application {
  status: "Synced" | "OutOfSync" | "Unknown" | "Progressing" | "Degraded";
  resources: {
    [cluster: string]: {
      cpu: string;
      memory: string;
      storage: string;
      pods: number;
    };
  };
  worklog: {
    date: string;
    action: string;
    user: string;
  }[];
  remoteRepo: {
    url: string;
    branch: string;
    baseCommitUrl: string;
    latestCommit: {
      id: string;
      message: string;
      author: string;
      timestamp: string;
    };
  };
  deployed_environments: string[];
  health: "Healthy" | "Degraded" | "Progressing" | "Suspended" | "Missing";
  argocdUrl: string;
}

// 添加发布历史的接口
interface ReleaseHistory {
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

export const releaseHistoriesData: Record<string, ReleaseHistory[]> = {
  SIT: [
    {
      commitLog: "feat(auth): add OIDC authentication support",
      commitHash: "a1b2c3d4e5f6g7h8i9j0",
      commitAuthor: "Alice Smith",
      operator: "John Doe",
      releaseDate: "2024-03-20 14:30:00",
      isCurrent: true,
      status: "success",
      deploymentDetails: {
        duration: "2m 30s",
        podReplicas: "3/3",
        configChanges: [],
      },
    },
    {
      commitLog: "fix(api): resolve memory leak in connection pool",
      commitHash: "b2c3d4e5f6g7h8i9j0k",
      commitAuthor: "Bob Johnson",
      operator: "Jane Smith",
      releaseDate: "2024-03-19 11:20:00",
      isCurrent: false,
      status: "success",
      deploymentDetails: {
        duration: "1m 45s",
        podReplicas: "3/3",
        configChanges: [
          "fix: implement connection pool cleanup",
          "test: add memory leak test cases",
        ],
      },
    },
    {
      commitLog: "feat(metrics): implement custom prometheus metrics",
      commitHash: "c3d4e5f6g7h8i9j0k1l",
      commitAuthor: "Charlie Wilson",
      operator: "Charlie Wilson",
      releaseDate: "2024-03-18 09:15:00",
      isCurrent: false,
      status: "success",
      deploymentDetails: {
        duration: "2m 15s",
        podReplicas: "3/3",
        configChanges: [
          "feat: add custom metrics endpoints",
          "docs: update metrics documentation",
        ],
      },
    },
    {
      commitLog: "refactor(core): optimize database queries",
      commitHash: "d4e5f6g7h8i9j0k1l2m",
      commitAuthor: "David Lee",
      operator: "Alice Smith",
      releaseDate: "2024-03-17 16:45:00",
      isCurrent: false,
      status: "success",
      deploymentDetails: {
        duration: "1m 30s",
        podReplicas: "3/3",
        configChanges: [
          "perf: implement query caching",
          "refactor: optimize SQL joins",
        ],
      },
    },
    {
      commitLog: "fix(security): update vulnerable dependencies",
      commitHash: "e5f6g7h8i9j0k1l2m3n",
      commitAuthor: "Eve Brown",
      operator: "John Doe",
      releaseDate: "2024-03-16 13:20:00",
      isCurrent: false,
      status: "success",
      deploymentDetails: {
        duration: "1m 15s",
        podReplicas: "3/3",
        configChanges: [
          "fix: upgrade dependencies versions",
          "test: add security test cases",
        ],
      },
    },
    {
      commitLog: "feat(ui): implement new dashboard components",
      commitHash: "f6g7h8i9j0k1l2m3n4o",
      commitAuthor: "Frank White",
      operator: "Frank White",
      releaseDate: "2024-03-15 10:30:00",
      isCurrent: false,
      status: "success",
      deploymentDetails: {
        duration: "2m 00s",
        podReplicas: "3/3",
        configChanges: [
          "feat: add new chart components",
          "style: update theme colors",
        ],
      },
    },
    {
      commitLog: "chore(deps): upgrade kubernetes client version",
      commitHash: "g7h8i9j0k1l2m3n4o5p",
      commitAuthor: "Grace Taylor",
      operator: "Bob Johnson",
      releaseDate: "2024-03-14 15:45:00",
      isCurrent: false,
      status: "success",
      deploymentDetails: {
        duration: "1m 45s",
        podReplicas: "3/3",
        configChanges: [
          "chore: update k8s client to v1.28",
          "test: update integration tests",
        ],
      },
    },
  ],
  UAT: [
    {
      commitLog: "feat(auth): add OIDC authentication support",
      commitHash: "a1b2c3d4e5f6g7h8i9j0",
      commitAuthor: "Alice Smith",
      operator: "John Doe",
      releaseDate: "2024-03-20 14:30:00",
      isCurrent: true,
      status: "success",
      deploymentDetails: {
        duration: "2m 30s",
        podReplicas: "3/3",
        configChanges: [],
      },
    },
  ],
  PRD: [
    {
      commitLog: "feat(auth): add OIDC authentication support",
      commitHash: "a1b2c3d4e5f6g7h8i9j0",
      commitAuthor: "Alice Smith",
      operator: "John Doe",
      releaseDate: "2024-03-20 14:30:00",
      isCurrent: true,
      status: "success",
      deploymentDetails: {
        duration: "2m 30s",
        podReplicas: "3/3",
        configChanges: [],
      },
    },
  ],
};

export const applicationsData: ExtendedApplication[] = [
  {
    id: 1,
    name: "external secret",
    uri: "/apps/external secret",
    lastUpdate: "2023-04-01",
    owner: "John Doe",
    creator: "Alice Smith",
    lastUpdater: "Bob Johnson",
    lastCommitId: "abc123",
    lastCommitLog: "Updated dependencies",
    podCount: 3,
    cpuCount: "2 cores",
    memoryAmount: "4Gi",
    secretCount: 2,
    status: "Synced",
    resources: {
      SIT: {
        cpu: "200m",
        memory: "256Mi",
        storage: "1Gi",
        pods: 2,
      },
      UAT: {
        cpu: "300m",
        memory: "512Mi",
        storage: "2Gi",
        pods: 3,
      },
      PRD: {
        cpu: "500m",
        memory: "1Gi",
        storage: "5Gi",
        pods: 5,
      },
    },
    worklog: [
      { date: "2023-06-10", action: "Deployment updated", user: "John Doe" },
      { date: "2023-06-09", action: "Config map changed", user: "Jane Smith" },
    ],
    remoteRepo: {
      url: "https://github.com/org/external-secret",
      branch: "main",
      baseCommitUrl: "https://github.com/org/external-secret/commit",
      latestCommit: {
        id: "abc123def",
        message: "Update secret rotation policy",
        author: "John Doe",
        timestamp: "2024-03-20T10:30:00Z",
      },
    },
    deployed_environments: ["SIT", "UAT", "PRD"],
    health: "Healthy",
    argocdUrl: "https://argocd.example.com/applications/external-secret",
  },
  {
    id: 2,
    name: "argo-rollout",
    uri: "/apps/argo-rollout",
    lastUpdate: "2023-04-02",
    owner: "Jane Smith",
    creator: "Charlie Wilson",
    lastUpdater: "Alice Brown",
    lastCommitId: "def456",
    lastCommitLog: "Added new feature",
    podCount: 5,
    cpuCount: "4 cores",
    memoryAmount: "8Gi",
    secretCount: 3,
    status: "OutOfSync",
    resources: {
      SIT: {
        cpu: "400m",
        memory: "512Mi",
        storage: "2Gi",
        pods: 3,
      },
      UAT: {
        cpu: "500m",
        memory: "640Mi",
        storage: "2.5Gi",
        pods: 4,
      },
      PRD: {
        cpu: "600m",
        memory: "768Mi",
        storage: "3Gi",
        pods: 5,
      },
    },
    worklog: [
      { date: "2023-06-08", action: "Rollout updated", user: "Charlie Wilson" },
      { date: "2023-06-07", action: "Service changed", user: "Alice Brown" },
    ],
    remoteRepo: {
      url: "https://github.com/org/argo-rollout",
      branch: "main",
      baseCommitUrl: "https://github.com/org/argo-rollout/commit",
      latestCommit: {
        id: "def456ghi",
        message: "Add new feature",
        author: "Charlie Wilson",
        timestamp: "2024-03-19T10:30:00Z",
      },
    },
    deployed_environments: ["SIT", "UAT"],
    health: "Healthy",
    argocdUrl: "https://argocd.example.com/applications/argo-rollout",
  },
  {
    id: 3,
    name: "kube-dashboard",
    uri: "/apps/kube-dashboard",
    lastUpdate: "2023-04-03",
    owner: "Bob Johnson",
    creator: "John Doe",
    lastUpdater: "Charlie Wilson",
    lastCommitId: "ghi789",
    lastCommitLog: "Fixed bug",
    podCount: 2,
    cpuCount: "1 core",
    memoryAmount: "2Gi",
    secretCount: 1,
    status: "Progressing",
    resources: {
      SIT: {
        cpu: "100m",
        memory: "128Mi",
        storage: "512Mi",
        pods: 1,
      },
      UAT: {
        cpu: "150m",
        memory: "192Mi",
        storage: "640Mi",
        pods: 2,
      },
      PRD: {
        cpu: "200m",
        memory: "256Mi",
        storage: "768Mi",
        pods: 3,
      },
    },
    worklog: [
      { date: "2023-06-06", action: "Dashboard updated", user: "Bob Johnson" },
      { date: "2023-06-05", action: "Config changed", user: "John Doe" },
    ],
    remoteRepo: {
      url: "https://github.com/org/kube-dashboard",
      branch: "main",
      baseCommitUrl: "https://github.com/org/kube-dashboard/commit",
      latestCommit: {
        id: "ghi789jkl",
        message: "Fixed bug",
        author: "Bob Johnson",
        timestamp: "2024-03-15T10:30:00Z",
      },
    },
    deployed_environments: ["SIT", "UAT"],
    health: "Healthy",
    argocdUrl: "https://argocd.example.com/applications/kube-dashboard",
  },
  {
    id: 4,
    name: "ray",
    uri: "/apps/ray",
    lastUpdate: "2023-04-04",
    owner: "Alice Brown",
    creator: "Jane Smith",
    lastUpdater: "Bob Johnson",
    lastCommitId: "jkl012",
    lastCommitLog: "Refactored code",
    podCount: 4,
    cpuCount: "2 cores",
    memoryAmount: "4Gi",
    secretCount: 2,
    status: "Unknown",
    resources: {
      SIT: {
        cpu: "300m",
        memory: "384Mi",
        storage: "1Gi",
        pods: 2,
      },
      UAT: {
        cpu: "400m",
        memory: "480Mi",
        storage: "1.2Gi",
        pods: 3,
      },
      PRD: {
        cpu: "500m",
        memory: "576Mi",
        storage: "1.5Gi",
        pods: 4,
      },
    },
    worklog: [
      { date: "2023-06-04", action: "Ray updated", user: "Alice Brown" },
      { date: "2023-06-03", action: "Config changed", user: "Jane Smith" },
    ],
    remoteRepo: {
      url: "https://github.com/org/ray",
      branch: "main",
      baseCommitUrl: "https://github.com/org/ray/commit",
      latestCommit: {
        id: "jkl012mno",
        message: "Refactored code",
        author: "Alice Brown",
        timestamp: "2024-03-14T10:30:00Z",
      },
    },
    deployed_environments: ["SIT", "UAT"],
    health: "Healthy",
    argocdUrl: "https://argocd.example.com/applications/ray",
  },
  {
    id: 5,
    name: "tidb",
    uri: "/apps/tidb",
    lastUpdate: "2023-04-05",
    owner: "Charlie Wilson",
    creator: "Alice Smith",
    lastUpdater: "John Doe",
    lastCommitId: "mno345",
    lastCommitLog: "Improved performance",
    podCount: 6,
    cpuCount: "3 cores",
    memoryAmount: "6Gi",
    secretCount: 4,
    status: "Degraded",
    resources: {
      SIT: {
        cpu: "600m",
        memory: "768Mi",
        storage: "2Gi",
        pods: 4,
      },
      UAT: {
        cpu: "700m",
        memory: "896Mi",
        storage: "2.5Gi",
        pods: 5,
      },
      PRD: {
        cpu: "800m",
        memory: "1024Mi",
        storage: "3Gi",
        pods: 6,
      },
    },
    worklog: [
      { date: "2023-06-02", action: "TiDB updated", user: "Charlie Wilson" },
      { date: "2023-06-01", action: "Config changed", user: "Alice Smith" },
    ],
    remoteRepo: {
      url: "https://github.com/org/tidb",
      branch: "main",
      baseCommitUrl: "https://github.com/org/tidb/commit",
      latestCommit: {
        id: "mno345pqr",
        message: "Improved performance",
        author: "Charlie Wilson",
        timestamp: "2024-03-13T10:30:00Z",
      },
    },
    deployed_environments: ["SIT", "UAT"],
    health: "Healthy",
    argocdUrl: "https://argocd.example.com/applications/tidb",
  },
];
