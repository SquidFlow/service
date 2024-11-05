"use client"

import React from 'react';
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/theme-toggle";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Github } from 'lucide-react';
import Link from 'next/link';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { User, Clock, Building, Shield } from 'lucide-react';

interface HeaderProps {
  isLoggedIn?: boolean;
  username?: string;
  userRole?: string;
  onLogout?: () => void;
}

export default function Header({ isLoggedIn, username, userRole, onLogout }: HeaderProps) {
  const githubRepoUrl = "https://github.com/h4-poc/dashboard";

  return (
    <header className="bg-background border-b border-border">
      <div className="w-full px-6 py-4 flex justify-between items-center">
        <div className="flex items-center">
          <h1 className="text-3xl font-bold text-foreground mr-6">
            <span className="bg-primary text-primary-foreground px-3 py-1.5 rounded-md mr-2 transition-all duration-300 hover:bg-secondary hover:text-secondary-foreground">
              H4
            </span>
            Platform
          </h1>
        </div>
        <nav className="flex items-center space-x-6">
          {!isLoggedIn && (
            <>
              <Button variant="ghost" size="sm" className="text-base">
                Docs
              </Button>
              <a
                href={githubRepoUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center text-base"
              >
                <Github className="mr-3 h-5 w-5" />
                GitHub
              </a>
              <Link href="/login" passHref>
                <Button variant="default" size="sm" className="text-base px-6">
                  Login
                </Button>
              </Link>
            </>
          )}
          {isLoggedIn && (
            <>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <div className="flex items-center space-x-3 cursor-pointer">
                    <span className="text-base text-foreground">Welcome, {username}</span>
                    <Avatar className="h-10 w-10 hover:ring-2 hover:ring-primary transition-all">
                      <AvatarFallback className="bg-primary/10 text-lg">
                        {username ? username[0].toUpperCase() : 'U'}
                      </AvatarFallback>
                    </Avatar>
                  </div>
                </DropdownMenuTrigger>
                <DropdownMenuContent className="w-64" align="end" forceMount>
                  <DropdownMenuLabel className="font-normal p-4">
                    <div className="flex flex-col space-y-1.5">
                      <p className="text-base font-medium leading-none">{username}</p>
                      <p className="text-sm leading-none text-muted-foreground">
                        user@example.com
                      </p>
                    </div>
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuGroup>
                    <DropdownMenuItem className="p-3">
                      <Shield className="mr-3 h-5 w-5" />
                      <span className="text-sm">Role: {userRole}</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem className="p-3">
                      <Clock className="mr-3 h-5 w-5" />
                      <span className="text-sm">Timezone: UTC+8</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem className="p-3">
                      <Building className="mr-3 h-5 w-5" />
                      <span className="text-sm">Tenant: Default</span>
                    </DropdownMenuItem>
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    className="text-red-600 cursor-pointer p-3"
                    onClick={onLogout}
                  >
                    Log out
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
              <ThemeToggle />
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
