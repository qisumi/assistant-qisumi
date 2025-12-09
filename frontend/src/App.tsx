import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Login from './pages/Login'
import Tasks from './pages/Tasks'
import TaskDetail from './pages/TaskDetail'
import GlobalAssistant from './pages/GlobalAssistant'
import CreateFromText from './pages/CreateFromText'
import Settings from './pages/Settings'

function App() {
  return (
    <Router>
      <div className="App">
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/tasks" element={<Tasks />} />
          <Route path="/tasks/:id" element={<TaskDetail />} />
          <Route path="/global-assistant" element={<GlobalAssistant />} />
          <Route path="/create-from-text" element={<CreateFromText />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="/" element={<Tasks />} />
        </Routes>
      </div>
    </Router>
  )
}

export default App
