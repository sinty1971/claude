import { useState } from 'react';
import KoujiFolderGrid from './KoujiFolderGrid';

const KoujiFolderPage = () => {
  const [showHelp, setShowHelp] = useState(false);

  return (
    <div className="page-container">
      <div className="page-header">
        <h1>工事プロジェクト管理</h1>
        <div className="header-actions">
          <button 
            className="help-button"
            onClick={() => setShowHelp(!showHelp)}
          >
            ヘルプ
          </button>
        </div>
      </div>

      {showHelp && (
        <div className="help-section">
          <h3>工事フォルダーについて</h3>
          <ul>
            <li>フォルダー名は「YYYY-MMDD 会社名 現場名」の形式で命名してください</li>
            <li>例: 「2025-0618 豊田築炉 名和工場」</li>
            <li>ステータスは日付に基づいて自動的に判定されます：
              <ul>
                <li><span style={{ color: '#FF9800' }}>予定</span>: 開始日が未来の場合</li>
                <li><span style={{ color: '#4CAF50' }}>進行中</span>: 現在進行中の場合</li>
                <li><span style={{ color: '#9E9E9E' }}>完了</span>: 終了日を過ぎた場合</li>
              </ul>
            </li>
            <li>プロジェクトの期間は開始日から3ヶ月間と仮定されます</li>
          </ul>
        </div>
      )}

      <KoujiFolderGrid />
    </div>
  );
};

export default KoujiFolderPage;