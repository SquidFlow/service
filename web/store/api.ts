const API_BASE = '/api/v1'

export const API_PATHS = {
  // Applications
  applications: {
    list: `${API_BASE}/deploy/applications`,
    create: `${API_BASE}/deploy/applications`,
    get: (name: string) => `${API_BASE}/deploy/applications/${name}`,
    delete: (name: string) => `${API_BASE}/deploy/applications/${name}`,
    update: (name: string) => `${API_BASE}/deploy/applications/${name}`,
    validate: `${API_BASE}/deploy/applications/validate`,
    sync: `${API_BASE}/deploy/applications/sync`,
    releasehistories: `${API_BASE}/releasehistories`
  },

  // Clusters
  clusters: {
    list: `${API_BASE}/clusters`,
    create: `${API_BASE}/clusters`,
    get: (name: string) => `${API_BASE}/clusters/${name}`,
    delete: (name: string) => `${API_BASE}/clusters/${name}`,
    update: (name: string) => `${API_BASE}/clusters/${name}`,
    quota: (name: string) => `${API_BASE}/clusters/${name}/quota`
  },

  // Tenants
  tenants: {
    list: `${API_BASE}/tenants`,
    create: `${API_BASE}/tenants`,
    get: (name: string) => `${API_BASE}/tenants/${name}`,
    delete: (name: string) => `${API_BASE}/tenants/${name}`
  },

  // Secret Stores
  secretStores: {
    list: `${API_BASE}/security/externalsecrets/secretstore`,
    create: `${API_BASE}/security/externalsecrets/secretstore`,
    get: (id: string) => `${API_BASE}/security/externalsecrets/secretstore/${id}`,
    delete: (id: string) => `${API_BASE}/security/externalsecrets/secretstore/${id}`,
    update: (id: string) => `${API_BASE}/security/externalsecrets/secretstore/${id}`
  },

  // App Codes
  appCodes: {
    list: `${API_BASE}/appcode`
  },

  // Health Check
  health: `${API_BASE}/healthz`
}

