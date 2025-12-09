import React from 'react'
import { useParams } from 'react-router-dom'

const TaskDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()

  // TODO: Implement task detail logic
  return (
    <div>
      <h1>任务详情 - {id}</h1>
      <div className="card">
        <p>任务详情将显示在这里</p>
        {/* TODO: Add task detail UI */}
      </div>
      <div className="card" style={{ marginTop: '20px' }}>
        <h3>任务会话</h3>
        {/* TODO: Add chat interface */}
      </div>
    </div>
  )
}

export default TaskDetail
