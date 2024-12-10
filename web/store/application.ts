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
import { API_PATHS } from './api';

interface ApplicationState extends BaseState<ApplicationTemplate> {
  dryRunData: DryRunResult | null;
  deploymentStatus: {
    isDeploying: boolean;
    error: Error | null;
    lastDeployedApp?: string;
  };
}

interface ApplicationActions extends BaseActions<ApplicationTemplate> {
  validateApplication: (payload: ValidatePayload) => Promise<ValidateResult>;
  createApplication: (payload: CreateApplicationPayload) => Promise<void>;
  dryRun: (payload: CreateApplicationPayload) => Promise<DryRunResult>;
  getApplicationDetail: (name: string) => Promise<ApplicationTemplate>;
  deleteApplications: (names: string[]) => Promise<void>;
}

export const useApplicationStore = create<ApplicationState & ApplicationActions>((set, get) => ({
  data: [],
  isLoading: false,
  error: null,
  dryRunData: null,
  deploymentStatus: {
    isDeploying: false,
    error: null,
  },

  // Actions
  fetch: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get<ApplicationResponse>(API_PATHS.applications.list);
      set({ data: response.data.items || [] });
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to fetch applications') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  validateApplication: async (payload) => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.post<ValidateResult>(
        API_PATHS.applications.validate,
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

  createApplication: async (payload) => {
    set(state => ({
      deploymentStatus: { ...state.deploymentStatus, isDeploying: true, error: null }
    }));
    try {
      const formattedPayload = {
        application_source: {
          repo: payload.application_source.repo,
          target_revision: payload.application_source.target_revision,
          path: payload.application_source.path,
          submodules: payload.application_source.submodules,
          application_specifier: payload.application_source.application_specifier
        },
        application_instantiation: {
          application_name: payload.application_instantiation.application_name,
          tenant_name: payload.application_instantiation.tenant_name,
          appcode: payload.application_instantiation.appcode,
          description: payload.application_instantiation.description
        },
        application_target: payload.application_target.map(target => ({
          cluster: target.cluster,
          namespace: target.namespace
        })),
        is_dryrun: payload.is_dryrun
      };

      await requestor.post(API_PATHS.applications.create, formattedPayload);
      set(state => ({
        deploymentStatus: {
          ...state.deploymentStatus,
          lastDeployedApp: payload.application_instantiation.application_name
        }
      }));
      await get().fetch(); // refresh app list
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

  dryRun: async (payload) => {
    set({ isLoading: true, error: null });
    try {
      const formattedPayload = {
        application_source: {
          repo: payload.application_source.repo,
          target_revision: payload.application_source.target_revision,
          path: payload.application_source.path,
          submodules: payload.application_source.submodules,
        },
        application_instantiation: {
          application_name: payload.application_instantiation.application_name,
          tenant_name: payload.application_instantiation.tenant_name,
          appcode: payload.application_instantiation.appcode,
          description: payload.application_instantiation.description
        },
        application_target: payload.application_target.map(target => ({
          cluster: target.cluster,
          namespace: target.namespace
        })),
        is_dryrun: true
      };
      if (payload.application_source.application_specifier) {
        (formattedPayload.application_source as any).application_specifier = payload.application_source.application_specifier;
      }

      const response = await requestor.post<DryRunResult>(
        API_PATHS.applications.create,
        formattedPayload
      );
      return response.data;
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to dry run') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  getApplicationDetail: async (name: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await requestor.get<ApplicationTemplate>(
        API_PATHS.applications.get(name)
      );
      return response.data;
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to get application detail') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  deleteApplications: async (names: string[]) => {
    set({ isLoading: true, error: null });
    try {
      await Promise.all(
        names.map(name => requestor.delete(API_PATHS.applications.delete(name)))
      );
      await get().fetch(); // 重新获取应用列表
    } catch (error) {
      set({ error: error instanceof Error ? error : new Error('Failed to delete applications') });
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  reset: () => set({
    data: [],
    isLoading: false,
    error: null,
    dryRunData: null,
    deploymentStatus: {
      isDeploying: false,
      error: null,
      lastDeployedApp: undefined
    }
  })
}));