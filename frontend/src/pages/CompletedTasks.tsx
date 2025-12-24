import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  Card, Button, Typography, Spin, Row, Col
} from 'antd';
import {
  ArrowLeftOutlined, CheckCircleOutlined
} from '@ant-design/icons';
import { fetchCompletedTasks } from '@/api/tasks';
import type { Task } from '@/types';
import { TaskCard } from '@/components/ui';
import { PageHeader } from '@/components/layout/PageHeader';

const { Text } = Typography;

const CompletedTasks: React.FC = () => {
  const navigate = useNavigate();

  const { data: tasks, isLoading, isError } = useQuery({
    queryKey: ['completedTasks'],
    queryFn: fetchCompletedTasks,
  });

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', paddingTop: 100 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        title="已完成任务"
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/tasks')}
          >
            返回任务列表
          </Button>
        }
      />

      {isError ? (
        <Card style={{ textAlign: 'center', padding: 48 }}>
          <Text type="danger">加载任务失败，请稍后重试</Text>
        </Card>
      ) : tasks && tasks.length > 0 ? (
        <Row gutter={[16, 16]}>
          {tasks.map((task: Task, index: number) => (
            <Col key={task.id} xs={24} sm={12} md={12} lg={8} xl={8} xxl={6}>
              <TaskCard
                task={task}
                onClick={() => navigate(`/tasks/${task.id}`)}
                index={index}
              />
            </Col>
          ))}
        </Row>
      ) : (
        <div style={{
          padding: 64,
          textAlign: 'center',
          background: '#ffffff',
          borderRadius: '12px'
        }}>
          <CheckCircleOutlined style={{ fontSize: 64, color: '#d3d3d7', marginBottom: 16 }} />
          <div>
            <Text type="secondary" style={{ fontSize: '16px' }}>
              暂无已完成任务
            </Text>
          </div>
          <div style={{ marginTop: 24 }}>
            <Button type="primary" onClick={() => navigate('/tasks')}>
              查看待办任务
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};

export default CompletedTasks;
