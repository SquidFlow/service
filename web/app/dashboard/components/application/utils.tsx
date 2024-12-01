import React, { ReactNode } from 'react';
import { AlertCircle, Clock, CheckCircle, XCircle } from "lucide-react";
import { ApplicationTemplate } from '@/types/application';

export const renderResourceValue = (app: ApplicationTemplate, env: string, field: 'cpu' | 'memory' | 'storage' | 'pods') => {
  return app.resources?.[env]?.[field] || '0';
};

export const getStatusIcon = (status: string): ReactNode => {
  switch (status) {
    case "Succeeded":
    case "Synced":
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="lucide lucide-circle-check-big h-4 w-4 text-green-500"
        >
          <path d="M21.801 10A10 10 0 1 1 17 3.335" />
          <path d="m9 11 3 3L22 4" />
        </svg>
      );
    case "OutOfSync":
      return <AlertCircle className="h-4 w-4 text-yellow-500" />;
    case "Progressing":
      return <Clock className="h-4 w-4 text-blue-500" />;
    case "Degraded":
      return <XCircle className="h-4 w-4 text-red-500" />;
    default:
      return <AlertCircle className="h-4 w-4 text-gray-500" />;
  }
};