import { useState, useEffect } from 'react';
import { TenantInfo as TenantInfoType } from './types';
import { TenantInfo } from './TenantInfo';
import { TimezoneSetting } from './TimezoneSetting';

export function RenderSettings() {
  const [tenantInfo, setTenantInfo] = useState<TenantInfoType>({
    name: '',
    id: '',
    type: '',
    createdAt: ''
  });

  useEffect(() => {
    const username = localStorage.getItem('username') || '';
    const userRole = localStorage.getItem('userRole') || '';
    const mockTenantId = 'tenant-' + username.toLowerCase();

    setTenantInfo({
      name: username,
      id: mockTenantId,
      type: userRole,
      createdAt: '2024-01-01'
    });
  }, []);

  const handleTimezoneChange = (timezone: string) => {
    localStorage.setItem('timezone', timezone);
  };

  return (
    <div className="space-y-6">
      <TenantInfo tenantInfo={tenantInfo} />
      <TimezoneSetting
        defaultTimezone={localStorage.getItem('timezone') || 'Asia/Shanghai'}
        onTimezoneChange={handleTimezoneChange}
      />
    </div>
  );
}