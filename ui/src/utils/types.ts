export interface Log {
  id: string;
  level: string;
  message: string;
  timestamp: string;
}

export type ApiResponse<T> = {
  data: T[];
  metadata: {
    limit: number;
    offset: number;
    total: number;
  };
};
