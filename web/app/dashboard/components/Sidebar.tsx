"use client";

import { usePathname, useRouter } from 'next/navigation';
import { Rocket, Shield, Settings } from "lucide-react";
import { Button } from "@/components/ui/button";

interface MenuItem {
  label: string;
  icon: React.ReactNode;
  href?: string;
  subItems?: {
    label: string;
    href: string;
  }[];
}

const menuItems: MenuItem[] = [
  {
    label: 'Deploy',
    icon: <Rocket className="h-4 w-4" />,
    subItems: [
      {
        label: 'Applications',
        href: '/dashboard/deploy/application'
      },
      {
        label: 'Clusters',
        href: '/dashboard/deploy/cluster'
      }
    ]
  },
  {
    label: 'Security',
    icon: <Shield className="h-4 w-4" />,
    subItems: [
      {
        label: 'External Secrets',
        href: '/dashboard/security/externalsecrets'
      }
    ]
  },
  {
    label: 'Settings',
    icon: <Settings className="h-4 w-4" />,
    href: '/dashboard/setting'
  }
];

export function Sidebar() {
  const router = useRouter();
  const pathname = usePathname();

  const isActive = (href: string) => pathname === href;
  const isActiveParent = (item: MenuItem) => {
    if (item.href) {
      return isActive(item.href);
    }
    return item.subItems?.some(subItem => pathname.startsWith(subItem.href)) ?? false;
  };

  const handleClick = (item: MenuItem) => {
    if (item.href) {
      router.push(item.href);
    } else if (item.subItems?.[0]) {
      router.push(item.subItems[0].href);
    }
  };

  return (
    <nav className="space-y-2">
      {menuItems.map((item) => (
        <div key={item.label} className="space-y-1">
          <Button
            variant={isActiveParent(item) ? "secondary" : "ghost"}
            className="w-full justify-start"
            onClick={() => handleClick(item)}
          >
            {item.icon}
            <span className="ml-2">{item.label}</span>
          </Button>
          {item.subItems && (
            <div className="ml-4 space-y-1">
              {item.subItems.map((subItem) => (
                <Button
                  key={subItem.href}
                  variant={isActive(subItem.href) ? "secondary" : "ghost"}
                  className="w-full justify-start pl-6"
                  onClick={() => router.push(subItem.href)}
                >
                  {subItem.label}
                </Button>
              ))}
            </div>
          )}
        </div>
      ))}
    </nav>
  );
}