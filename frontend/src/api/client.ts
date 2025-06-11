import { createClient } from '@hey-api/client-fetch';
import type { components } from './schema';

// Create the API client
export const client = createClient({
  baseUrl: 'http://localhost:8080/api',
});

// Type exports for easy use
export type Folder = components['schemas']['Folder'];
export type FolderListResponse = components['schemas']['FolderListResponse'];
export type ErrorResponse = components['schemas']['ErrorResponse'];
export type TimeParseRequest = components['schemas']['TimeParseRequest'];
export type TimeParseResponse = components['schemas']['TimeParseResponse'];
export type TimeFormat = components['schemas']['TimeFormat'];
export type SupportedFormatsResponse = components['schemas']['SupportedFormatsResponse'];
export type KoujiFolder = components['schemas']['KoujiFolder'];
export type KoujiFolderListResponse = components['schemas']['KoujiFolderListResponse'];

// API methods
export const folderApi = {
  /**
   * Get folders from the specified path
   */
  async getFolders(path?: string): Promise<FolderListResponse> {
    const response = await client.get({
      url: '/folders',
      query: path ? { path } : undefined,
    });

    if (response.error) {
      throw new Error((response.error as any).message || 'Failed to fetch folders');
    }

    return response.data as unknown as FolderListResponse;
  },
};

// Time API methods
export const timeApi = {
  /**
   * Parse a time string into various formats
   */
  async parseTime(timeString: string): Promise<TimeParseResponse> {
    const response = await client.post({
      url: '/time/parse',
      body: {
        time_string: timeString,
      } as TimeParseRequest,
    });

    if (response.error) {
      throw new Error((response.error as any).message || 'Failed to parse time');
    }

    return response.data as unknown as TimeParseResponse;
  },

  /**
   * Get list of supported time formats
   */
  async getSupportedFormats(): Promise<SupportedFormatsResponse> {
    const response = await client.get({
      url: '/time/formats',
    });

    if (response.error) {
      throw new Error((response.error as any).message || 'Failed to fetch formats');
    }

    return response.data as unknown as SupportedFormatsResponse;
  },
};

// Kouji Folder API methods
export const koujiFolderApi = {
  /**
   * Get kouji folders from the construction project path
   */
  async getKoujiFolders(path?: string): Promise<KoujiFolderListResponse> {
    const response = await client.get({
      url: '/kouji-folders',
      query: path ? { path } : undefined,
    });

    if (response.error) {
      throw new Error((response.error as any).message || 'Failed to fetch kouji folders');
    }

    return response.data as unknown as KoujiFolderListResponse;
  },
};

// Export a default API object
export const api = {
  folders: folderApi,
  koujiFolders: koujiFolderApi,
  time: timeApi,
};