import type { components } from '../api/schema';

const API_BASE_URL = 'http://localhost:8080/api';

export type Folder = components['schemas']['Folder'];
export type FolderListResponse = components['schemas']['FolderListResponse'];
export type ErrorResponse = components['schemas']['ErrorResponse'];
export type TimeParseRequest = components['schemas']['TimeParseRequest'];
export type TimeParseResponse = components['schemas']['TimeParseResponse'];
export type TimeFormat = components['schemas']['TimeFormat'];
export type SupportedFormatsResponse = components['schemas']['SupportedFormatsResponse'];

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json() as ErrorResponse;
      throw new Error(error.message || 'API request failed');
    }

    return response.json() as Promise<T>;
  }

  async getFolders(path?: string): Promise<FolderListResponse> {
    const params = new URLSearchParams();
    if (path) {
      params.append('path', path);
    }
    
    return this.request<FolderListResponse>(
      `/folders${params.toString() ? `?${params.toString()}` : ''}`
    );
  }

  async parseTime(timeString: string): Promise<TimeParseResponse> {
    return this.request<TimeParseResponse>('/time/parse', {
      method: 'POST',
      body: JSON.stringify({ time_string: timeString }),
    });
  }

  async getSupportedFormats(): Promise<SupportedFormatsResponse> {
    return this.request<SupportedFormatsResponse>('/time/formats');
  }

  async saveKoujiEntries(path?: string, outputPath?: string): Promise<{ message: string; output_path: string; count: number }> {
    const params = new URLSearchParams();
    if (path) {
      params.append('path', path);
    }
    if (outputPath) {
      params.append('output_path', outputPath);
    }
    
    return this.request<{ message: string; output_path: string; count: number }>(
      `/kouji-entries/save${params.toString() ? `?${params.toString()}` : ''}`,
      { method: 'POST' }
    );
  }
}

export const apiClient = new ApiClient();

export const folderService = {
  getFolders: (path?: string) => apiClient.getFolders(path),
  saveKoujiEntries: (path?: string, outputPath?: string) => apiClient.saveKoujiEntries(path, outputPath),
};