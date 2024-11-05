"use client"

import { faker } from '@faker-js/faker';
import { ResourceQuotaInterface, Node } from '../interfaces';

// Mock data
export const resourceQuota: ResourceQuotaInterface = {
  cpu: '10 cores',
  memory: '32Gi',
  storage: '500Gi'
};

// Generate mock nodes data
export const nodes: Node[] = Array.from({ length: faker.number.int({ min: 3, max: 6 }) }, (_, index) => ({
  id: index + 1,
  name: `node-${faker.number.int({ min: 1, max: 10 })}`,
  ip: faker.internet.ip(),
  status: faker.helpers.arrayElement(['Ready', 'NotReady'] as const),
  cpu: `${faker.number.int({ min: 2, max: 8 })} cores`,
  memory: `${faker.number.int({ min: 8, max: 32 })}Gi`
}));

// Generate mock usage trends data
const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'];
export const usageTrends = months.map((name, index) => ({
  name,
  cpu: 4 + index + faker.number.float({ min: 0, max: 1 }),
  memory: 16 + index * 4 + faker.number.int({ min: 0, max: 4 }),
  storage: 100 + index * 50 + faker.number.int({ min: 0, max: 50 })
}));

// Generate mock aggregated resources data
export const aggregatedResources = {
  tenants: Array.from({ length: 3 }, () => ({
    name: `Tenant ${faker.helpers.arrayElement(['A', 'B', 'C', 'D', 'E'])}`,
    cpu: `${faker.number.int({ min: 2, max: 5 })} cores`,
    memory: `${faker.helpers.arrayElement([8, 16, 32])}Gi`,
    storage: `${faker.number.int({ min: 100, max: 300 })}Gi`
  })),

  appcodes: Array.from({ length: faker.number.int({ min: 3, max: 5 })}, () => ({
    name: `App${faker.number.int({ min: 1, max: 5 })}`,
    cpu: `${faker.number.int({ min: 2, max: 4 })} cores`,
    memory: `${faker.helpers.arrayElement([4, 8, 16])}Gi`,
    storage: `${faker.number.int({ min: 50, max: 200 })}Gi`
  })),

  applications: Array.from({ length: faker.number.int({ min: 4, max: 6 })}, () => ({
    name: `Application${faker.number.int({ min: 1, max: 6 })}`,
    cpu: `${faker.number.int({ min: 1, max: 3 })} cores`,
    memory: `${faker.helpers.arrayElement([4, 8, 12, 16])}Gi`,
    storage: `${faker.number.int({ min: 50, max: 250 })}Gi`
  }))
};
