import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import type { Folder } from '../types/folder';
import { folderService } from '../services/api';
import { FileEntryModal } from './FileEntryModal';

export const FileEntryGrid: React.FC = () => {
  const navigate = useNavigate();
  const [folders, setFolders] = useState<Folder[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPath, setCurrentPath] = useState('~/penguin');
  const [pathInput, setPathInput] = useState('~/penguin');
  const [selectedFolder, setSelectedFolder] = useState<Folder | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  // 工事プロジェクトディレクトリかどうかをチェック
  const isKoujiProjectPath = (path: string) => {
    const normalizedPath = path.replace(/\\/g, '/');
    return normalizedPath.includes('/豊田築炉/2-工事') || 
           normalizedPath.endsWith('/2-工事') ||
           normalizedPath.includes('2-工事');
  };

  const loadFolders = async (path?: string) => {
    const targetPath = path || '~/penguin';
    
    // 工事プロジェクトディレクトリの場合は工事プロジェクトページにリダイレクト
    if (isKoujiProjectPath(targetPath)) {
      navigate('/kouji');
      return;
    }

    setLoading(true);
    setError(null);
    
    try {
      console.log('Loading folders for path:', targetPath);
      const response = await folderService.getFolders(path);
      console.log('API Response:', response);
      setFolders(response.folders);
      setCurrentPath(targetPath);
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
    if (folder.is_directory) {
      // ディレクトリの場合は移動
      const newPath = folder.path;
      
      // 工事プロジェクトディレクトリの場合は工事プロジェクトページにリダイレクト
      if (isKoujiProjectPath(newPath)) {
        navigate('/kouji');
        return;
      }
      
      setPathInput(newPath);
      loadFolders(newPath);
    } else {
      // ファイルの場合はモーダル表示
      setSelectedFolder(folder);
      setIsModalOpen(true);
    }
  };

  const handlePathSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // /home/shin/penguinより親に行かないようにバリデーション
    const minPath = '/home/shin/penguin';
    if (pathInput.startsWith(minPath) || pathInput === minPath) {
      loadFolders(pathInput);
    } else {
      // バリデーションエラーの場合、最小パスに設定
      setPathInput(minPath);
      loadFolders(minPath);
    }
  };

  const handleGoBack = () => {
    // 親ディレクトリのパスを取得
    const pathParts = currentPath.split('/');
    if (pathParts.length > 1) {
      const parentPath = pathParts.slice(0, -1).join('/');
      const newPath = parentPath || '/';
      
      // /home/shin/penguinより親に行かないようにバリデーション
      const minPath = '/home/shin/penguin';
      if (newPath.startsWith(minPath) || newPath === minPath) {
        setPathInput(newPath);
        loadFolders(newPath);
      }
    }
  };


  // 特別なフォルダーかどうかをチェック
  const isSpecialFolder = (folder: Folder) => {
    if (!folder.is_directory) return false;
    return isKoujiProjectPath(folder.path) || folder.name === '2-工事';
  };

  const getFolderIcon = (folder: Folder) => {
    if (folder.is_directory) {
      // 工事プロジェクトフォルダーの場合は特別なアイコン
      if (isSpecialFolder(folder)) {
        return '🏗️';
      }
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
          <button type="button" onClick={handleGoBack} className="back-button">
            <span className="back-arrow">⮜</span>
          </button>
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
        {folders.map((folder, index) => {
          const isSpecial = isSpecialFolder(folder);
          return (
            <div
              key={index}
              className={`folder-item ${isSpecial ? 'folder-item--special' : ''}`}
              onClick={() => handleFolderClick(folder)}
            >
              <div className={`folder-icon ${isSpecial ? 'folder-icon--special' : ''}`}>
                {getFolderIcon(folder)}
              </div>
              <div className="folder-info">
                <div className={`folder-name ${isSpecial ? 'folder-name--special' : ''}`}>
                  {folder.name}
                  {isSpecial && <span className="special-badge">工事一覧</span>}
                </div>
                <div className="folder-meta">
                  <span>{folder.is_directory ? 'フォルダー' : 'ファイル'}</span>
                  <span className="folder-date">
                    {' · 更新: '}
                    {new Date(folder.modified_time).toLocaleDateString('ja-JP', {
                      year: 'numeric',
                      month: '2-digit',
                      day: '2-digit',
                      hour: '2-digit',
                      minute: '2-digit'
                    })}
                  </span>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      <FileEntryModal
        folder={selectedFolder}
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
      />
    </div>
  );
};