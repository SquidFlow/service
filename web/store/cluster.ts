import { create } from 'zustand';
import type { ClusterInfo, ClusterResponse, ResourceQuota } from '@/types';
import requestor from '@/requestor';
import type { BaseState, BaseActions } from '@/types/store';
import { CLUSTER } from '@/app/api';

interface ClusterState extends BaseState<ClusterInfo> {
  searchTerm: string;
  selectedEnvironment: string;
  selectedProvider: string;
  selectedCluster?: ClusterInfo;
}

interface ClusterActions extends BaseActions<ClusterInfo> {
  setSearchTerm: (term: string) => void;
  setSelectedEnvironment: (env: string) => void;
  setSelectedProvider: (provider: string) => void;
  setSelectedCluster: (cluster: ClusterInfo | undefined) => void;
  getFilteredClusters: () => ClusterInfo[];
  updateClusterQuota: (name: string, quota: ResourceQuota) => Promise<void>;
  getClusterList: () => Promise<void>;
}

export const useClusterStore = create<ClusterState & ClusterActions>((set, get) => ({
  data: [],
  isLoading: false,
  error: null,
  searchTerm: "",
  selectedEnvironment: "All Environments",
  selectedProvider: "All Providers",
  selectedCluster: undefined,

  fetch: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get<ClusterResponse>('/api/v1/destinationCluster');
      set({ data: response.data.items || [] });
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to fetch clusters') });
    } finally {
      set({ isLoading: false });
    }
  },

  setSearchTerm: (term) => set({ searchTerm: term }),
  setSelectedEnvironment: (env) => set({ selectedEnvironment: env }),
  setSelectedProvider: (provider) => set({ selectedProvider: provider }),
  setSelectedCluster: (cluster) => set({ selectedCluster: cluster }),

  getFilteredClusters: () => {
    const { data, searchTerm, selectedEnvironment, selectedProvider } = get();
    return data.filter((cluster) => {
      const name = (cluster.name || '').toLowerCase();
      const env = (cluster.env || '').toLowerCase();
      const searchLower = searchTerm.toLowerCase();

      const matchesSearch = name.includes(searchLower) || env.includes(searchLower);
      const matchesEnv = selectedEnvironment === "All Environments" || cluster.env === selectedEnvironment;
      const matchesProvider = selectedProvider === "All Providers" || cluster.provider === selectedProvider;

      return matchesSearch && matchesEnv && matchesProvider;
    });
  },

  updateClusterQuota: async (name, quota) => {
    set({ isLoading: true, error: null });
    try {
      await requestor.put(`/api/v1/destinationCluster/${name}/quota`, quota);
      await get().fetch(); // 刷新列表
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to update cluster quota') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  reset: () => set({
    data: [],
    isLoading: false,
    error: null,
    searchTerm: "",
    selectedEnvironment: "All Environments",
    selectedProvider: "All Providers",
    selectedCluster: undefined
  }),

  getClusterList: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get<ClusterResponse>(CLUSTER);
      set({ data: response.data.items || [] });
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to fetch clusters') });
    } finally {
      set({ isLoading: false });
    }
  },
}));