import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import {
	CheckCircle,
	AlertCircle,
	Plus,
	RefreshCw,
	Trash2,
	ChevronRight,
} from "lucide-react";
// import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
	Environment,
	// kustomizationsData,
	type Kustomization,
} from "./applicationTemplateMock";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import {
	useDeleteTemplate,
	useGetTemplateDetail,
	useKustomizationsData,
	usePostValidate,
} from "@/app/api";

// const mockPathEnvironments: Record<string, string[]> = {
//   production: ['SIT', 'UAT', 'PRD'],
//   staging: ['SIT', 'UAT'],
//   development: ['SIT'],
//   overlays: ['SIT', 'UAT', 'PRD'],
//   base: ['SIT', 'UAT'],
// };

export function ApplicationTemplate() {
	const { kustomizationsData, triggerGetKustomizationsData } =
		useKustomizationsData();
	const { triggerGetTemplateDetail } = useGetTemplateDetail();
	const { deleteTemplates } = useDeleteTemplate();
	const { triggerValidate } = usePostValidate();
	const [searchTerm, setSearchTerm] = useState("");
	const [selectedAppTemplate, setSelectedAppTemplate] =
		useState<Kustomization | null>(null);
	const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
	// const [currentStep, setCurrentStep] = useState(1);

	const [formData, setFormData] = useState({
		name: "",
		description: "",
		repoUrl: "git@github.com:h4-poc/platform.git",
		branch: "main",
		path: "manifest/fluent-operator",
		refType: "branch" as "branch" | "tag" | "commit",
		appType: "kustomize" as const,
		isValidated: false,
		validation: {
			isValidating: false,
			status: "pending" as "pending" | "success" | "error",
			environments: [] as Environment[],
			message: "",
		},
	});

	// 添加选中项的状态
	const [selectedItems, setSelectedItems] = useState<string[]>([]);

	// 在组件顶部添加新的状态
	const [expandedSections, setExpandedSections] = useState<{
		core: boolean;
		network: boolean;
		rbac: boolean;
		storage: boolean;
		autoscaling: boolean;
		custom: boolean;
	}>({
		core: false,
		network: false,
		rbac: false,
		storage: false,
		autoscaling: false,
		custom: false,
	});

	// 添加新的状态控制 Resources 卡片的展开/折叠
	const [isResourcesExpanded, setIsResourcesExpanded] = useState(false);

	const filteredKustomizations = kustomizationsData.filter(
		(k) =>
			k.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
			k.path.toLowerCase().includes(searchTerm.toLowerCase())
	);

	const renderApplicationTemplateDetail = (kustomization: Kustomization) => {
		return (
			<div className="space-y-8">
				{/* Header Section */}
				<div className="flex justify-between items-center bg-white dark:bg-gray-800 p-6 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
					<div className="space-y-1">
						<h2 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-gray-100 dark:to-gray-400 bg-clip-text text-transparent">
							{kustomization.name}
						</h2>
						<p className="text-gray-500 dark:text-gray-400">
							Last applied:{" "}
							{new Date(kustomization.lastApplied).toLocaleString()}
						</p>
					</div>
					<Button
						variant="outline"
						onClick={() => setSelectedAppTemplate(null)}
						className="border-gray-200 dark:border-gray-700"
					>
						Back to List
					</Button>
				</div>

				<div className="grid grid-cols-2 gap-6">
					{/* Basic Information Card */}
					<Card className="bg-gradient-to-br from-white to-gray-50 dark:from-gray-800 dark:to-gray-900 shadow-sm border border-gray-100 dark:border-gray-700">
						<CardHeader className="border-b border-gray-100 dark:border-gray-700">
							<CardTitle className="flex items-center space-x-2 text-lg">
								<div className="h-4 w-1 bg-blue-500 rounded-full" />
								<span>Basic Information</span>
							</CardTitle>
						</CardHeader>
						<CardContent className="space-y-6 pt-6">
							<div className="grid grid-cols-2 gap-4">
								<div>
									<p className="text-sm font-medium text-gray-500 dark:text-gray-400">
										Name
									</p>
									<p className="mt-1 font-medium">{kustomization.name}</p>
								</div>
								<div>
									<p className="text-sm font-medium text-gray-500 dark:text-gray-400">
										Owner
									</p>
									<div className="mt-1 flex items-center space-x-2">
										{/* <Avatar className="h-6 w-6 bg-blue-100 dark:bg-blue-900">
											<AvatarFallback className="text-xs text-blue-700 dark:text-blue-300">
												{kustomization.owner
													.split(" ")
													.map((n) => n[0])
													.join("")}
											</AvatarFallback>
										</Avatar> */}
										<span className="font-medium">{kustomization.owner}</span>
									</div>
								</div>
							</div>
							<div>
								<p className="text-sm font-medium text-gray-500 dark:text-gray-400">
									Description
								</p>
								<p className="mt-1 text-gray-700 dark:text-gray-300">
									{kustomization.description || "No description provided"}
								</p>
							</div>
						</CardContent>
					</Card>

					{/* Source Configuration Card */}
					<Card className="bg-gradient-to-br from-white to-gray-50 dark:from-gray-800 dark:to-gray-900 shadow-sm border border-gray-100 dark:border-gray-700">
						<CardHeader className="border-b border-gray-100 dark:border-gray-700">
							<CardTitle className="flex items-center space-x-2 text-lg">
								<div className="h-4 w-1 bg-purple-500 rounded-full" />
								<span>Source Configuration</span>
							</CardTitle>
						</CardHeader>
						<CardContent className="space-y-6 pt-6">
							<div>
								<p className="text-sm font-medium text-gray-500 dark:text-gray-400">
									Application Type
								</p>
								<Badge
									variant="outline"
									className="mt-1 capitalize bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400"
								>
									{kustomization.appType}
								</Badge>
							</div>
							<div>
								<p className="text-sm font-medium text-gray-500 dark:text-gray-400">
									Repository URL
								</p>
								<a
									href={kustomization.source.url}
									target="_blank"
									rel="noopener noreferrer"
									className="mt-1 inline-flex items-center space-x-1 font-mono text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 hover:underline"
								>
									{kustomization.source.url}
								</a>
							</div>
							<div>
								<p className="text-sm font-medium text-gray-500 dark:text-gray-400">
									Reference
								</p>
								<div className="mt-1 flex items-center space-x-2">
									<Badge variant="outline" className="capitalize">
										{kustomization.source.branch
											? "branch"
											: kustomization.source.tag
												? "tag"
												: "commit"}
									</Badge>
									<a
										href={`${kustomization.source.url}/tree/${
											kustomization.source.branch ||
											kustomization.source.tag ||
											kustomization.source.commit ||
											"main"
										}`}
										target="_blank"
										rel="noopener noreferrer"
										className="font-mono text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 hover:underline"
									>
										{kustomization.source.branch ||
											kustomization.source.tag ||
											kustomization.source.commit ||
											"main"}
									</a>
								</div>
							</div>
							<div>
								<p className="text-sm font-medium text-gray-500 dark:text-gray-400">
									Path
								</p>
								<a
									href={`${kustomization.source.url}/tree/${
										kustomization.source.branch ||
										kustomization.source.tag ||
										kustomization.source.commit ||
										"main"
									}/${kustomization.path}`}
									target="_blank"
									rel="noopener noreferrer"
									className="mt-1 inline-block font-mono text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 hover:underline"
								>
									{kustomization.path}
								</a>
							</div>
						</CardContent>
					</Card>
				</div>

				{/* Application Template Adapt Environments Card */}
				<Card className="bg-gradient-to-br from-white to-gray-50 dark:from-gray-800 dark:to-gray-900 shadow-sm border border-gray-100 dark:border-gray-700">
					<CardHeader className="border-b border-gray-100 dark:border-gray-700">
						<CardTitle className="flex items-center space-x-2 text-lg">
							<div className="h-4 w-1 bg-green-500 rounded-full" />
							<span>Application Template Adapt Environments</span>
						</CardTitle>
					</CardHeader>
					<CardContent className="pt-6">
						<div className="flex flex-wrap gap-2">
							{kustomization.environments?.map((env) => (
								<Badge
									key={env}
									variant="outline"
									className="px-3 py-1 capitalize bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400"
								>
									{env}
								</Badge>
							))}
						</div>
					</CardContent>
				</Card>

				{/* Resources Card */}
				<Card className="bg-gradient-to-br from-white to-gray-50 dark:from-gray-800 dark:to-gray-900 shadow-sm border border-gray-100 dark:border-gray-700">
					<CardHeader
						className="border-b border-gray-100 dark:border-gray-700 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50"
						onClick={() => setIsResourcesExpanded(!isResourcesExpanded)}
					>
						<div className="flex items-center justify-between">
							<CardTitle className="flex items-center space-x-2 text-lg">
								<div className="h-4 w-1 bg-orange-500 rounded-full" />
								<span>Resources</span>
							</CardTitle>
							<Button variant="ghost" size="sm" className="w-8 h-8 p-0">
								{isResourcesExpanded ? (
									<div className="text-lg font-semibold text-gray-500">−</div>
								) : (
									<div className="text-lg font-semibold text-gray-500">+</div>
								)}
							</Button>
						</div>
					</CardHeader>
					{isResourcesExpanded && (
						<CardContent className="space-y-6 pt-6">
							{/* Core Resources */}
							<div>
								<button
									onClick={() =>
										setExpandedSections((prev) => ({
											...prev,
											core: !prev.core,
										}))
									}
									className="flex items-center justify-between w-full text-left mb-3 group"
								>
									<h4 className="text-sm font-medium text-gray-500">
										Core Resources
									</h4>
									<ChevronRight
										className={`h-4 w-4 text-gray-400 transition-transform duration-200 ${
											expandedSections.core ? "rotate-90" : ""
										}`}
									/>
								</button>
								{expandedSections.core && (
									<div className="grid grid-cols-4 gap-4">
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Deployments</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{kustomization.resources.deployments}
											</p>
										</div>
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Services</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{kustomization.resources.services}
											</p>
										</div>
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">ConfigMaps</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{kustomization.resources.configmaps}
											</p>
										</div>
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Secrets</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{kustomization.resources.secrets}
											</p>
										</div>
									</div>
								)}
							</div>

							{/* Network Resources */}
							<div>
								<button
									onClick={() =>
										setExpandedSections((prev) => ({
											...prev,
											network: !prev.network,
										}))
									}
									className="flex items-center justify-between w-full text-left mb-3 group"
								>
									<h4 className="text-sm font-medium text-gray-500">
										Network Resources
									</h4>
									<ChevronRight
										className={`h-4 w-4 text-gray-400 transition-transform duration-200 ${
											expandedSections.network ? "rotate-90" : ""
										}`}
									/>
								</button>
								{expandedSections.network && (
									<div className="grid grid-cols-4 gap-4">
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Ingresses</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{kustomization.resources.ingresses}
											</p>
										</div>
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Network Policies</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{kustomization.resources.networkPolicies}
											</p>
										</div>
									</div>
								)}
							</div>

							{/* RBAC Resources */}
							<div>
								<button
									onClick={() =>
										setExpandedSections((prev) => ({
											...prev,
											rbac: !prev.rbac,
										}))
									}
									className="flex items-center justify-between w-full text-left mb-3 group"
								>
									<h4 className="text-sm font-medium text-gray-500">
										RBAC Resources
									</h4>
									<ChevronRight
										className={`h-4 w-4 text-gray-400 transition-transform duration-200 ${
											expandedSections.rbac ? "rotate-90" : ""
										}`}
									/>
								</button>
								{expandedSections.rbac && (
									<div className="grid grid-cols-4 gap-4">
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Service Accounts</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{kustomization.resources.serviceAccounts}
											</p>
										</div>
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Roles</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{kustomization.resources.roles}
											</p>
										</div>
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Role Bindings</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{kustomization.resources.roleBindings}
											</p>
										</div>
									</div>
								)}
							</div>

							{/* Storage Resources */}
							<div>
								<button
									onClick={() =>
										setExpandedSections((prev) => ({
											...prev,
											storage: !prev.storage,
										}))
									}
									className="flex items-center justify-between w-full text-left mb-3 group"
								>
									<h4 className="text-sm font-medium text-gray-500">
										Storage Resources
									</h4>
									<ChevronRight
										className={`h-4 w-4 text-gray-400 transition-transform duration-200 ${
											expandedSections.storage ? "rotate-90" : ""
										}`}
									/>
								</button>
								{expandedSections.storage && (
									<div className="grid grid-cols-4 gap-4">
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">PVCs</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{kustomization.resources.persistentVolumeClaims}
											</p>
										</div>
									</div>
								)}
							</div>

							{/* Autoscaling Resources */}
							<div>
								<button
									onClick={() =>
										setExpandedSections((prev) => ({
											...prev,
											autoscaling: !prev.autoscaling,
										}))
									}
									className="flex items-center justify-between w-full text-left mb-3 group"
								>
									<h4 className="text-sm font-medium text-gray-500">
										Autoscaling Resources
									</h4>
									<ChevronRight
										className={`h-4 w-4 text-gray-400 transition-transform duration-200 ${
											expandedSections.autoscaling ? "rotate-90" : ""
										}`}
									/>
								</button>
								{expandedSections.autoscaling && (
									<div className="grid grid-cols-4 gap-4">
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">HPAs</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{kustomization.resources.horizontalPodAutoscalers}
											</p>
										</div>
									</div>
								)}
							</div>

							{/* Custom Resources */}
							<div>
								<button
									onClick={() =>
										setExpandedSections((prev) => ({
											...prev,
											custom: !prev.custom,
										}))
									}
									className="flex items-center justify-between w-full text-left mb-3 group"
								>
									<h4 className="text-sm font-medium text-gray-500">
										Custom Resources
									</h4>
									<ChevronRight
										className={`h-4 w-4 text-gray-400 transition-transform duration-200 ${
											expandedSections.custom ? "rotate-90" : ""
										}`}
									/>
								</button>
								{expandedSections.custom && (
									<div className="grid grid-cols-4 gap-4">
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">External Secrets</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{
													kustomization.resources.customResourceDefinitions
														.externalSecrets
												}
											</p>
										</div>
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Certificates</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{
													kustomization.resources.customResourceDefinitions
														.certificates
												}
											</p>
										</div>
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Ingress Routes</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{
													kustomization.resources.customResourceDefinitions
														.ingressRoutes
												}
											</p>
										</div>
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Prometheus Rules</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{
													kustomization.resources.customResourceDefinitions
														.prometheusRules
												}
											</p>
										</div>
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">
												Service Mesh Policies
											</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{
													kustomization.resources.customResourceDefinitions
														.serviceMeshPolicies
												}
											</p>
										</div>
										<div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
											<p className="text-sm text-gray-500">Virtual Services</p>
											<p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
												{
													kustomization.resources.customResourceDefinitions
														.virtualServices
												}
											</p>
										</div>
									</div>
								)}
							</div>
						</CardContent>
					)}
				</Card>

				{/* Events Card */}
				<Card className="bg-gradient-to-br from-white to-gray-50 dark:from-gray-800 dark:to-gray-900 shadow-sm border border-gray-100 dark:border-gray-700">
					<CardHeader className="border-b border-gray-100 dark:border-gray-700">
						<CardTitle className="flex items-center space-x-2 text-lg">
							<div className="h-4 w-1 bg-indigo-500 rounded-full" />
							<span>Events</span>
						</CardTitle>
					</CardHeader>
					<CardContent className="pt-6">
						<Table>
							<TableHeader>
								<TableRow className="hover:bg-transparent">
									<TableHead className="w-[200px]">Time</TableHead>
									<TableHead className="w-[100px]">Type</TableHead>
									<TableHead className="w-[150px]">Reason</TableHead>
									<TableHead>Message</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{kustomization.events.map((event, index) => (
									<TableRow
										key={index}
										className="hover:bg-gray-50/50 dark:hover:bg-gray-800/50"
									>
										<TableCell className="font-mono text-sm">
											{new Date(event.time).toLocaleString()}
										</TableCell>
										<TableCell>
											<Badge
												variant={
													event.type === "Normal" ? "default" : "destructive"
												}
												className={
													event.type === "Normal"
														? "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400"
														: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400"
												}
											>
												{event.type}
											</Badge>
										</TableCell>
										<TableCell className="font-medium">
											{event.reason}
										</TableCell>
										<TableCell className="text-gray-600 dark:text-gray-300">
											{event.message}
										</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					</CardContent>
				</Card>
			</div>
		);
	};

	const renderCreateDialog = () => {
		return (
			<Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
				<DialogContent
					className="sm:max-w-[800px]"
					style={{
						overflow: "auto",
						maxHeight: "600px",
						WebkitOverflowScrolling: "touch",
					}}
				>
					<DialogHeader>
						<DialogTitle className="text-xl font-semibold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-gray-100 dark:to-gray-400 bg-clip-text text-transparent">
							Create Application Template
						</DialogTitle>
						<DialogDescription>
							Configure your application template settings
						</DialogDescription>
					</DialogHeader>

					<div className="py-4 space-y-8">
						{/* Basic Information Section */}
						<div className="space-y-4">
							<div className="flex items-center space-x-2">
								<div className="h-8 w-1 bg-blue-500 rounded-full" />
								<h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
									Basic Information
								</h3>
							</div>
							<div className="grid grid-cols-2 gap-6 pl-6">
								<div className="space-y-2">
									<Label htmlFor="name" className="text-sm font-medium">
										Template Name
									</Label>
									<Input
										id="name"
										value={formData.name}
										onChange={(e) =>
											setFormData({ ...formData, name: e.target.value })
										}
										placeholder="Enter template name"
										className="border-gray-200 dark:border-gray-700 focus:ring-2 focus:ring-blue-500"
									/>
								</div>
								<div className="space-y-2">
									<Label htmlFor="description" className="text-sm font-medium">
										Description
									</Label>
									<Textarea
										id="description"
										value={formData.description}
										onChange={(e) =>
											setFormData({ ...formData, description: e.target.value })
										}
										placeholder="Enter template description"
										className="h-[38px] border-gray-200 dark:border-gray-700 focus:ring-2 focus:ring-blue-500"
									/>
								</div>
							</div>
						</div>

						<Separator className="my-6" />

						{/* Source Configuration Section */}
						<div className="space-y-4">
							<div className="flex items-center space-x-2">
								<div className="h-8 w-1 bg-purple-500 rounded-full" />
								<h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
									Source Configuration
								</h3>
							</div>
							<div className="space-y-6 pl-6">
								{/* Application Type Display */}
								<div className="space-y-2">
									<Label className="text-sm font-medium">
										Application Type
									</Label>
									<div className="h-[38px] flex items-center">
										<div className="px-4 py-2 bg-purple-100 text-purple-900 dark:bg-purple-900/30 dark:text-purple-300 rounded-md inline-flex items-center">
											<span className="capitalize">{formData.appType}</span>
										</div>
									</div>
								</div>

								<div className="grid grid-cols-2 gap-6">
									<div className="space-y-2">
										<Label htmlFor="repoUrl" className="text-sm font-medium">
											Repository URL
										</Label>
										<div className="h-[38px]">
											<Input
												id="repoUrl"
												value={formData.repoUrl}
												onChange={(e) =>
													setFormData({ ...formData, repoUrl: e.target.value })
												}
												placeholder="Enter repository URL"
												className="h-full font-mono text-sm border-gray-200 dark:border-gray-700 focus:ring-2 focus:ring-purple-500"
											/>
										</div>
									</div>
									<div className="space-y-2">
										<Label className="text-sm font-medium">
											Reference Type
										</Label>
										<div className="flex space-x-4 h-[38px] items-center">
											{(["branch", "tag", "commit"] as const).map((type) => (
												<div
													key={type}
													onClick={() =>
														setFormData({
															...formData,
															refType: type,
															branch: "",
														})
													}
													className={`
                            flex items-center h-full px-4 rounded-md cursor-pointer transition-colors
                            ${
															formData.refType === type
																? "bg-purple-100 text-purple-900 dark:bg-purple-900/30 dark:text-purple-300"
																: "bg-gray-50 text-gray-600 hover:bg-gray-100 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
														}
                          `}
												>
													<span className="capitalize">{type}</span>
												</div>
											))}
										</div>
									</div>
								</div>

								<div className="space-y-2">
									<Label htmlFor="branch" className="text-sm font-medium">
										{formData.refType === "branch"
											? "Branch Name"
											: formData.refType === "tag"
												? "Tag Name"
												: "Commit Hash"}
									</Label>
									<Input
										id="branch"
										value={formData.branch}
										onChange={(e) => {
											let value = e.target.value;
											if (formData.refType === "commit") {
												value = value.toLowerCase().replace(/[^0-9a-f]/g, "");
											}
											setFormData({ ...formData, branch: value });
										}}
										placeholder={
											formData.refType === "branch"
												? "e.g., main, develop"
												: formData.refType === "tag"
													? "e.g., v1.0.0, release-1.2"
													: "e.g., 1a2b3c4d"
										}
										maxLength={formData.refType === "commit" ? 40 : undefined}
										pattern={
											formData.refType === "commit"
												? "[0-9a-f]{5,40}"
												: undefined
										}
										className="font-mono text-sm border-gray-200 dark:border-gray-700 focus:ring-2 focus:ring-purple-500"
									/>
									{formData.refType === "commit" && (
										<p className="text-xs text-gray-500 mt-1">
											Enter a valid git commit hash (5-40 characters,
											hexadecimal)
										</p>
									)}
								</div>

								<div className="space-y-2">
									<Label htmlFor="path" className="text-sm font-medium">
										Path
									</Label>
									<div className="flex space-x-2">
										<Input
											id="path"
											value={formData.path}
											onChange={(e) => {
												setFormData((prev) => ({
													...prev,
													path: e.target.value,
													isValidated: false,
													validation: {
														...prev.validation,
														status: "pending",
														environments: [],
														message: "",
													},
												}));
											}}
											placeholder="Enter path to template files"
											className="font-mono text-sm border-gray-200 dark:border-gray-700 focus:ring-2 focus:ring-purple-500"
										/>
										<Button
											onClick={() => validateTemplate()}
											disabled={
												!formData.path || formData.validation.isValidating
											}
											className={`min-w-[100px] transition-colors ${
												formData.validation.status === "success"
													? "bg-green-500 hover:bg-green-600 text-white"
													: formData.validation.status === "error"
														? "bg-red-500 hover:bg-red-600 text-white"
														: "bg-purple-500 hover:bg-purple-600 text-white"
											}`}
										>
											{formData.validation.isValidating ? (
												<>
													<RefreshCw className="h-4 w-4 animate-spin mr-2" />
													<span>Validating</span>
												</>
											) : formData.validation.status === "success" ? (
												<>
													<CheckCircle className="h-4 w-4 mr-2" />
													<span>Validated</span>
												</>
											) : formData.validation.status === "error" ? (
												<>
													<AlertCircle className="h-4 w-4 mr-2" />
													<span>Invalid</span>
												</>
											) : (
												<span>Validate</span>
											)}
										</Button>
									</div>
									{!formData.isValidated && formData.path && (
										<div className="flex items-center text-sm text-yellow-600 dark:text-yellow-400 mt-2">
											<AlertCircle className="h-4 w-4 mr-1" />
											Please validate the template path before creating
										</div>
									)}
									{formData.validation.message && (
										<div
											className={`mt-2 text-sm ${
												formData.validation.status === "success"
													? "text-green-600 dark:text-green-400"
													: formData.validation.status === "error"
														? "text-red-600 dark:text-red-400"
														: "text-gray-500"
											}`}
										>
											{formData.validation.message}
										</div>
									)}
								</div>
							</div>
						</div>

						<Separator className="my-6" />

						{/* Application Configuration Section */}
						<div className="space-y-4">
							<div className="flex items-center space-x-2">
								<div className="h-8 w-1 bg-green-500 rounded-full" />
								<h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
									Application Template Adapt Environments
								</h3>
							</div>
							<div className="space-y-6 pl-6">
								<div className="space-y-2">
									<Label className="text-sm font-medium">Environments</Label>
									<div className="min-h-[38px] p-2 bg-gray-50 dark:bg-gray-800 rounded-md">
										{formData.validation.environments.length > 0 ? (
											<div className="flex flex-wrap gap-2">
												{formData.validation.environments.map((env) => (
													<Badge
														key={env.environment}
														variant="outline"
														className="capitalize bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400"
													>
														{env.environment}
													</Badge>
												))}
											</div>
										) : (
											<div className="text-sm text-gray-500 dark:text-gray-400 py-1">
												Environments will be detected after validation
											</div>
										)}
									</div>
								</div>
							</div>
						</div>
					</div>

					<DialogFooter className="mt-6">
						<Button
							variant="outline"
							onClick={() => setIsCreateDialogOpen(false)}
							className="border-gray-200 dark:border-gray-700"
						>
							Cancel
						</Button>
						<Button
							onClick={() => {
								console.log("Form submitted:", formData);
								setIsCreateDialogOpen(false);
							}}
							disabled={
								!formData.isValidated ||
								formData.validation.status !== "success"
							}
							className={`
                ${!formData.isValidated ? "cursor-not-allowed" : ""}
                ${
									formData.isValidated &&
									formData.validation.status === "success"
										? "bg-blue-500 hover:bg-blue-600"
										: "bg-gray-300 dark:bg-gray-700"
								}
              `}
						>
							Create
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		);
	};

	// 添加选择处理函数
	const handleSelect = (id: string) => {
		setSelectedItems((prev) => {
			if (prev.includes(id)) {
				return prev.filter((item) => item !== id);
			}
			return [...prev, id];
		});
	};

	// 添加全选处理函数
	const handleSelectAll = () => {
		if (selectedItems.length === filteredKustomizations.length) {
			setSelectedItems([]);
		} else {
			setSelectedItems(filteredKustomizations.map((k) => k.id));
		}
	};

	const handleDelete = async () => {
		try {
			const result = await deleteTemplates(selectedItems);
			await triggerGetKustomizationsData();
			console.log(result);
		} catch (error) {}
	};

	// 修改 validateTemplate 函数
	const validateTemplate = async () => {
		setFormData((prev) => ({
			...prev,
			validation: {
				...prev.validation,
				isValidating: true,
				status: "pending",
				message: "Validating template...",
			},
		}));

		try {
			const detectedEnvs: Environment[] = await triggerValidate({
				templateSource: formData.repoUrl,
				targetRevision: formData.branch,
				path: formData.path,
			});

			setFormData((prev) => ({
				...prev,
				isValidated: true,
				validation: {
					isValidating: false,
					status: "success",
					environments: detectedEnvs,
					message: `Template validated successfully. Found ${detectedEnvs.length} environments: ${detectedEnvs.map((e) => e.environment).join(", ")}`,
				},
			}));
		} catch (error) {}

		// try {
		//   // 模拟API调用延迟
		//   await new Promise(resolve => setTimeout(resolve, 1000));

		//   // 查找匹配的路径
		//   const matchedPath = Object.keys(mockPathEnvironments).find(
		//     templatePath => path.toLowerCase().includes(templatePath.toLowerCase())
		//   );

		//   if (matchedPath) {
		//     // 找到匹配的路径，返回预定义的环境
		//     const detectedEnvs = mockPathEnvironments[matchedPath];
		//     setFormData(prev => ({
		//       ...prev,
		//       isValidated: true,
		//       validation: {
		//         isValidating: false,
		//         status: 'success',
		//         environments: detectedEnvs,
		//         message: `Template validated successfully. Found ${detectedEnvs.length} environments: ${detectedEnvs.join(', ')}`
		//       }
		//     }));
		//   } else {
		//     // 使用路径检测逻辑作为后备
		//     const detectedEnvs = detectEnvironments(path);

		//     if (detectedEnvs.length > 0) {
		//       setFormData(prev => ({
		//         ...prev,
		//         isValidated: true,
		//         validation: {
		//           isValidating: false,
		//           status: 'success',
		//           environments: detectedEnvs,
		//           message: `Template validated successfully. Found ${detectedEnvs.length} environments: ${detectedEnvs.join(', ')}`
		//         }
		//       }));
		//     } else {
		//       throw new Error('No environments detected in the specified path');
		//     }
		//   }
		// } catch (error) {
		//   setFormData(prev => ({
		//     ...prev,
		//     isValidated: false,
		//     validation: {
		//       isValidating: false,
		//       status: 'error',
		//       environments: [],
		//       message: 'Validation failed: ' + (error as Error).message
		//     }
		//   }));
		// }
	};

	// const detectEnvironments = (path: string): string[] => {
	// 	// 这里是示例逻辑，���际实现时需要根据���的文件结构来检测
	// 	const envPatterns = {
	// 		dev: /(dev|development)/i,
	// 		staging: /(staging|uat)/i,
	// 		prod: /(prod|production)/i,
	// 	};

	// 	const detectedEnvs: string[] = [];
	// 	Object.entries(envPatterns).forEach(([env, pattern]) => {
	// 		if (pattern.test(path)) {
	// 			detectedEnvs.push(env);
	// 		}
	// 	});

	// 	return detectedEnvs;
	// };

	if (selectedAppTemplate) {
		return renderApplicationTemplateDetail(selectedAppTemplate);
	}

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-center">
				<h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-gray-100 dark:to-gray-400 bg-clip-text text-transparent">
					ApplicationTemplate
				</h1>
				<div className="flex space-x-2">
					<Button size="sm" onClick={() => setIsCreateDialogOpen(true)}>
						<Plus className="mr-2 h-4 w-4" />
						Create
					</Button>
					<Button
						size="sm"
						variant="outline"
						className="hover:bg-blue-50 dark:hover:bg-blue-900"
						disabled={selectedItems.length === 0}
						onClick={() => console.log("Sync selected:", selectedItems)}
					>
						<RefreshCw className="mr-2 h-4 w-4" />
						Sync
					</Button>
					<Button
						size="sm"
						variant="destructive"
						className="hover:bg-red-600"
						disabled={selectedItems.length === 0}
						onClick={handleDelete}
					>
						<Trash2 className="mr-2 h-4 w-4" />
						Delete
					</Button>
				</div>
			</div>

			<Card className="bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm border-gray-200/50 dark:border-gray-700/50">
				<CardHeader>
					<CardTitle className="text-base font-medium text-gray-900 dark:text-gray-100">
						ApplicationTemplate List
					</CardTitle>
					<Input
						placeholder="Search by name, path, or owner..."
						value={searchTerm}
						onChange={(e) => setSearchTerm(e.target.value)}
						className="mt-2 border-gray-300 dark:border-gray-600 focus:ring-2 focus:ring-blue-500"
					/>
				</CardHeader>
				<CardContent>
					<Table>
						<TableHeader>
							<TableRow className="bg-gray-50 dark:bg-gray-900/50">
								<TableHead className="w-[50px]">
									<div
										className="flex items-center justify-center w-5 h-5 rounded border border-gray-300 dark:border-gray-600 cursor-pointer"
										onClick={handleSelectAll}
									>
										{selectedItems.length > 0 &&
											selectedItems.length ===
												filteredKustomizations.length && (
												<CheckCircle className="h-4 w-4 text-blue-500" />
											)}
										{selectedItems.length > 0 &&
											selectedItems.length < filteredKustomizations.length && (
												<div className="w-3 h-3 bg-blue-500 rounded-sm" />
											)}
									</div>
								</TableHead>
								<TableHead>Name</TableHead>
								<TableHead>Path</TableHead>
								<TableHead>Validated</TableHead>
								<TableHead>Owner</TableHead>
								<TableHead>Source Type</TableHead>
								<TableHead>App Type</TableHead>
								<TableHead>Environments</TableHead>
								<TableHead></TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{filteredKustomizations.map((kustomization) => (
								<TableRow
									key={kustomization.id}
									className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
								>
									<TableCell>
										<div
											className="flex items-center justify-center w-5 h-5 rounded border border-gray-300 dark:border-gray-600 cursor-pointer"
											onClick={() => handleSelect(kustomization.id)}
										>
											{selectedItems.includes(kustomization.id) && (
												<CheckCircle className="h-4 w-4 text-blue-500" />
											)}
										</div>
									</TableCell>
									<TableCell>
										<Button
											variant="link"
											onClick={async () => {
												try {
													const detail = await triggerGetTemplateDetail({
														id: kustomization.id,
													});
													setSelectedAppTemplate(detail);
												} catch (error) {
													console.log(error);
												}
											}}
											className="text-sm"
										>
											{kustomization.name}
										</Button>
									</TableCell>
									<TableCell>
										<span className="font-mono text-sm">
											{kustomization.path}
										</span>
									</TableCell>
									<TableCell>
										<div className="flex items-center space-x-2">
											{kustomization.validated ? (
												<CheckCircle className="h-4 w-4 text-green-500" />
											) : (
												<AlertCircle className="h-4 w-4 text-yellow-500" />
											)}
											<span
												className={`text-sm ${kustomization.validated ? "text-green-600" : "text-yellow-600"}`}
											>
												{kustomization.validated
													? "Validated"
													: "Not Validated"}
											</span>
										</div>
									</TableCell>
									<TableCell>
										<div className="flex items-center space-x-2">
											{/* <Avatar className="h-6 w-6">
												<AvatarFallback className="bg-primary/10 text-xs">
													{kustomization.owner
														.split(" ")
														.map((n) => n[0])
														.join("")}
												</AvatarFallback>
											</Avatar> */}
											<span className="text-sm">{kustomization.owner}</span>
										</div>
									</TableCell>
									<TableCell>
										<span className="capitalize px-2 py-1 bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300 rounded-full text-sm">
											{kustomization.source.type}
										</span>
									</TableCell>
									<TableCell>
										<span
											className={`capitalize px-2 py-1 rounded-full text-sm ${
												kustomization.appType === "kustomization"
													? "bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300"
													: "bg-purple-50 text-purple-700 dark:bg-purple-900/50 dark:text-purple-300"
											}`}
										>
											{kustomization.appType}
										</span>
									</TableCell>
									<TableCell>
										<div className="flex flex-wrap gap-1">
											{kustomization.environments?.map((env) => (
												<span
													key={env}
													className="px-2 py-1 text-xs rounded-full bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300"
												>
													{env}
												</span>
											))}
										</div>
									</TableCell>
									<TableCell>
										<Button variant="ghost" size="sm">
											<ChevronRight className="h-4 w-4" />
										</Button>
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
				</CardContent>
			</Card>
			{renderCreateDialog()}
		</div>
	);
}
