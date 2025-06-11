import { useState, useEffect } from 'react';
import { api, type KoujiFolder } from '../api/client';

const KoujiFolderGrid = () => {
  const [folders, setFolders] = useState<KoujiFolder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [path, setPath] = useState('');
  const [totalSize, setTotalSize] = useState<number>(0);

  const loadKoujiFolders = async (targetPath?: string) => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await api.koujiFolders.getKoujiFolders(targetPath);
      setFolders(response.folders);
      setPath(response.path);
      setTotalSize(response.total_size || 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load kouji folders');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadKoujiFolders();
  }, []);

  const getStatusColor = (status?: string) => {
    switch (status) {
      case '進行中':
        return '#4CAF50';
      case '完了':
        return '#9E9E9E';
      case '予定':
        return '#FF9800';
      default:
        return '#2196F3';
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString('ja-JP');
  };

  if (loading) {
    return <div className="loading">工事フォルダーを読み込み中...</div>;
  }

  if (error) {
    return (
      <div className="error">
        <p>エラー: {error}</p>
        <button onClick={() => loadKoujiFolders()}>再試行</button>
      </div>
    );
  }

  return (
    <div className="folder-container">
      <div className="folder-header">
        <h2>工事フォルダー一覧</h2>
        <div className="folder-info">
          <p>パス: {path}</p>
          <p>フォルダー数: {folders.length}</p>
          {totalSize > 0 && <p>合計サイズ: {formatFileSize(totalSize)}</p>}
        </div>
      </div>

      {folders.length === 0 ? (
        <div className="empty-state">
          <p>工事フォルダーが見つかりませんでした</p>
          <p>フォルダー名は「YYYY-MMDD 会社名 現場名」の形式である必要があります</p>
        </div>
      ) : (
        <div className="folder-grid">
          {folders.map((folder, index) => (
            <div key={index} className="folder-item kouji-folder-item">
              <div className="folder-icon">
                📁
              </div>
              <div className="folder-details">
                <div className="folder-name" title={folder.name}>
                  {folder.name}
                </div>
                
                <div className="kouji-metadata">
                  <div className="project-info">
                    <div className="project-id">{folder.project_id}</div>
                    <div className="project-name">{folder.project_name}</div>
                  </div>
                  
                  <div className="project-status">
                    <span 
                      className="status-badge"
                      style={{ backgroundColor: getStatusColor(folder.status) }}
                    >
                      {folder.status}
                    </span>
                  </div>
                  
                  <div className="project-dates">
                    <div>開始: {formatDate(folder.start_date)}</div>
                    <div>終了: {formatDate(folder.end_date)}</div>
                  </div>
                  
                  {folder.description && (
                    <div className="project-description">
                      {folder.description}
                    </div>
                  )}
                  
                  {folder.tags && folder.tags.length > 0 && (
                    <div className="project-tags">
                      {folder.tags.map((tag, tagIndex) => (
                        <span key={tagIndex} className="tag">
                          {tag}
                        </span>
                      ))}
                    </div>
                  )}
                </div>

                <div className="folder-meta">
                  <span>{formatFileSize(folder.size)}</span>
                  <span>{formatDate(folder.modified_time)}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default KoujiFolderGrid;