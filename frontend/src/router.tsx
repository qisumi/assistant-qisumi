import { createBrowserRouter, Navigate } from 'react-router-dom';
import AppLayout from './components/layout/AppLayout';
import ProtectedRoute from './components/common/ProtectedRoute';
import Login from './pages/Login';
import Tasks from './pages/Tasks';
import TaskDetail from './pages/TaskDetail';
import CreateFromText from './pages/CreateFromText';
import GlobalAssistant from './pages/GlobalAssistant';
import Settings from './pages/Settings';

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <AppLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        element: <Navigate to="/tasks" replace />,
      },
      {
        path: 'tasks',
        element: <Tasks />,
      },
      {
        path: 'tasks/:id',
        element: <TaskDetail />,
      },
      {
        path: 'create-from-text',
        element: <CreateFromText />,
      },
      {
        path: 'global-assistant',
        element: <GlobalAssistant />,
      },
      {
        path: 'settings',
        element: <Settings />,
      },
    ],
  },
]);
