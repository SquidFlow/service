"use client";

import { ExternalSecrets } from './ExternalSecrets';

interface SecurityProps {
  activeSubMenu: string;
}

export function Security({ activeSubMenu }: SecurityProps) {
  if (activeSubMenu === 'ExternalSecrets') {
    return <ExternalSecrets />;
  }

  return null;
}