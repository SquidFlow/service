import useSWR from "swr";
import requestor from "@/requestor";
import { useState } from "react";
import { SecretStore } from "./dashboard/components/securityMock";
import { ClusterInfo } from "./dashboard/components/destinationClusterMock";
import { TenantInfo } from "./dashboard/components/mockData";
import { Kustomization } from "./dashboard/components/applicationTemplateMock";

const ARGOCDAPPLICATIONS = "/api/v1/deploy/argocdapplications";
const TEMPLATES = "/api/v1/applications/templates";
const TENANTS = "/api/v1/tenants";
const CLUSTER = "/api/v1/destinationCluster";
const SECRETSTORE = "/api/v1/security/externalsecrets/secretstore";
export interface ApplicationTemplate {
	id: number;
	name: string;
	owner: string;
	description?: string;
	path: string;
	environments: string[];
	appType: "kustomize" | "helm" | "helm+kustomize";
	source: {
		url: string;
		targetRevision: string;
	};
	uri: string;
	lastUpdate: string;
	creator: string;
	lastUpdater: string;
	lastCommitId: string;
	lastCommitLog: string;
	podCount: number;
	cpuCount: string;
	memoryUsage: string;
	storageUsage: string;
	memoryAmount: string;
	secretCount: number;
	status: "Synced" | "OutOfSync" | "Unknown" | "Progressing" | "Degraded";
	resources: {
		[cluster: string]: {
			cpu: string;
			memory: string;
			storage: string;
			pods: number;
		};
	};
	deploymentStats: {
		deployments: number;
		services: number;
		configmaps: number;
	};
	worklog: Array<{
		date: string;
		action: string;
		user: string;
	}>;
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
	deployedEnvironments: string[];
	health: {
		status: "Healthy" | "Degraded" | "Progressing" | "Suspended" | "Missing";
		message?: string;
	};
	argocdUrl: string;
	events: Array<{
		time: string;
		type: string;
	}>;
	metadata: {
		createdAt: string;
		updatedAt: string;
		version: string;
	};
}

interface ApplicationParams {
	id?: number;
	name?: string;
	project?: string;
	appType?: string;
	owner?: string;
	validated?: string;
}

interface ApplicationResponse {
	success: boolean;
	total: number;
	apps: ApplicationTemplate[];
}

// Add new interface for validate payload
interface ValidatePayload {
	templateSource: string;
	targetRevision: string;
	path: string;
	// name: string;
	// path: string;
	// owner: string;
	// source: {
	// 	url: string;
	// 	targetRevision: string;
	// 	type: "git";
	// };
	// appType: "kustomize" | "helm" | "helm+kustomize";
	// description?: string;
}

export const useApplications = (params: ApplicationParams) => {
	const { data, error } = useSWR<ApplicationResponse>(
		ARGOCDAPPLICATIONS,
		async (url: string) => {
			const response = await requestor.get<ApplicationResponse>(url, {
				params,
			});
			return response.data;
		}
	);

	const applications = data?.apps || [];
	return {
		applications,
		error,
	};
};

export const useKustomizationsData = () => {
	const { data, error } = useSWR(TEMPLATES, async (url: string) => {
		const response = await requestor.get(url);
		return response.data;
	});

	const kustomizationsData: Kustomization[] = data?.items || [];
	return {
		kustomizationsData,
		error,
	};
};

export const usePostValidate = () => {
	const [isValidating, setIsValidating] = useState(false);
	const [error, setError] = useState<Error | null>(null);
	const [data, setData] = useState<unknown | null>(null);

	const triggerValidate = async (payload: ValidatePayload) => {
		setIsValidating(true);
		setError(null);
		try {
			const response = await requestor.post(
				`${ARGOCDAPPLICATIONS}/validate`,
				payload
			);
			setData(response.data);
			return response.data; // 返回接口数据
		} catch (err) {
			setError(err instanceof Error ? err : new Error("Unknown error"));
		} finally {
			setIsValidating(false);
		}
	};

	return {
		data,
		error,
		isValidating,
		triggerValidate,
	};
};

export const useGetAvailableTenants = () => {
	const { data, error } = useSWR(TENANTS, async (url: string) => {
		const response = await requestor.get(url);
		return response.data;
	});

	const availableTenants: TenantInfo[] = data?.projects || [];
	return {
		availableTenants,
		error,
	};
};

export const useGetClusterList = () => {
	const { data, error } = useSWR(CLUSTER, async (url) => {
		const response = await requestor.get(url);
		return response.data;
	});

	const clusterList: ClusterInfo[] = data || [];
	return {
		clusterList,
		error,
	};
};

export const useGetSecretStore = () => {
	const { data, error } = useSWR(SECRETSTORE, async (url) => {
		const response = await requestor.get(url);
		return response.data;
	});

	const secretStoreList: SecretStore[] = data || [];
	return {
		secretStoreList,
		error,
	};
};

interface DryRunPayload {}

// 自定义钩子函数用于 dryrun 操作，接收 payload 作为参数
export const useDryRun = () => {
	const [isLoading, setIsLoading] = useState(false);
	const [error, setError] = useState<Error | null>(null);
	const [data, setData] = useState(null);

	const triggerDryRun = async (payload: DryRunPayload) => {
		setIsLoading(true);
		setError(null);
		try {
			const response = await requestor.post(
				`${ARGOCDAPPLICATIONS}/dryruntemplate`,
				payload
			);
			setData(response.data);
			return response.data; // 返回接口数据，方便外部使用
		} catch (err) {
			setError(err instanceof Error ? err : new Error("Unknown error"));
		} finally {
			setIsLoading(false);
		}
	};

	return {
		data,
		error,
		isLoading,
		triggerDryRun,
	};
};
