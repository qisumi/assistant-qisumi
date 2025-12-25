import React, { useState } from 'react';
import { Layout, Menu, Button, Drawer, Avatar, Space, Tooltip } from 'antd';
import {
    UnorderedListOutlined,
    SettingOutlined,
    PlusCircleOutlined,
    MessageOutlined,
    LogoutOutlined,
    UserOutlined,
    CheckCircleOutlined,
    MenuOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';
import { useResponsive } from '@/hooks';

const { Header, Content, Sider } = Layout;

const AppLayout: React.FC = () => {
    const navigate = useNavigate();
    const location = useLocation();
    const { logout, user } = useAuthStore();
    const [mobileDrawerOpen, setMobileDrawerOpen] = useState(false);
    const { isMobile, isTablet, isDesktop } = useResponsive();

    // Responsive sizing
    const contentMargin = isMobile ? 12 : isTablet ? 16 : 20;
    const contentPadding = isMobile ? 16 : 20;
    const headerPadding = isMobile ? '0 16px' : '0 24px';
    const siderWidth = isMobile ? 220 : 250;
    const headerHeight = isMobile ? 56 : 64;

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

    const handleMenuClick = ({ key }: { key: string }) => {
        navigate(key);
        setMobileDrawerOpen(false); // Close drawer on mobile after navigation
    };

    // Sidebar content (reused for desktop Sider and mobile Drawer)
    const sidebarContent = (
        <>
            <div style={{
                height: isMobile ? 28 : 32,
                margin: isMobile ? 12 : 16,
                background: 'linear-gradient(135deg, #5E6AD2 0%, #722ed1 100%)',
                borderRadius: 6,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'white',
                fontWeight: 600,
                fontSize: isMobile ? '13px' : '14px'
            }}>
                小奇
            </div>
            <Menu
                mode="inline"
                selectedKeys={[location.pathname]}
                items={menuItems}
                onClick={handleMenuClick}
                style={{ borderRight: 0 }}
            />
        </>
    );

    return (
        <Layout style={{ height: '100vh', overflow: 'hidden' }}>
            {/* Desktop Sidebar - hidden on mobile */}
            {isDesktop ? (
                <Sider
                    breakpoint="lg"
                    collapsedWidth="0"
                    style={{ background: '#ffffff', borderRight: '1px solid #e8e8eb', height: '100vh' }}
                    width={siderWidth}
                >
                    {sidebarContent}
                </Sider>
            ) : null}

            {/* Mobile Drawer */}
            <Drawer
                placement="left"
                open={mobileDrawerOpen}
                onClose={() => setMobileDrawerOpen(false)}
                styles={{
                    body: { padding: 0 },
                    header: { display: 'none' }
                }}
                width={siderWidth}
            >
                {sidebarContent}
            </Drawer>

            {/* Main Layout */}
            <Layout>
                {/* Header */}
                <Header style={{
                    padding: headerPadding,
                    background: '#ffffff',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    borderBottom: '1px solid #e8e8eb',
                    height: headerHeight
                }}>
                    {/* Mobile menu trigger */}
                    {!isDesktop ? (
                        <Button
                            type="text"
                            icon={<MenuOutlined />}
                            onClick={() => setMobileDrawerOpen(true)}
                        />
                    ) : (
                        <div />
                    )}

                    {/* User info and logout */}
                    <Space size={isMobile ? 'small' : 'middle'}>
                        <Tooltip title={user?.display_name || user?.email}>
                            <Avatar
                                size={isMobile ? 'small' : 'default'}
                                icon={<UserOutlined />}
                                style={{
                                    backgroundColor: '#5E6AD2',
                                    cursor: 'pointer'
                                }}
                            />
                        </Tooltip>
                        <Button
                            type="text"
                            icon={<LogoutOutlined />}
                            onClick={handleLogout}
                            style={{ color: '#6b6b76' }}
                        >
                            {!isMobile && '退出'}
                        </Button>
                    </Space>
                </Header>

                {/* Content Area */}
                <Content style={{
                    margin: contentMargin,
                    background: '#f7f7f8',
                    height: `calc(100vh - ${headerHeight + contentMargin * 2}px)`,
                    overflow: 'auto'
                }}>
                    <div
                        style={{
                            padding: contentPadding,
                            minHeight: '100%',
                            background: '#ffffff',
                            borderRadius: isMobile ? 8 : 12,
                            boxShadow: '0 1px 2px rgba(0, 0, 0, 0.05)'
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
