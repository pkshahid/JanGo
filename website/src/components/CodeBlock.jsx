import { useEffect, useRef } from 'react'
import Prism from 'prismjs'
import 'prismjs/components/prism-go'
import 'prismjs/components/prism-bash'
import 'prismjs/components/prism-markup'
import 'prismjs/themes/prism-tomorrow.css'
import './CodeBlock.css'

function CodeBlock({ code, language = 'go', title }) {
  const codeRef = useRef(null)

  useEffect(() => {
    if (codeRef.current) {
      Prism.highlightElement(codeRef.current)
    }
  }, [code, language])

  return (
    <div className="code-block-wrapper">
      {title && (
        <div className="code-block-header">
          <span className="code-block-title">{title}</span>
          <span className="code-block-lang">{language}</span>
        </div>
      )}
      <div className="code-block">
        <pre>
          <code ref={codeRef} className={`language-${language}`}>{code}</code>
        </pre>
      </div>
    </div>
  )
}

export default CodeBlock
