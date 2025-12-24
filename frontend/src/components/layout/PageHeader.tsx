import React from 'react';
import { Typography, Breadcrumb } from 'antd';
import { HomeOutlined } from '@ant-design/icons';

const { Title } = Typography;

interface BreadcrumbItem {
  title: string;
  path?: string;
}

interface PageHeaderProps {
  title: string;
  breadcrumb?: BreadcrumbItem[];
  extra?: React.ReactNode;
}

/**
 * Consistent page header component with optional breadcrumbs
 */
export const PageHeader: React.FC<PageHeaderProps> = ({
  title,
  breadcrumb,
  extra
}) => {
  return (
    <div
      style={{
        marginBottom: '24px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        flexWrap: 'wrap',
        gap: '16px'
      }}
    >
      <div style={{ flex: 1, minWidth: 0 }}>
        {/* Breadcrumb */}
        {breadcrumb && breadcrumb.length > 0 && (
          <Breadcrumb
            style={{ marginBottom: '8px' }}
            items={[
              { title: <HomeOutlined /> },
              ...breadcrumb.map(item => ({
                title: item.title
              }))
            ]}
          />
        )}

        {/* Title */}
        <Title
          level={2}
          style={{
            margin: 0,
            fontSize: '24px',
            fontWeight: 600,
            color: '#18181b'
          }}
          ellipsis
        >
          {title}
        </Title>
      </div>

      {/* Extra actions */}
      {extra && (
        <div style={{ flexShrink: 0 }}>
          {extra}
        </div>
      )}
    </div>
  );
};
