import { Routes, Route } from 'react-router-dom'
import Navbar from './components/Navbar'
import Footer from './components/Footer'
import Home from './pages/Home'
import About from './pages/About'
import Features from './pages/Features'
import Documentation from './pages/Documentation'
import DeveloperGuides from './pages/DeveloperGuides'
import Configuration from './pages/Configuration'

function App() {
  return (
    <div className="app">
      <Navbar />
      <main>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/about" element={<About />} />
          <Route path="/features" element={<Features />} />
          <Route path="/docs" element={<Documentation />} />
          <Route path="/guides" element={<DeveloperGuides />} />
          <Route path="/configuration" element={<Configuration />} />
        </Routes>
      </main>
      <Footer />
    </div>
  )
}

export default App
