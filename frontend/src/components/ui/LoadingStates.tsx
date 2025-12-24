import React from 'react';
import { Spin, Card, Skeleton } from 'antd';

/**
 * Full-page spinner for loading states
 */
export const PageSpinner: React.FC<{ size?: 'small' | 'default' | 'large' }> = ({
  size = 'default'
}) => {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '400px',
        width: '100%'
      }}
    >
      <Spin size={size} />
    </div>
  );
};

/**
 * Content skeleton for card-based layouts
 */
export const ContentSkeleton: React.FC<{ count?: number }> = ({ count = 3 }) => {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {Array.from({ length: count }).map((_, index) => (
        <Card key={index}>
          <Skeleton active paragraph={{ rows: 3 }} />
        </Card>
      ))}
    </div>
  );
};

/**
 * Inline loader for small components
 */
export const InlineLoader: React.FC<{ text?: string }> = ({ text }) => {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <Spin size="small" />
      {text && <span style={{ color: '#6b6b76' }}>{text}</span>}
    </div>
  );
};
