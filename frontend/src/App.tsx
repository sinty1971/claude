import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { FolderGrid } from './components/FolderGrid';
import './App.css';

function App() {
  return (
    <Router>
      <div className="app">
        <Routes>
          <Route path="/" element={<FolderGrid />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
