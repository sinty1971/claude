export interface Folder {
  name: string;
  path: string;
  is_directory: boolean;
  size: number;
  modified_time: string;
}

export interface FolderListResponse {
  folders: Folder[];
  count: number;
  path: string;
}