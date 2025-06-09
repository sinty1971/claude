import type { FolderListResponse } from '../types/folder';

const API_BASE = 'http://localhost:8080/api';

export const folderService = {
  async getFolders(path?: string): Promise<FolderListResponse> {
    const url = new URL(`${API_BASE}/folders`);
    if (path) {
      url.searchParams.set('path', path);
    }

    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  }
};