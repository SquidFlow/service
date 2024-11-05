export interface User {
  username: string;
  password: string;
  role: string;
}

export const usersData: User[] = [
  { username: 'user1', password: 'user1', role: 'user' },
  { username: 'admin', password: 'admin', role: 'admin' },
];
