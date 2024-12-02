import { AxiosError } from 'axios';

export interface BaseState<T> {
  data: T[];
  isLoading: boolean;
  error: Error | AxiosError | null;
}

export interface BaseActions<T> {
  fetch: () => Promise<void>;
  create?: (item: Partial<T>) => Promise<void>;
  update?: (id: string | number, item: Partial<T>) => Promise<void>;
  remove?: (id: string | number) => Promise<void>;
  reset: () => void;
}