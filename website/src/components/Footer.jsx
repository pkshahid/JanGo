import { Link } from 'react-router-dom'
import './Footer.css'

function Footer() {
  return (
    <footer className="footer">
      <div className="container">
        <div className="footer-grid">
          <div className="footer-brand">
            <Link to="/" className="navbar-brand">
              <span className="brand-icon">J</span>
              <span className="brand-text">JanGo</span>
            </Link>
            <p className="footer-desc">
              A full Django-equivalent web framework written in Go.
              Batteries included, blazing fast.
            </p>
          </div>

          <div className="footer-col">
            <h4>Framework</h4>
            <Link to="/features">Features</Link>
            <Link to="/about">About</Link>
            <Link to="/docs">Documentation</Link>
            <Link to="/guides">Developer Guides</Link>
          </div>

          <div className="footer-col">
            <h4>Resources</h4>
            <Link to="/configuration">Configuration</Link>
            <a href="https://github.com/pkshahid/JanGo" target="_blank" rel="noopener noreferrer">GitHub</a>
            <a href="https://github.com/pkshahid/JanGo/issues" target="_blank" rel="noopener noreferrer">Issues</a>
            <a href="https://github.com/pkshahid/JanGo/releases" target="_blank" rel="noopener noreferrer">Releases</a>
          </div>

          <div className="footer-col">
            <h4>Community</h4>
            <a href="https://github.com/pkshahid/JanGo/discussions" target="_blank" rel="noopener noreferrer">Discussions</a>
            <a href="https://github.com/pkshahid/JanGo/blob/main/LICENSE" target="_blank" rel="noopener noreferrer">License</a>
            <a href="https://github.com/pkshahid/JanGo#contributing" target="_blank" rel="noopener noreferrer">Contributing</a>
          </div>
        </div>

        <div className="footer-bottom">
          <p>&copy; {new Date().getFullYear()} JanGo Framework. Built with Go.</p>
          <p className="footer-tagline">Django&apos;s power, Go&apos;s speed.</p>
        </div>
      </div>
    </footer>
  )
}

export default Footer
