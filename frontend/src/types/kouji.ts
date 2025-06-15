// Extended KoujiEntry type with company and location names
export interface KoujiEntryExtended {
  name: string;
  path: string;
  is_directory: boolean;
  size: number;
  modified_time: string;
  project_id?: string;
  project_name?: string;
  company_name?: string;
  location_name?: string;
  status?: string;
  created_date?: string;
  start_date?: string;
  end_date?: string;
  description?: string;
  tags?: string[];
  file_count?: number;
  subdir_count?: number;
}

export interface KoujiEntriesResponseExtended {
  entries: KoujiEntryExtended[];
  count: number;
  path: string;
  total_size?: number;
}