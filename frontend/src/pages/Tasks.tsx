import React from 'react'

const Tasks: React.FC = () => {
  // TODO: Implement task list logic
  return (
    <div>
      <div className="flex justify-between align-center" style={{ marginBottom: '20px' }}>
        <h1>任务列表</h1>
        <div className="flex gap-10">
          <button>创建任务</button>
          <button>从文本创建</button>
        </div>
      </div>
      <div className="card">
        <p>任务列表将显示在这里</p>
        {/* TODO: Add task list UI */}
      </div>
    </div>
  )
}

export default Tasks
