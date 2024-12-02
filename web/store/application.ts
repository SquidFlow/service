import { create } from 'zustand';
import type {
  ApplicationTemplate,
  ApplicationResponse,
  ValidatePayload,
  CreateApplicationPayload,
  DryRunResult,
  ValidateResult
} from '@/types';
import requestor from '@/requestor';
import type { BaseState, BaseActions } from '@/types/store';
import { ARGOCDAPPLICATIONS, RELEASEHISTORIES } from '@/app/api';

interface ApplicationState extends BaseState<ApplicationTemplate> {
  dryRunData: DryRunResult | null;
  deploymentStatus: {
    isDeploying: boolean;
    error: Error | null;
    lastDeployedApp?: string;
  };
  releaseHistories: {
    SIT: any[];
    UAT: any[];
    PRD: any[];
  };
}

interface ApplicationActions extends BaseActions<ApplicationTemplate> {
  validateApplication: (payload: ValidatePayload) => Promise<ValidateResult>;
  createApplication: (payload: CreateApplicationPayload) => Promise<void>;
  dryRun: (payload: CreateApplicationPayload) => Promise<DryRunResult>;
  getApplicationDetail: (name: string) => Promise<ApplicationTemplate>;
  deleteApplication: (name: string) => Promise<void>;
  deleteApplications: (names: string[]) => Promise<void>;
  clearDryRunData: () => void;
  getReleaseHistories: () => Promise<void>;
}

export const useApplicationStore = create<ApplicationState & ApplicationActions>((set, get) => ({
  // 初始状态
  data: [],
  isLoading: false,
  error: null,
  dryRunData: null,
  deploymentStatus: {
    isDeploying: false,
    error: null,
  },
  releaseHistories: { SIT: [], UAT: [], PRD: [] },

  // Actions
  fetch: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get<ApplicationResponse>('/api/v1/deploy/argocdapplications');
      set({ data: response.data.items || [] });
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to fetch applications') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  getApplicationDetail: async (name: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get<ApplicationTemplate>(
        `/api/v1/deploy/argocdapplications/${name}`
      );
      return response.data;
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to get application detail') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  validateApplication: async (payload) => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.post<ValidateResult>(
        `${ARGOCDAPPLICATIONS}/validate`,
        payload
      );
      return response.data;
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to validate') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  dryRun: async (payload) => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.post<DryRunResult>('/api/v1/deploy/argocdapplications/dryrun', payload);
      const result = response.data;
      set({ dryRunData: result });
      return result;
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to dry run') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  createApplication: async (payload) => {
    set(state => ({
      deploymentStatus: { ...state.deploymentStatus, isDeploying: true, error: null }
    }));
    try {
      await requestor.post('/api/v1/deploy/argocdapplications', payload);
      set(state => ({
        deploymentStatus: {
          ...state.deploymentStatus,
          lastDeployedApp: payload.application_name
        }
      }));
      await get().fetch(); // 刷新列表
    } catch (error) {
      const errorMsg = error instanceof Error ? error : new Error('Failed to create application');
      set(state => ({
        deploymentStatus: { ...state.deploymentStatus, error: errorMsg }
      }));
      throw error;
    } finally {
      set(state => ({
        deploymentStatus: { ...state.deploymentStatus, isDeploying: false }
      }));
    }
  },

  deleteApplication: async (name: string) => {
    set({ isLoading: true, error: null });
    try {
      await requestor.delete(`/api/v1/deploy/argocdapplications/${name}`);
      await get().fetch(); // 刷新列表
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to delete application') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  deleteApplications: async (names: string[]) => {
    set({ isLoading: true, error: null });
    try {
      const deletePromises = names.map(name =>
        requestor.delete(`/api/v1/deploy/argocdapplications/${name}`)
      );
      await Promise.all(deletePromises);
      await get().fetch(); // 刷新列表
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to delete applications') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  clearDryRunData: () => set({ dryRunData: null }),

  reset: () => set({
    data: [],
    isLoading: false,
    error: null,
    dryRunData: null,
    deploymentStatus: {
      isDeploying: false,
      error: null,
      lastDeployedApp: undefined
    },
    releaseHistories: { SIT: [], UAT: [], PRD: [] }
  }),

  getReleaseHistories: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get(RELEASEHISTORIES);
      set({ releaseHistories: response.data || { SIT: [], UAT: [], PRD: [] } });
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to fetch release histories') });
    } finally {
      set({ isLoading: false });
    }
  },
}));