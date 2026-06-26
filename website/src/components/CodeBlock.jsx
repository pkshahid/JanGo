import './CodeBlock.css'

function CodeBlock({ code, language = 'go', title }) {
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
          <code>{code}</code>
        </pre>
      </div>
    </div>
  )
}

export default CodeBlock
