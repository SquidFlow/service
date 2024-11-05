"use client"

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AlertCircle } from 'lucide-react';
import { usersData as users } from './loginMock';

export function FakeAuthentication() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const userExists = users.find(u => u.username === username);
    if (!userExists) {
      setError(`User "${username}" does not exist. Please check your username.`);
      return;
    }

    const user = users.find(u => u.username === username && u.password === password);
    if (!user) {
      setError('Incorrect password. Please try again.');
      return;
    }

    localStorage.setItem('isLoggedIn', 'true');
    localStorage.setItem('userRole', user.role);
    localStorage.setItem('username', user.username);

    if (user.role === 'admin') {
      router.push('/admin');
    } else {
      router.push('/dashboard');
    }
  };

  return (
    <form onSubmit={handleLogin} className="space-y-4">
      <div className="space-y-2">
        <label htmlFor="username" className="text-sm font-medium text-foreground">Username</label>
        <Input
          id="username"
          type="text"
          value={username}
          onChange={(e) => {
            setUsername(e.target.value);
            setError(null);
          }}
          required
          className={`bg-background text-foreground border-border focus:border-primary focus:ring-primary
            ${error && error.includes('username') ? 'border-red-500' : ''}`}
        />
      </div>
      <div className="space-y-2">
        <label htmlFor="password" className="text-sm font-medium text-foreground">Password</label>
        <Input
          id="password"
          type="password"
          value={password}
          onChange={(e) => {
            setPassword(e.target.value);
            setError(null);
          }}
          required
          className={`bg-background text-foreground border-border focus:border-primary focus:ring-primary
            ${error && error.includes('password') ? 'border-red-500' : ''}`}
        />
      </div>
      {error && (
        <div className="flex items-center space-x-2 text-red-500 text-sm bg-red-50 dark:bg-red-900/20 p-3 rounded-md">
          <AlertCircle className="h-4 w-4" />
          <p>{error}</p>
        </div>
      )}
      <Button
        type="submit"
        className="w-full bg-primary text-primary-foreground hover:bg-primary/90 transition duration-200 ease-in-out"
      >
        Login
      </Button>
    </form>
  );
}