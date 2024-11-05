"use client"

import { ChevronRight } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

interface BreadcrumbItem {
  label: string;
  href: string;
  onClick?: () => void;
}

interface BreadcrumbProps {
  items: BreadcrumbItem[];
  onNavigate?: (item: BreadcrumbItem) => void;
}

export function Breadcrumb({ items, onNavigate }: BreadcrumbProps) {
  const router = useRouter();

  const handleClick = (item: BreadcrumbItem) => {
    if (item.onClick) {
      item.onClick();
    } else if (onNavigate) {
      onNavigate(item);
    } else {
      router.push(item.href);
    }
  };

  return (
    <nav className="flex items-center space-x-1 text-sm text-muted-foreground">
      {items.map((item, index) => (
        <div key={index} className="flex items-center">
          {index > 0 && <ChevronRight className="h-4 w-4 mx-1" />}
          <button
            onClick={() => handleClick(item)}
            className="hover:text-foreground transition-colors"
          >
            {item.label}
          </button>
        </div>
      ))}
    </nav>
  );
}