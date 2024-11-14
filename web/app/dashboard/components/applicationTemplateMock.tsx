export interface Kustomization {
	id: number;
	name: string;
	path: string;
	validated: boolean;
	owner: string;
	environments: string[];
	lastApplied: string;
	appType: "kustomization";
	source: {
		type: "git";
		url: string;
		branch?: string;
		tag?: string;
		commit?: string;
	};
	description?: string;
	resources: {
		deployments: number;
		services: number;
		configmaps: number;
		secrets: number;
		ingresses: number;
		serviceAccounts: number;
		roles: number;
		roleBindings: number;
		networkPolicies: number;
		persistentVolumeClaims: number;
		horizontalPodAutoscalers: number;
		customResourceDefinitions: {
			externalSecrets: number;
			certificates: number;
			ingressRoutes: number;
			prometheusRules: number;
			serviceMeshPolicies: number;
			virtualServices: number;
		};
	};
	events: {
		time: string;
		type: "Normal" | "Warning";
		reason: string;
		message: string;
	}[];
}

export interface Environment {
	environment: string;
	isValid: boolean;
}

export const kustomizationsData: Kustomization[] = [
	{
		id: 1,
		name: "argocd",
		path: "./overlays/production",
		validated: true,
		owner: "DevOps Team",
		environments: ["SIT", "UAT", "PRD"],
		lastApplied: "2024-03-15T10:30:00Z",
		appType: "kustomization",
		source: {
			type: "git",
			url: "https://github.com/org/argocd",
			branch: "main",
		},
		resources: {
			deployments: 3,
			services: 2,
			configmaps: 4,
			secrets: 2,
			ingresses: 1,
			serviceAccounts: 2,
			roles: 3,
			roleBindings: 3,
			networkPolicies: 1,
			persistentVolumeClaims: 0,
			horizontalPodAutoscalers: 2,
			customResourceDefinitions: {
				externalSecrets: 1,
				certificates: 2,
				ingressRoutes: 1,
				prometheusRules: 3,
				serviceMeshPolicies: 1,
				virtualServices: 2,
			},
		},
		events: [
			{
				time: "2024-03-15T10:30:00Z",
				type: "Normal",
				reason: "Applied",
				message: "Successfully updated argocd configuration",
			},
		],
	},
	{
		id: 3,
		name: "fluent-operator",
		path: "./base",
		validated: true,
		owner: "Logging Team",
		environments: ["SIT", "UAT"],
		lastApplied: "2024-03-15T08:15:00Z",
		appType: "kustomization",
		source: {
			type: "git",
			url: "https://github.com/org/fluent-operator",
			branch: "main",
		},
		resources: {
			deployments: 1,
			services: 1,
			configmaps: 2,
			secrets: 1,
			ingresses: 0,
			serviceAccounts: 1,
			roles: 1,
			roleBindings: 1,
			networkPolicies: 0,
			persistentVolumeClaims: 0,
			horizontalPodAutoscalers: 0,
			customResourceDefinitions: {
				externalSecrets: 0,
				certificates: 1,
				ingressRoutes: 0,
				prometheusRules: 0,
				serviceMeshPolicies: 0,
				virtualServices: 0,
			},
		},
		events: [
			{
				time: "2024-03-15T08:15:00Z",
				type: "Normal",
				reason: "Updated",
				message: "Updated output configuration",
			},
		],
	},
	{
		id: 4,
		name: "vault",
		path: "./overlays/vault",
		validated: true,
		owner: "Security Team",
		environments: ["SIT", "UAT", "PRD"],
		lastApplied: "2024-03-15T07:30:00Z",
		appType: "kustomization",
		source: {
			type: "git",
			url: "https://github.com/org/vault-config",
			branch: "main",
		},
		resources: {
			deployments: 3,
			services: 2,
			configmaps: 5,
			secrets: 4,
			ingresses: 1,
			serviceAccounts: 2,
			roles: 3,
			roleBindings: 3,
			networkPolicies: 1,
			persistentVolumeClaims: 0,
			horizontalPodAutoscalers: 2,
			customResourceDefinitions: {
				externalSecrets: 1,
				certificates: 2,
				ingressRoutes: 1,
				prometheusRules: 3,
				serviceMeshPolicies: 1,
				virtualServices: 2,
			},
		},
		events: [
			{
				time: "2024-03-15T07:30:00Z",
				type: "Normal",
				reason: "Validation",
				message: "Security configuration validated",
			},
		],
	},
	{
		id: 5,
		name: "loki",
		path: "./overlays/monitoring",
		validated: true,
		owner: "Monitoring Team",
		environments: ["SIT", "UAT", "PRD"],
		lastApplied: "2024-03-14T15:30:00Z",
		appType: "kustomization",
		source: {
			type: "git",
			url: "https://github.com/org/loki-config",
			branch: "main",
		},
		resources: {
			deployments: 2,
			services: 2,
			configmaps: 3,
			secrets: 1,
			ingresses: 0,
			serviceAccounts: 1,
			roles: 1,
			roleBindings: 1,
			networkPolicies: 0,
			persistentVolumeClaims: 0,
			horizontalPodAutoscalers: 0,
			customResourceDefinitions: {
				externalSecrets: 0,
				certificates: 1,
				ingressRoutes: 0,
				prometheusRules: 0,
				serviceMeshPolicies: 0,
				virtualServices: 0,
			},
		},
		events: [
			{
				time: "2024-03-14T15:30:00Z",
				type: "Normal",
				reason: "Added",
				message: "Added Loki configuration",
			},
		],
	},
	{
		id: 6,
		name: "eck-operator",
		path: "./overlays/elastic",
		validated: false,
		owner: "Platform Team",
		environments: ["SIT", "UAT"],
		lastApplied: "2024-03-14T14:20:00Z",
		appType: "kustomization",
		source: {
			type: "git",
			url: "https://github.com/org/eck-operator",
			branch: "main",
		},
		resources: {
			deployments: 1,
			services: 1,
			configmaps: 2,
			secrets: 2,
			ingresses: 0,
			serviceAccounts: 1,
			roles: 1,
			roleBindings: 1,
			networkPolicies: 0,
			persistentVolumeClaims: 0,
			horizontalPodAutoscalers: 0,
			customResourceDefinitions: {
				externalSecrets: 0,
				certificates: 1,
				ingressRoutes: 0,
				prometheusRules: 0,
				serviceMeshPolicies: 0,
				virtualServices: 0,
			},
		},
		events: [
			{
				time: "2024-03-14T14:20:00Z",
				type: "Warning",
				reason: "Updating",
				message: "Removing ES and Kibana components",
			},
		],
	},
];
