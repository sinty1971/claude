import React, { useState, useEffect } from 'react';
import type { Folder } from '../types/folder';
import { folderService } from '../services/api';
import { FolderModal } from './FolderModal';

export const FolderGrid: React.FC = () => {
  const [folders, setFolders] = useState<Folder[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPath, setCurrentPath] = useState('~/penguin/豊田築炉/2-工事');
  const [pathInput, setPathInput] = useState('~/penguin/豊田築炉/2-工事');
  const [selectedFolder, setSelectedFolder] = useState<Folder | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const loadFolders = async (path?: string) => {
    setLoading(true);
    setError(null);
    
    try {
      console.log('Loading folders for path:', path || 'default');
      const response = await folderService.getFolders(path);
      console.log('API Response:', response);
      setFolders(response.folders);
      setCurrentPath(response.path);
    } catch (err) {
      console.error('Error loading folders:', err);
      setError(err instanceof Error ? err.message : 'エラーが発生しました');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFolders();
  }, []);

  const handleFolderClick = (folder: Folder) => {
    setSelectedFolder(folder);
    setIsModalOpen(true);
  };

  const handlePathSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    loadFolders(pathInput);
  };


  const getFolderIcon = (folder: Folder) => {
    if (folder.is_directory) {
      return '📁';
    }
    const ext = folder.name.split('.').pop()?.toLowerCase();
    switch (ext) {
      case 'pdf': return '📄';
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif': return '🖼️';
      case 'mp4':
      case 'avi':
      case 'mov': return '🎬';
      case 'mp3':
      case 'wav': return '🎵';
      default: return '📄';
    }
  };

  return (
    <div className="folder-container">
      <div className="header">
        <h1>フォルダー管理システム</h1>
        
        <form onSubmit={handlePathSubmit} className="path-form">
          <input
            type="text"
            value={pathInput}
            onChange={(e) => setPathInput(e.target.value)}
            placeholder="フォルダーパスを入力"
            className="path-input"
          />
          <button type="submit" className="load-button">読み込み</button>
        </form>
      </div>

      <div className="folder-info">
        <span className="folder-count">{folders.length} 項目</span>
        <span className="current-path">{currentPath}</span>
      </div>

      {loading && <div className="loading">読み込み中...</div>}
      {error && <div className="error">{error}</div>}

      <div className="folder-list">
        {folders.map((folder, index) => (
          <div
            key={index}
            className="folder-item"
            onClick={() => handleFolderClick(folder)}
          >
            <div className="folder-icon">{getFolderIcon(folder)}</div>
            <div className="folder-info">
              <div className="folder-name">{folder.name}</div>
              <div className="folder-meta">
                <span>{folder.is_directory ? 'フォルダー' : 'ファイル'}</span>
                {folder.created_date && (
                  <span className="folder-date">
                    {' · 作成: '}
                    {new Date(folder.created_date).toLocaleDateString('ja-JP', {
                      year: 'numeric',
                      month: '2-digit',
                      day: '2-digit',
                      hour: '2-digit',
                      minute: '2-digit'
                    })}
                  </span>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      <FolderModal
        folder={selectedFolder}
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
      />
    </div>
  );
};