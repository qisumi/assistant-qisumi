import React, { useState } from 'react'

const CreateFromText: React.FC = () => {
  const [text, setText] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    // TODO: Implement create task from text logic
    console.log('Creating task from text:', text)
  }

  return (
    <div>
      <h1>从文本创建任务</h1>
      <div className="card">
        <form onSubmit={handleSubmit}>
          <div>
            <label htmlFor="text">输入文本（会议纪要、聊天记录、备忘录等）</label>
            <textarea
              id="text"
              rows={10}
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="请粘贴或输入文本..."
              required
            ></textarea>
          </div>
          <button type="submit" style={{ marginTop: '10px' }}>
            生成任务
          </button>
        </form>
      </div>
    </div>
  )
}

export default CreateFromText
