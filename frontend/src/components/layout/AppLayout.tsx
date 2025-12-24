import React from 'react';
import { Layout, Menu, Button, theme } from 'antd';
import {
    UnorderedListOutlined,
    SettingOutlined,
    PlusCircleOutlined,
    MessageOutlined,
    LogoutOutlined,
    UserOutlined,
    CheckCircleOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';

const { Header, Content, Sider } = Layout;

const AppLayout: React.FC = () => {
    const navigate = useNavigate();
    const location = useLocation();
    const { logout, user } = useAuthStore();
    const {
        token: { colorBgContainer, borderRadiusLG },
    } = theme.useToken();

    const menuItems = [
        {
            key: '/tasks',
            icon: <UnorderedListOutlined />,
            label: '任务列表',
        },
        {
            key: '/completed-tasks',
            icon: <CheckCircleOutlined />,
            label: '已完成任务',
        },
        {
            key: '/create-from-text',
            icon: <PlusCircleOutlined />,
            label: '从文本创建',
        },
        {
            key: '/global-assistant',
            icon: <MessageOutlined />,
            label: '小奇（全局）',
        },
        {
            key: '/settings',
            icon: <SettingOutlined />,
            label: '设置',
        },
    ];

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <Sider breakpoint="lg" collapsedWidth="0">
                <div style={{ height: 32, margin: 16, background: 'rgba(255, 255, 255, 0.2)', borderRadius: 6, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', fontWeight: 'bold' }}>
                    Qisumi 助手
                </div>
                <Menu
                    theme="dark"
                    mode="inline"
                    selectedKeys={[location.pathname]}
                    items={menuItems}
                    onClick={({ key }) => navigate(key)}
                />
            </Sider>
            <Layout>
                <Header style={{ padding: '0 24px', background: colorBgContainer, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div />
                    <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
                        <span><UserOutlined /> {user?.display_name || user?.email}</span>
                        <Button type="text" icon={<LogoutOutlined />} onClick={handleLogout}>退出</Button>
                    </div>
                </Header>
                <Content style={{ margin: '24px 16px 0' }}>
                    <div
                        style={{
                            padding: 24,
                            minHeight: 360,
                            background: colorBgContainer,
                            borderRadius: borderRadiusLG,
                        }}
                    >
                        <Outlet />
                    </div>
                </Content>
            </Layout>
        </Layout>
    );
};

export default AppLayout;
