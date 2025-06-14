import React from 'react';
import type { Folder } from '../types/folder';

interface FolderModalProps {
  folder: Folder | null;
  isOpen: boolean;
  onClose: () => void;
}

export const FolderModal: React.FC<FolderModalProps> = ({ folder, isOpen, onClose }) => {
  if (!isOpen || !folder) return null;

  const formatSize = (bytes: number): string => {
    const units = ['B', 'KB', 'MB', 'GB'];
    let size = bytes;
    let unitIndex = 0;
    
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }
    
    return `${size.toFixed(1)} ${units[unitIndex]}`;
  };

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleString('ja-JP');
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>{folder.name}</h2>
          <button className="close-button" onClick={onClose}>×</button>
        </div>
        
        <div className="modal-body">
          <div className="info-row">
            <span className="label">種類:</span>
            <span className="value">{folder.is_directory ? 'フォルダー' : 'ファイル'}</span>
          </div>
          
          <div className="info-row">
            <span className="label">ID:</span>
            <span className="value">{folder.id}</span>
          </div>
          
          <div className="info-row">
            <span className="label">サイズ:</span>
            <span className="value">{formatSize(folder.size)}</span>
          </div>
          
          <div className="info-row">
            <span className="label">パス:</span>
            <span className="value">{folder.path}</span>
          </div>
          
          
          <div className="info-row">
            <span className="label">更新日時:</span>
            <span className="value">{formatDate(folder.modified_time)}</span>
          </div>
        </div>
      </div>
    </div>
  );
};