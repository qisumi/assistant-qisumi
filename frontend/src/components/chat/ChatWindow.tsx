import React, { useState, useRef, useEffect } from 'react';
import { List, Input, Button, Avatar, Space, Typography, Tag } from 'antd';
import { UserOutlined, RobotOutlined } from '@ant-design/icons';
import type { Message } from '@/types';

const { Text } = Typography;

interface ChatWindowProps {
  messages: Message[];
  onSend: (content: string) => void;
  sending?: boolean;
  height?: number | string;
}

export const ChatWindow: React.FC<ChatWindowProps> = ({
  messages,
  onSend,
  sending = false,
  height = 500,
}) => {
  const [inputValue, setInputValue] = useState('');
  const scrollRef = useRef<HTMLDivElement>(null);

  const handleSend = () => {
    if (!inputValue.trim() || sending) return;
    onSend(inputValue);
    setInputValue('');
  };

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height }}>
      <div
        ref={scrollRef}
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: '16px',
          backgroundColor: '#f5f5f5',
          borderRadius: '8px',
          marginBottom: '16px',
        }}
      >
        <List
          dataSource={messages}
          renderItem={(item) => (
            <List.Item style={{ border: 'none', padding: '8px 0' }}>
              <div
                style={{
                  display: 'flex',
                  flexDirection: item.role === 'user' ? 'row-reverse' : 'row',
                  width: '100%',
                  alignItems: 'flex-start',
                }}
              >
                <Avatar
                  icon={item.role === 'user' ? <UserOutlined /> : <RobotOutlined />}
                  style={{
                    backgroundColor: item.role === 'user' ? '#1677ff' : '#52c41a',
                    flexShrink: 0,
                  }}
                />
                <div
                  style={{
                    margin: item.role === 'user' ? '0 12px 0 0' : '0 0 0 12px',
                    maxWidth: '70%',
                  }}
                >
                  <div
                    style={{
                      textAlign: item.role === 'user' ? 'right' : 'left',
                      marginBottom: '4px',
                    }}
                  >
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      {item.role === 'user' ? '我' : (item.agentName || '助手')}
                    </Text>
                    {item.agentName && (
                      <Tag style={{ marginLeft: '8px' }}>
                        {item.agentName}
                      </Tag>
                    )}
                  </div>
                  <div
                    style={{
                      backgroundColor: item.role === 'user' ? '#1677ff' : '#fff',
                      color: item.role === 'user' ? '#fff' : 'inherit',
                      padding: '8px 12px',
                      borderRadius: '8px',
                      boxShadow: '0 2px 4px rgba(0,0,0,0.05)',
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-word',
                    }}
                  >
                    {item.content}
                  </div>
                </div>
              </div>
            </List.Item>
          )}
        />
      </div>
      <Space.Compact style={{ width: '100%' }}>
        <Input
          placeholder="输入消息..."
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onPressEnter={handleSend}
          disabled={sending}
        />
        <Button type="primary" onClick={handleSend} loading={sending}>
          发送
        </Button>
      </Space.Compact>
    </div>
  );
};
