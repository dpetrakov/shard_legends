// API client utilities
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'https://dev.slcw.dimlight.online';

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

export const api = {
  async get(endpoint: string) {
    const response = await fetch(`${API_URL}${endpoint}`);
    
    if (!response.ok) {
      const errorText = await response.text();
      throw new ApiError(
        response.status,
        errorText || `HTTP error! status: ${response.status}`
      );
    }
    
    return response;
  },

  async getJson<T>(endpoint: string): Promise<T> {
    const response = await this.get(endpoint);
    return response.json();
  },

  async post(endpoint: string, data?: any) {
    const response = await fetch(`${API_URL}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: data ? JSON.stringify(data) : undefined,
    });
    
    if (!response.ok) {
      const errorText = await response.text();
      throw new ApiError(
        response.status,
        errorText || `HTTP error! status: ${response.status}`
      );
    }
    
    return response;
  },

  async postJson<T>(endpoint: string, data?: any): Promise<T> {
    const response = await this.post(endpoint, data);
    return response.json();
  }
};

// Пример использования:
// const data = await api.getJson<{message: string}>('/api/ping');
// console.log(data.message); // "pong"