import { create } from 'zustand';
import type { SecretStore } from '@/types/security';
import type { SecretStoreResponse } from '@/types/api';
import requestor from '@/requestor';
import type { BaseState, BaseActions } from '@/types/store';

interface SecretState extends BaseState<SecretStore> {
  selectedStore?: SecretStore;
}

interface SecretActions extends BaseActions<SecretStore> {
  setSelectedStore: (store: SecretStore | undefined) => void;
  updateSecretStore: (id: string, store: Partial<SecretStore>) => Promise<void>;
}

export const useSecretStore = create<SecretState & SecretActions>((set, get) => ({
  data: [],
  isLoading: false,
  error: null,
  selectedStore: undefined,

  fetch: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get<SecretStoreResponse>('/api/v1/security/externalsecrets/secretstore');
      set({ data: response.data.items || [] });
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to fetch secret stores') });
    } finally {
      set({ isLoading: false });
    }
  },

  setSelectedStore: (store) => set({ selectedStore: store }),

  updateSecretStore: async (id, store) => {
    set({ isLoading: true, error: null });
    try {
      await requestor.put(`/api/v1/security/externalsecrets/secretstore/${id}`, store);
      await get().fetch(); // 刷新列表
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to update secret store') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  reset: () => set({ data: [], isLoading: false, error: null, selectedStore: undefined }),
}));