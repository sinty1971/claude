import { useState, useEffect } from 'react';
import { api } from '../api/client';
import type { KoujiEntryExtended } from '../types/kouji';
import DateEditModal from './DateEditModal';

const KoujiEntriesGrid = () => {
  const [projects, setProjects] = useState<KoujiEntryExtended[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [path, setPath] = useState('');
  const [totalSize, setTotalSize] = useState<number>(0);
  const [editingProject, setEditingProject] = useState<KoujiEntryExtended | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const loadKoujiEntries = async (targetPath?: string) => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await api.koujiEntries.getKoujiEntries(targetPath);
      
      // Sort entries by date in descending order (newest first)
      const sortedEntries = (response.folders as unknown as KoujiEntryExtended[]).sort((a, b) => {
        const dateA = new Date(a.start_date || '');
        const dateB = new Date(b.start_date || '');
        return dateB.getTime() - dateA.getTime();
      });
      
      setProjects(sortedEntries);
      setPath(response.path || targetPath || '');
      setTotalSize(response.total_size || 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load kouji entries');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadKoujiEntries();
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

  const handleEditDates = (project: KoujiEntryExtended) => {
    setEditingProject(project);
    setIsModalOpen(true);
  };

  const handleDateUpdateSuccess = () => {
    // Reload entries to get updated data
    loadKoujiEntries();
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setEditingProject(null);
  };

  if (loading) {
    return <div className="loading">工事プロジェクトを読み込み中...</div>;
  }

  if (error) {
    return (
      <div className="error">
        <p>エラー: {error}</p>
        <button onClick={() => loadKoujiEntries()}>再試行</button>
      </div>
    );
  }

  return (
    <div className="folder-container">
      <div className="folder-header">
        <h2>工事プロジェクト一覧</h2>
        <div className="folder-info">
          <p>パス: {path}</p>
          <p>プロジェクト数: {projects.length}</p>
          {totalSize > 0 && <p>合計サイズ: {formatFileSize(totalSize)}</p>}
        </div>
      </div>

      {projects.length === 0 ? (
        <div className="empty-state">
          <p>工事プロジェクトが見つかりませんでした</p>
          <p>フォルダー名は「YYYY-MMDD 会社名 現場名」の形式である必要があります</p>
        </div>
      ) : (
        <div className="folder-grid">
          {projects.map((project, index) => (
            <div key={index} className="folder-item kouji-folder-item">
              <div className="folder-icon">
                📁
              </div>
              <div className="folder-details">
                <div className="folder-name" title={project.name}>
                  {project.name}
                </div>
                
                <div className="kouji-metadata">
                  <div className="project-info">
                    <div className="project-id">{project.project_id}</div>
                    <div className="project-name">{project.project_name}</div>
                  </div>
                  
                  {(project.company_name || project.location_name) && (
                    <div className="company-location-info">
                      {project.company_name && (
                        <span className="company-name">会社: {project.company_name}</span>
                      )}
                      {project.location_name && (
                        <span className="location-name">現場: {project.location_name}</span>
                      )}
                    </div>
                  )}
                  
                  <div className="project-status">
                    <span 
                      className="status-badge"
                      style={{ backgroundColor: getStatusColor(project.status) }}
                    >
                      {project.status}
                    </span>
                  </div>
                  
                  <div className="project-dates">
                    <div>開始: {formatDate(project.start_date)}</div>
                    <div>終了: {formatDate(project.end_date)}</div>
                    <button 
                      className="edit-dates-button"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleEditDates(project);
                      }}
                    >
                      日付編集
                    </button>
                  </div>
                  
                  {project.description && (
                    <div className="project-description">
                      {project.description}
                    </div>
                  )}
                  
                  {project.tags && project.tags.length > 0 && (
                    <div className="project-tags">
                      {project.tags.map((tag, tagIndex) => (
                        <span key={tagIndex} className="tag">
                          {tag}
                        </span>
                      ))}
                    </div>
                  )}
                </div>

                <div className="folder-meta">
                  <span>{formatFileSize(project.size)}</span>
                  <span>{formatDate(project.modified_time)}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {editingProject && (
        <DateEditModal
          isOpen={isModalOpen}
          onClose={closeModal}
          projectId={editingProject.project_id || ''}
          projectName={editingProject.project_name || editingProject.name}
          currentStartDate={editingProject.start_date || ''}
          currentEndDate={editingProject.end_date || ''}
          onSuccess={handleDateUpdateSuccess}
        />
      )}
    </div>
  );
};

export default KoujiEntriesGrid;