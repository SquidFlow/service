"use client"

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation';
import { Button } from "@/components/ui/button"
import { ChevronUp, ChevronRight, Lock, Settings, ChevronLeft, Workflow } from "lucide-react"
import { MenuItem } from './interfaces';
import { ScrollArea } from "@/components/ui/scroll-area"
import { Breadcrumb } from "@/app/components/breadcrumb";
import Header from '@/app/components/header';
import type { BreadcrumbItem } from '@/app/components/breadcrumb';

const menuItems: MenuItem[] = [
  {
    title: 'Deploy',
    icon: Workflow,
    subItems: ['Application', 'DestinationCluster']
  },
  {
    title: 'Security',
    icon: Lock,
    subItems: ['ExternalSecrets']
  }
]

const settingsItem: MenuItem = {
  title: 'Settings',
  icon: Settings,
  subItems: []
};

export default function DashboardLayout({
  children
}: {
  children: React.ReactNode
}) {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const router = useRouter();
  const [activeMenu, setActiveMenu] = useState('Deploy')
  const [activeSubMenu, setActiveSubMenu] = useState('Application')
  const [expandedMenus, setExpandedMenus] = useState<string[]>(['Deploy']);
  const [username, setUsername] = useState('');
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [selectedAppName, setSelectedAppName] = useState<string | null>(null);

  const toggleMenu = (title: string) => {
    setExpandedMenus(prev =>
      prev.includes(title)
        ? prev.filter(item => item !== title)
        : [...prev, title]
    );
  };

  const handleLogout = () => {
    localStorage.removeItem('isLoggedIn');
    localStorage.removeItem('userRole');
    router.push('/login');
  };

  useEffect(() => {
    const loggedIn = localStorage.getItem('isLoggedIn') === 'true';
    const userRole = localStorage.getItem('userRole');
    const username = localStorage.getItem('username');
    if (!loggedIn) {
      router.push('/login');
    } else if (userRole === 'admin') {
      router.push('/admin');
    } else {
      setIsLoggedIn(true);
      setUsername(username || '');
    }
  }, [router]);

  if (!isLoggedIn) {
    return null;
  }

  const getBreadcrumbItems = (): BreadcrumbItem[] => {
    const baseBreadcrumbItems: BreadcrumbItem[] = [
      {
        label: 'Dashboard',
        href: '/dashboard',
        onClick: () => {
          setActiveMenu('Deploy');
          setActiveSubMenu('Application');
          setSelectedAppName(null);
        }
      },
    ];

    switch (activeMenu) {
      case 'Deploy':
        const deployItems: BreadcrumbItem[] = [
          {
            label: 'Deploy',
            href: '/dashboard',
            onClick: () => {
              setActiveMenu('Deploy');
              setActiveSubMenu('Application');
              setSelectedAppName(null);
            }
          },
          {
            label: activeSubMenu,
            href: '/dashboard',
            onClick: () => {
              setActiveSubMenu(activeSubMenu);
              setSelectedAppName(null);
            }
          }
        ];

        if (selectedAppName) {
          deployItems.push({
            label: selectedAppName,
            href: '/dashboard',
            onClick: () => {}
          });
        }

        return [...baseBreadcrumbItems, ...deployItems];

      case 'Security':
        return [
          ...baseBreadcrumbItems,
          {
            label: 'Security',
            href: '/dashboard',
            onClick: () => {
              setActiveMenu('Security');
            }
          }
        ];

      default:
        return baseBreadcrumbItems;
    }
  };

  return (
    <div className="flex flex-col h-screen bg-background text-foreground">
      <Header
        isLoggedIn={isLoggedIn}
        username={username}
        userRole={localStorage.getItem('userRole') ?? undefined}
        onLogout={handleLogout}
      />
      <div className="px-6 py-2 border-b border-border">
        <Breadcrumb
          items={getBreadcrumbItems()}
          onNavigate={(item) => {
            if (item.onClick) {
              item.onClick();
            }
          }}
        />
      </div>
      <div className="flex flex-1 overflow-hidden">
        <aside className={`${
          isSidebarCollapsed ? 'w-16' : 'w-72'
        } border-r border-border bg-card flex flex-col transition-all duration-300 ease-in-out relative`}>
          <Button
            variant="ghost"
            size="icon"
            className="absolute -right-3 top-3 h-6 w-6 rounded-full border border-border bg-background shadow-sm z-10"
            onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
          >
            {isSidebarCollapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4" />
            )}
          </Button>

          <ScrollArea className="flex-1">
            <div className="p-6">
              <nav className="space-y-4">
                {menuItems.map((item) => (
                  <div key={item.title} className="space-y-1">
                    <Button
                      variant={activeMenu === item.title ? "secondary" : "ghost"}
                      className={`w-full justify-between px-4 py-3 hover:bg-accent hover:text-accent-foreground
                        ${activeMenu === item.title ? 'bg-secondary/50 shadow-sm font-medium' : 'text-muted-foreground'}
                        rounded-lg transition-all duration-200 ease-in-out group ${
                          isSidebarCollapsed ? 'justify-center' : ''
                        }`}
                      onClick={() => {
                        setActiveMenu(item.title)
                        if (item.subItems.length > 0 && !isSidebarCollapsed) {
                          toggleMenu(item.title)
                          const subItem = item.subItems[0];
                          setActiveSubMenu(subItem);
                          if (item.title === 'Deploy' && subItem === 'Application') {
                            router.push('/dashboard/deploy/application');
                          }
                        }
                      }}
                    >
                      <span className={`flex items-center text-sm ${
                        isSidebarCollapsed ? 'justify-center' : ''
                      }`}>
                        <item.icon className={`h-4 w-4 transition-colors
                          ${activeMenu === item.title ? 'text-foreground' : 'text-muted-foreground group-hover:text-foreground'}`}
                        />
                        {!isSidebarCollapsed && <span className="ml-3">{item.title}</span>}
                      </span>
                      {!isSidebarCollapsed && item.subItems.length > 0 && (
                        <span className={`transition-transform duration-200
                          ${expandedMenus.includes(item.title) ? 'rotate-180' : ''}
                          ${activeMenu === item.title ? 'text-foreground' : 'text-muted-foreground'}`}>
                          <ChevronUp className="h-4 w-4" />
                        </span>
                      )}
                    </Button>
                    {!isSidebarCollapsed && expandedMenus.includes(item.title) && (
                      <div className="ml-4 pl-4 border-l border-border/50 space-y-1">
                        {item.subItems.map((subItem) => (
                          <Button
                            key={subItem}
                            variant={activeSubMenu === subItem ? "secondary" : "ghost"}
                            className={`w-full justify-start px-4 py-2 text-sm rounded-lg
                              ${activeSubMenu === subItem
                                ? 'bg-secondary/50 text-foreground font-medium'
                                : 'text-muted-foreground hover:text-foreground'
                              }
                              transition-all duration-200 group`}
                            onClick={() => {
                              setActiveMenu(item.title);
                              setActiveSubMenu(subItem);
                              if (item.title === 'Deploy') {
                                if (subItem === 'Application') {
                                  router.push('/dashboard/deploy/application');
                                } else if (subItem === 'DestinationCluster') {
                                  router.push('/dashboard/deploy/destinationcluster');
                                }
                              } else if (item.title === 'Security') {
                                if (subItem === 'ExternalSecrets') {
                                  router.push('/dashboard/security/externalsecrets');
                                }
                              }
                            }}
                          >
                            <span className="relative flex items-center">
                              <span className={`absolute -left-6 top-1/2 -translate-y-1/2 w-2 h-2 rounded-full bg-border
                                group-hover:bg-foreground/50 transition-colors
                                ${activeSubMenu === subItem ? 'bg-foreground' : ''}`}
                              />
                              {subItem}
                            </span>
                          </Button>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </nav>
            </div>
          </ScrollArea>

          <div className="p-6 border-t border-border mt-auto">
            <Button
              variant={activeMenu === settingsItem.title ? "secondary" : "ghost"}
              className={`w-full justify-start px-4 py-3 hover:bg-accent hover:text-accent-foreground
                ${activeMenu === settingsItem.title ? 'bg-secondary/50 shadow-sm font-medium' : 'text-muted-foreground'}
                rounded-lg transition-all duration-200 ease-in-out group ${
                  isSidebarCollapsed ? 'justify-center' : ''
                }`}
              onClick={() => {
                setActiveMenu(settingsItem.title);
                router.push('/dashboard/setting');
              }}
            >
              <span className={`flex items-center text-sm ${
                isSidebarCollapsed ? 'justify-center' : ''
              }`}>
                <settingsItem.icon className={`h-4 w-4 transition-colors
                  ${activeMenu === settingsItem.title ? 'text-foreground' : 'text-muted-foreground group-hover:text-foreground'}`}
                />
                {!isSidebarCollapsed && <span className="ml-3">{settingsItem.title}</span>}
              </span>
            </Button>
          </div>
        </aside>
        <main className="flex-1 overflow-auto">
          {children}
        </main>
      </div>
    </div>
  );
}