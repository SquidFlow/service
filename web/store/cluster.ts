import { create } from 'zustand';
import type { ClusterInfo } from '@/types';
import type { BaseState, BaseActions } from '@/types/store';
import requestor from '@/requestor';
import { CLUSTER } from '@/app/api';

interface ClusterState extends BaseState<ClusterInfo> {
  searchTerm: string;
  selectedEnvironment: string;
  selectedProvider: string;
  selectedCluster: ClusterInfo | null;
}

interface ClusterActions extends BaseActions<ClusterInfo> {
  setSearchTerm: (term: string) => void;
  setSelectedEnvironment: (env: string) => void;
  setSelectedProvider: (provider: string) => void;
  setSelectedCluster: (cluster: ClusterInfo | null) => void;
  getClusterList: () => Promise<ClusterInfo[]>;
  getFilteredClusters: () => ClusterInfo[];
  updateClusterQuota: (clusterName: string, quota: any) => Promise<void>;
}

export const useClusterStore = create<ClusterState & ClusterActions>((set, get) => ({
  data: [],
  isLoading: false,
  error: null,
  searchTerm: '',
  selectedEnvironment: 'All Environments',
  selectedProvider: 'All Providers',
  selectedCluster: null,

  setSearchTerm: (term) => set({ searchTerm: term }),
  setSelectedEnvironment: (env) => set({ selectedEnvironment: env }),
  setSelectedProvider: (provider) => set({ selectedProvider: provider }),
  setSelectedCluster: (cluster) => set({ selectedCluster: cluster }),

  fetch: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get(CLUSTER);
      set({ data: response.data.items || [] });
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to fetch clusters') });
    } finally {
      set({ isLoading: false });
    }
  },

  getClusterList: async () => {
    const response = await requestor.get(CLUSTER);
    return response.data.items || [];
  },

  getFilteredClusters: () => {
    const { data, searchTerm, selectedEnvironment, selectedProvider } = get();

    return data.filter((cluster) => {
      const name = (cluster.name || '').toLowerCase();
      const environment = (cluster.environment || '').toLowerCase();
      const searchLower = searchTerm.toLowerCase();

      const matchesSearch = name.includes(searchLower) || environment.includes(searchLower);
      const matchesEnvironment = selectedEnvironment === 'All Environments' || cluster.environment === selectedEnvironment;
      const matchesProvider = selectedProvider === 'All Providers' || cluster.provider === selectedProvider;

      return matchesSearch && matchesEnvironment && matchesProvider;
    });
  },

  updateClusterQuota: async (clusterName, quota) => {
    set({ isLoading: true, error: null });
    try {
      await requestor.put(`${CLUSTER}/${clusterName}/quota`, quota);
      await get().fetch();
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
    searchTerm: '',
    selectedEnvironment: 'All Environments',
    selectedProvider: 'All Providers',
    selectedCluster: null
  }),
}));