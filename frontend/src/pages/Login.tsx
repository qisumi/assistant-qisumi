import React, { useState } from 'react'

const Login: React.FC = () => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLogin, setIsLogin] = useState(true)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    // TODO: Implement login/register logic
    console.log(`${isLogin ? 'Login' : 'Register'} with`, { email, password })
  }

  return (
    <div className="card" style={{ maxWidth: '400px', margin: '0 auto' }}>
      <h2>{isLogin ? '登录' : '注册'}</h2>
      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="email">邮箱</label>
          <input
            type="email"
            id="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        <div>
          <label htmlFor="password">密码</label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        <button type="submit" style={{ width: '100%', marginTop: '10px' }}>
          {isLogin ? '登录' : '注册'}
        </button>
      </form>
      <div style={{ marginTop: '15px', textAlign: 'center' }}>
        <button
          onClick={() => setIsLogin(!isLogin)}
          style={{ background: 'none', color: '#007bff', padding: '0' }}
        >
          {isLogin ? '还没有账号？点击注册' : '已有账号？点击登录'}
        </button>
      </div>
    </div>
  )
}

export default Login
