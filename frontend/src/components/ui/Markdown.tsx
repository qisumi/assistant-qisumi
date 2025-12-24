import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import rehypeRaw from 'rehype-raw';
import { Typography } from 'antd';
import type { Components } from 'react-markdown';

const { Text, Paragraph } = Typography;

interface MarkdownProps {
  content: string;
  className?: string;
}

/**
 * Markdown renderer with syntax highlighting and GFM support
 */
export const Markdown: React.FC<MarkdownProps> = ({ content, className }) => {
  if (!content || !content.trim()) {
    return <Text type="secondary">暂无内容</Text>;
  }

  const components: Components = {
    // Paragraphs
    p: ({ children }) => <Paragraph style={{ marginBottom: '1em' }}>{children}</Paragraph>,

    // Links
    a: ({ children, href }) => (
      <a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        style={{ color: '#5E6AD2' }}
      >
        {children}
      </a>
    ),

    // Code blocks (inline)
    code: ({ className, children }) => {
      const isInline = !className;
      return isInline ? (
        <Text
          code
          style={{
            padding: '0.2em 0.4em',
            margin: '0 0.1em',
            fontSize: '85%',
            backgroundColor: 'rgba(175, 184, 193, 0.2)',
            borderRadius: '4px'
          }}
        >
          {children}
        </Text>
      ) : (
        <code className={className}>{children}</code>
      );
    },

    // Headings
    h1: ({ children }) => (
      <h1 style={{ fontSize: '2em', marginTop: '1.5em', marginBottom: '0.5em', fontWeight: 600 }}>{children}</h1>
    ),
    h2: ({ children }) => (
      <h2 style={{ fontSize: '1.5em', marginTop: '1.5em', marginBottom: '0.5em', fontWeight: 600 }}>{children}</h2>
    ),
    h3: ({ children }) => (
      <h3 style={{ fontSize: '1.25em', marginTop: '1.5em', marginBottom: '0.5em', fontWeight: 600 }}>{children}</h3>
    ),

    // Lists
    ul: ({ children }) => <ul style={{ paddingLeft: '2em', marginBottom: '1em' }}>{children}</ul>,
    ol: ({ children }) => <ol style={{ paddingLeft: '2em', marginBottom: '1em' }}>{children}</ol>,
    li: ({ children }) => <li style={{ marginBottom: '0.25em' }}>{children}</li>,

    // Blockquotes
    blockquote: ({ children }) => (
      <blockquote
        style={{
          padding: '0 1em',
          color: '#6b6b76',
          borderLeft: '0.25em solid #d3d3d7',
          marginBottom: '1em'
        }}
      >
        {children}
      </blockquote>
    ),

    // Tables
    table: ({ children }) => (
      <div style={{ overflowX: 'auto', marginBottom: '1em' }}>
        <table style={{ borderSpacing: 0, borderCollapse: 'collapse', width: '100%' }}>
          {children}
        </table>
      </div>
    ),
    thead: ({ children }) => <thead>{children}</thead>,
    tbody: ({ children }) => <tbody>{children}</tbody>,
    tr: ({ children }) => <tr>{children}</tr>,
    th: ({ children }) => (
      <th
        style={{
          padding: '6px 13px',
          border: '1px solid #d3d3d7',
          fontWeight: 600,
          backgroundColor: '#f7f7f8'
        }}
      >
        {children}
      </th>
    ),
    td: ({ children }) => (
      <td style={{ padding: '6px 13px', border: '1px solid #d3d3d7' }}>{children}</td>
    ),

    // Images
    img: ({ src, alt }) => (
      <img
        src={src}
        alt={alt}
        style={{ maxWidth: '100%', height: 'auto', borderRadius: '8px', margin: '1em 0' }}
      />
    ),

    // Strong
    strong: ({ children }) => <strong>{children}</strong>,

    // Emphasis
    em: ({ children }) => <em>{children}</em>,

    // HR
    hr: () => <hr style={{ border: 'none', borderTop: '1px solid #d3d3d7', margin: '2em 0' }} />,
  };

  return (
    <div className={`markdown-content ${className || ''}`}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeHighlight, rehypeRaw]}
        components={components}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
};
