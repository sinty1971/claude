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
export type KoujiEntry = components['schemas']['KoujiFolder'];
export type KoujiEntriesResponse = components['schemas']['KoujiFolderListResponse'];

// API methods
export const folderApi = {
  /**
   * Get folders from the specified path
   */
  async getFolders(path?: string): Promise<FolderListResponse> {
    const response = await client.get({
      url: '/file-entries',
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

// Kouji Entries API methods
export const koujiEntriesApi = {
  /**
   * Get kouji entries from the construction project path
   */
  async getKoujiEntries(path?: string): Promise<KoujiEntriesResponse> {
    const response = await client.get({
      url: '/kouji-entries',
      query: path ? { path } : undefined,
    });

    if (response.error) {
      throw new Error((response.error as any).message || 'Failed to fetch kouji entries');
    }

    return response.data as unknown as KoujiEntriesResponse;
  },

  /**
   * Save kouji entries to YAML file
   */
  async saveKoujiEntries(path?: string, outputPath?: string): Promise<{message: string}> {
    const response = await client.post({
      url: '/kouji-entries/save',
      query: {
        ...(path && { path }),
        ...(outputPath && { output_path: outputPath }),
      },
    });

    if (response.error) {
      throw new Error((response.error as any).message || 'Failed to save kouji entries');
    }

    return response.data as unknown as {message: string};
  },

  /**
   * Update kouji entry dates
   */
  async updateKoujiEntryDates(projectId: string, startDate: string, endDate: string): Promise<{message: string; project_id: string}> {
    const response = await client.put({
      url: `/kouji-entries/${projectId}/dates`,
      headers: {
        'Content-Type': 'application/json',
      },
      body: {
        start_date: startDate,
        end_date: endDate,
      },
    });

    if (response.error) {
      const errorMessage = (response.error as any)?.message || 
                          (response.error as any)?.error || 
                          JSON.stringify(response.error) || 
                          'Failed to update entry dates';
      throw new Error(errorMessage);
    }

    return response.data as unknown as {message: string; project_id: string};
  },
};

// Export a default API object
export const api = {
  folders: folderApi,
  koujiEntries: koujiEntriesApi,
  time: timeApi,
};