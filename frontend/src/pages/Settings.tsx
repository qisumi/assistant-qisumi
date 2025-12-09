import React, { useState } from 'react'

const Settings: React.FC = () => {
  // TODO: Implement settings logic
  const [displayName, setDisplayName] = useState('')
  const [baseUrl, setBaseUrl] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [model, setModel] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    // TODO: Implement save settings logic
    console.log('Saving settings:', { displayName, baseUrl, apiKey, model })
  }

  return (
    <div>
      <h1>设置</h1>
      <div className="card">
        <h3>用户资料</h3>
        <form onSubmit={handleSubmit}>
          <div>
            <label htmlFor="displayName">显示名称</label>
            <input
              type="text"
              id="displayName"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="请输入显示名称"
            />
          </div>
          <h3 style={{ marginTop: '20px' }}>LLM 配置</h3>
          <div>
            <label htmlFor="baseUrl">API Base URL</label>
            <input
              type="text"
              id="baseUrl"
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
              placeholder="https://api.openai.com/v1"
              required
            />
          </div>
          <div>
            <label htmlFor="apiKey">API Key</label>
            <input
              type="password"
              id="apiKey"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="sk-..."
              required
            />
          </div>
          <div>
            <label htmlFor="model">默认模型</label>
            <input
              type="text"
              id="model"
              value={model}
              onChange={(e) => setModel(e.target.value)}
              placeholder="gpt-4"
              required
            />
          </div>
          <button type="submit" style={{ marginTop: '20px' }}>
            保存设置
          </button>
        </form>
      </div>
    </div>
  )
}

export default Settings
