import { create } from 'zustand';
import type { SecretStore, SecretStoreResponse } from '@/types';
import requestor from '@/requestor';
import type { BaseState, BaseActions } from '@/types/store';
import { load as yamlLoad } from 'js-yaml';

interface SecretState extends BaseState<SecretStore> {
  selectedStore: SecretStore | null;
}

interface SecretActions {
  setSelectedStore: (store: SecretStore | null) => void;
  fetch: () => Promise<void>;
  create: (yaml: string) => Promise<void>;
  update: (name: string, yaml: string) => Promise<void>;
  remove: (name: string) => Promise<void>;
  reset: () => void;
}

export const useSecretStore = create<SecretState & SecretActions>((set, get) => ({
  data: [],
  isLoading: false,
  error: null,
  selectedStore: null,

  setSelectedStore: (store) => set({ selectedStore: store }),

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

  create: async (yaml: string) => {
    set({ isLoading: true, error: null });
    try {
      const data = yamlLoad(yaml);
      await requestor.post('/api/v1/security/externalsecrets/secretstore', data);
      await get().fetch();
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to create secret store') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  update: async (name: string, yaml: string) => {
    set({ isLoading: true, error: null });
    try {
      const data = yamlLoad(yaml);
      await requestor.put(`/api/v1/security/externalsecrets/secretstore/${name}`, data);
      await get().fetch();
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to update secret store') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  remove: async (name: string) => {
    set({ isLoading: true, error: null });
    try {
      await requestor.delete(`/api/v1/security/externalsecrets/secretstore/${name}`);
      await get().fetch();
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to delete secret store') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  reset: () => set({
    data: [],
    isLoading: false,
    error: null,
    selectedStore: null
  }),
}));