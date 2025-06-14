import { BrowserRouter as Router, Routes, Route, Link, useLocation } from 'react-router-dom';
import { FolderGrid } from './components/FolderGrid';
import KoujiProjectPage from './components/KoujiProjectPage';
import './App.css';

function Navigation() {
  const location = useLocation();
  
  return (
    <nav className="navigation">
      <div className="nav-container">
        <div className="nav-logo">
          <h1>Penguin フォルダー管理</h1>
        </div>
        <div className="nav-links">
          <Link 
            to="/" 
            className={location.pathname === '/' ? 'nav-link active' : 'nav-link'}
          >
            フォルダー一覧
          </Link>
          <Link 
            to="/kouji" 
            className={location.pathname === '/kouji' ? 'nav-link active' : 'nav-link'}
          >
            工事一覧
          </Link>
        </div>
      </div>
    </nav>
  );
}

function App() {
  return (
    <Router>
      <div className="app">
        <Navigation />
        <main className="main-content">
          <Routes>
            <Route path="/" element={<FolderGrid />} />
            <Route path="/kouji" element={<KoujiProjectPage />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
