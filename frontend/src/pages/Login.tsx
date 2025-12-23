import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, message, Tabs } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { login, register } from '../api/auth';

const { Title } = Typography;

const Login: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const [activeTab, setActiveTab] = useState('login');
    const navigate = useNavigate();
    const setAuth = useAuthStore((state) => state.setAuth);

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
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5' }}>
            <Card style={{ width: 400, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}>
                <div style={{ textAlign: 'center', marginBottom: 24 }}>
                    <Title level={2}>Qisumi Assistant</Title>
                    <Typography.Text type="secondary">任务规划 & 备忘录系统</Typography.Text>
                </div>

                <Tabs
                    activeKey={activeTab}
                    onChange={setActiveTab}
                    centered
                    items={[
                        { key: 'login', label: '登录' },
                        { key: 'register', label: '注册' },
                    ]}
                />

                <Form
                    name="auth_form"
                    initialValues={{ remember: true }}
                    onFinish={onFinish}
                    layout="vertical"
                    style={{ marginTop: 24 }}
                >
                    <Form.Item
                        name="email"
                        rules={[{ required: true, message: '请输入邮箱!', type: 'email' }]}
                    >
                        <Input prefix={<UserOutlined />} placeholder="邮箱" size="large" />
                    </Form.Item>
                    <Form.Item
                        name="password"
                        rules={[{ required: true, message: '请输入密码!' }]}
                    >
                        <Input.Password prefix={<LockOutlined />} placeholder="密码" size="large" />
                    </Form.Item>

                    <Form.Item>
                        <Button type="primary" htmlType="submit" loading={loading} block size="large">
                            {activeTab === 'login' ? '登录' : '注册'}
                        </Button>
                    </Form.Item>
                </Form>
            </Card>
        </div>
    );
};

export default Login;
