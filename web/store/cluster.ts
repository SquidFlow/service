import { create } from 'zustand';
import requestor from '@/requestor';
import { API_PATHS } from './api';
import { ClusterInfo } from '@/types/cluster';

interface ClusterStore {
  data: ClusterInfo[];
  isLoading: boolean;
  error: Error | null;
  searchTerm: string;
  selectedEnvironment: string;
  selectedProvider: string;
  selectedCluster: ClusterInfo | null;

  setSearchTerm: (term: string) => void;
  setSelectedEnvironment: (env: string) => void;
  setSelectedProvider: (provider: string) => void;
  setSelectedCluster: (cluster: ClusterInfo | null) => void;

  getClusterList: () => Promise<void>;
  getFilteredClusters: () => ClusterInfo[];
  updateClusterQuota: (clusterName: string, quota: any) => Promise<void>;
  reset: () => void;
}

export const useClusterStore = create<ClusterStore>((set, get) => ({
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

  getClusterList: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get(API_PATHS.clusters.list);
      set({ data: response.data.items || [] });
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to fetch clusters') });
      console.error('Failed to fetch clusters:', error);
    } finally {
      set({ isLoading: false });
    }
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
      await requestor.put(API_PATHS.clusters.quota(clusterName), quota);
      await get().getClusterList();
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