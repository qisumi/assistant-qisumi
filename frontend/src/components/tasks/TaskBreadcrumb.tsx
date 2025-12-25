import React from 'react';
import { Breadcrumb } from 'antd';
import { useNavigate } from 'react-router-dom';
import type { Task } from '@/types';

interface TaskBreadcrumbProps {
  task: Task;
}

export const TaskBreadcrumb: React.FC<TaskBreadcrumbProps> = ({ task }) => {
  const navigate = useNavigate();

  return (
    <Breadcrumb
      style={{ marginBottom: 16 }}
      items={[
        { title: <a onClick={() => navigate('/tasks')}>任务列表</a> },
        { title: task.title },
      ]}
    />
  );
};
