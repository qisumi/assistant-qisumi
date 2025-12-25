import React from 'react';
import { Button, Space } from 'antd';
import { ArrowLeftOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';

interface TaskActionsProps {
  onBack: () => void;
  onEdit: () => void;
  onDelete: () => void;
}

export const TaskActions: React.FC<TaskActionsProps> = ({ onBack, onEdit, onDelete }) => {
  return (
    <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 12 }}>
      <Space wrap size="small">
        <Button icon={<ArrowLeftOutlined />} onClick={onBack} size="small">
          返回
        </Button>
        <Button icon={<EditOutlined />} onClick={onEdit} size="small">
          编辑
        </Button>
        <Button danger icon={<DeleteOutlined />} onClick={onDelete} size="small">
          删除
        </Button>
      </Space>
    </div>
  );
};
