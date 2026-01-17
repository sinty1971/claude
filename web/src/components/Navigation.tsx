import { Link, useLocation } from '@remix-run/react';
import '../styles/navigation.css';

export function Navigation() {
  const { pathname } = useLocation();

  return (
    <nav className="navigation">
      <div className="nav-container">
        <div className="nav-links">
          <Link
            to="/"
            className={pathname === '/' ? 'nav-link active' : 'nav-link'}
          >
            ホーム
          </Link>
          <Link
            to="/files"
            className={pathname === '/files' ? 'nav-link active' : 'nav-link'}
          >
            ファイル一覧
          </Link>
          <Link
            to="/kojies"
            className={pathname === '/kojies' ? 'nav-link active' : 'nav-link'}
          >
            工事一覧
          </Link>
          <Link
            to="/companies"
            className={pathname === '/companies' ? 'nav-link active' : 'nav-link'}
          >
            会社一覧
          </Link>
        </div>
        <div className="nav-logo">
          <h1>Penguin フォルダー管理</h1>
        </div>
      </div>
    </nav>
  );
}
