import { create } from 'zustand';
import type { TenantInfo, TenantResponse, AppCodeResponse } from '@/types';
import requestor from '@/requestor';
import type { BaseState, BaseActions } from '@/types/store';
import { API_PATHS } from './api';

interface TenantState extends BaseState<TenantInfo> {
  appCodes: string[];
  selectedTenant?: TenantInfo;
}

interface TenantActions extends BaseActions<TenantInfo> {
  fetchAppCodes: () => Promise<void>;
  setSelectedTenant: (tenant: TenantInfo | undefined) => void;
}

export const useTenantStore = create<TenantState & TenantActions>((set) => ({
  data: [],
  appCodes: [],
  isLoading: false,
  error: null,
  selectedTenant: undefined,

  fetch: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get<TenantResponse>(API_PATHS.tenants.list);
      set({ data: response.data.items || [] });
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to fetch tenants') });
    } finally {
      set({ isLoading: false });
    }
  },

  fetchAppCodes: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get<AppCodeResponse>(API_PATHS.appCodes.list);
      const appCodes = Array.isArray(response.data.appCodes) ? response.data.appCodes : [];
      console.log('Raw API response:', response.data);
      console.log('Processed app codes:', appCodes);
      set({ appCodes });
    } catch (error) {
      console.error('Failed to fetch app codes:', error);
      set({ error: error instanceof Error ? error : new Error('Failed to fetch app codes') });
      set({ appCodes: [] });
    } finally {
      set({ isLoading: false });
    }
  },

  setSelectedTenant: (tenant) => set({ selectedTenant: tenant }),

  reset: () => set({
    data: [],
    appCodes: [],
    isLoading: false,
    error: null,
    selectedTenant: undefined
  }),
}));