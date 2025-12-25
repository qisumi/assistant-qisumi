import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, Button, Typography, Spin, Divider, List, Row, Col, App } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { fetchTaskDetail, deleteTask, deleteTaskStep } from '@/api/tasks';
import { fetchSessionMessages, sendSessionMessage } from '@/api/sessions';
import { ChatWindow } from '@/components/chat/ChatWindow';
import { TaskEditForm } from '@/components/tasks/TaskEditForm';
import {
  TaskBreadcrumb,
  TaskActions,
  TaskHeader,
  TaskDescription,
  TaskTimeInfo,
  StepListItem,
  AddStepModal,
} from '@/components/tasks';
import { confirmDelete } from '@/utils/dialog';

const TaskDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [editingTask, setEditingTask] = React.useState(false);
  const [isAddStepModalVisible, setIsAddStepModalVisible] = React.useState(false);
  const { message } = App.useApp();

  const { data, isLoading, isError } = useQuery({
    queryKey: ['taskDetail', id],
    queryFn: () => fetchTaskDetail(id!),
    enabled: !!id,
  });

  const sessionId = data?.session.id;

  const { data: messagesData } = useQuery({
    queryKey: ['sessionMessages', sessionId],
    queryFn: () => fetchSessionMessages(sessionId!),
    enabled: !!sessionId,
  });

  const sendMutation = useMutation({
    mutationFn: (content: string) => sendSessionMessage(sessionId!, content),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['taskDetail', id] });
      if (sessionId) {
        queryClient.invalidateQueries({ queryKey: ['sessionMessages', sessionId] });
      }
    },
    onError: (err: any) => {
      console.error(err);
      message.error('发送失败，请稍后重试');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteTask,
    onSuccess: () => {
      message.success('任务已删除');
      navigate('/tasks');
    },
    onError: () => {
      message.error('删除失败，请稍后重试');
    },
  });

  const deleteStepMutation = useMutation({
    mutationFn: ({ taskId, stepId }: { taskId: string | number; stepId: string | number }) =>
      deleteTaskStep(taskId, stepId),
    onSuccess: () => {
      message.success('步骤已删除');
      queryClient.invalidateQueries({ queryKey: ['taskDetail', id] });
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '删除失败，请稍后重试');
    },
  });

  const handleDeleteTask = () => {
    if (data) {
      confirmDelete('任务', data.task.title, () => deleteMutation.mutate(data.task.id));
    }
  };

  const handleDeleteStep = (stepId: number, stepTitle: string) => {
    confirmDelete('步骤', stepTitle, () => deleteStepMutation.mutate({ taskId: id!, stepId }));
  };

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', paddingTop: 100 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div style={{ padding: 24 }}>
        <Card>
          <Typography.Text type="danger">加载任务详情失败</Typography.Text>
          <Button onClick={() => navigate('/tasks')} style={{ marginLeft: 16 }}>
            返回列表
          </Button>
        </Card>
      </div>
    );
  }

  const { task } = data;
  const messages = messagesData?.messages ?? data.messages ?? [];

  return (
    <div>
      {/* Breadcrumb */}
      <TaskBreadcrumb task={task} />

      {/* Responsive layout */}
      <Row gutter={[24, 24]}>
        {/* Left: Task info */}
        <Col xs={24} lg={14} xl={16}>
          <Card>
            {editingTask ? (
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
                  <Typography.Title level={4} style={{ margin: 0 }}>
                    编辑任务
                  </Typography.Title>
                  <Button onClick={() => setEditingTask(false)}>取消</Button>
                </div>
                <TaskEditForm
                  task={task}
                  onCancel={() => setEditingTask(false)}
                  onSuccess={() => setEditingTask(false)}
                />
              </div>
            ) : (
              <>
                {/* Action buttons */}
                <TaskActions
                  onBack={() => navigate('/tasks')}
                  onEdit={() => setEditingTask(true)}
                  onDelete={handleDeleteTask}
                />

                {/* Title and badges */}
                <TaskHeader task={task} />

                <Divider />

                {/* Description with markdown */}
                <TaskDescription task={task} />

                {/* Time info */}
                <TaskTimeInfo task={task} />

                {/* Steps */}
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                    <Typography.Title level={5} style={{ margin: 0 }}>
                      执行步骤
                    </Typography.Title>
                    <Button
                      type="dashed"
                      icon={<PlusOutlined />}
                      onClick={() => setIsAddStepModalVisible(true)}
                      size="small"
                    >
                      添加步骤
                    </Button>
                  </div>
                  <List
                    dataSource={task.steps || []}
                    renderItem={(step: any, index: number) => (
                      <StepListItem
                        key={step.id}
                        step={step}
                        taskId={task.id}
                        index={index}
                        totalSteps={task.steps?.length || 0}
                        onDelete={handleDeleteStep}
                      />
                    )}
                  />
                </div>
              </>
            )}
          </Card>
        </Col>

        {/* Right: Chat */}
        <Col xs={24} lg={10} xl={8}>
          <Card
            title="小奇（任务）"
            styles={{ body: { padding: 0 } }}
            style={{ position: 'sticky', top: 24 }}
          >
            <ChatWindow
              messages={messages}
              onSend={(content) => sendMutation.mutateAsync(content)}
              sending={sendMutation.isPending}
              height={600}
            />
          </Card>
        </Col>
      </Row>

      {/* Add step modal */}
      <AddStepModal
        open={isAddStepModalVisible}
        taskId={task.id}
        stepsCount={task.steps?.length || 0}
        onCancel={() => setIsAddStepModalVisible(false)}
        onSuccess={() => setIsAddStepModalVisible(false)}
      />
    </div>
  );
};

export default TaskDetail;
