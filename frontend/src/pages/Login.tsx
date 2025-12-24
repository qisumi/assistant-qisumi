import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, Tabs, Spin, App } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useAuthStore } from '../store/authStore';
import { login, register } from '../api/auth';

const { Title, Text } = Typography;

const Login: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const [activeTab, setActiveTab] = useState('login');
    const navigate = useNavigate();
    const setAuth = useAuthStore((state) => state.setAuth);
    const { message } = App.useApp();

    const onFinish = async (values: any) => {
        setLoading(true);
        try {
            let res;
            if (activeTab === 'login') {
                res = await login(values.email, values.password);
                message.success('登录成功');
            } else {
                res = await register(values.email, values.password);
                message.success('注册成功');
            }
            setAuth(res.token, res.user);
            navigate('/tasks');
        } catch (error: any) {
            message.error(error.response?.data?.error || '操作失败');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            minHeight: '100vh',
            background: 'linear-gradient(135deg, #5E6AD2 0%, #722ed1 100%)',
            padding: 24
        }}>
            <motion.div
                initial={{ opacity: 0, y: 30, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                transition={{
                    type: 'spring',
                    stiffness: 300,
                    damping: 20,
                }}
            >
                <Card
                    style={{
                        width: '100%',
                        maxWidth: 420,
                        boxShadow: '0 20px 40px rgba(0, 0, 0, 0.15)',
                        borderRadius: '16px',
                        position: 'relative',
                        overflow: 'hidden'
                    }}
                    styles={{ body: { padding: '40px 32px' } }}
                >
                    {/* Loading overlay */}
                    {loading && (
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            style={{
                                position: 'absolute',
                                top: 0,
                                left: 0,
                                right: 0,
                                bottom: 0,
                                background: 'rgba(255, 255, 255, 0.9)',
                                display: 'flex',
                                justifyContent: 'center',
                                alignItems: 'center',
                                zIndex: 10,
                                borderRadius: '16px'
                            }}
                        >
                            <Spin size="large" />
                        </motion.div>
                    )}

                    {/* Header */}
                    <div style={{ textAlign: 'center', marginBottom: 32 }}>
                        <Title level={2} style={{ marginBottom: 8, color: '#18181b' }}>
                            小奇
                        </Title>
                        <Text type="secondary" style={{ fontSize: '14px' }}>
                            智能任务规划 & 备忘录系统
                        </Text>
                    </div>

                    {/* Tabs */}
                    <Tabs
                        activeKey={activeTab}
                        onChange={setActiveTab}
                        centered
                        items={[
                            { key: 'login', label: '登录' },
                            { key: 'register', label: '注册' },
                        ]}
                        style={{ marginBottom: 24 }}
                    />

                    {/* Form */}
                    <Form
                        name="auth_form"
                        initialValues={{ remember: true }}
                        onFinish={onFinish}
                        layout="vertical"
                    >
                        <Form.Item
                            name="email"
                            rules={[{ required: true, message: '请输入邮箱!', type: 'email' }]}
                        >
                            <Input
                                prefix={<UserOutlined />}
                                placeholder="邮箱"
                                size="large"
                                style={{ borderRadius: '8px' }}
                            />
                        </Form.Item>

                        <Form.Item
                            name="password"
                            rules={[{ required: true, message: '请输入密码!' }]}
                        >
                            <Input.Password
                                prefix={<LockOutlined />}
                                placeholder="密码"
                                size="large"
                                style={{ borderRadius: '8px' }}
                            />
                        </Form.Item>

                        <Form.Item style={{ marginBottom: 0 }}>
                            <Button
                                type="primary"
                                htmlType="submit"
                                loading={loading}
                                block
                                size="large"
                                style={{
                                    borderRadius: '8px',
                                    height: '44px',
                                    fontSize: '16px',
                                    fontWeight: 500
                                }}
                            >
                                {activeTab === 'login' ? '登录' : '注册'}
                            </Button>
                        </Form.Item>
                    </Form>

                    {/* Footer note */}
                    <div style={{ textAlign: 'center', marginTop: 24 }}>
                        <Text type="secondary" style={{ fontSize: '12px' }}>
                            使用 Qisumi AI 管理您的任务
                        </Text>
                    </div>
                </Card>
            </motion.div>
        </div>
    );
};

export default Login;
