package session

import "assistant-qisumi/internal/domain"

// 类型别名 - 引用 domain 包中的定义，避免循环依赖
type Session = domain.Session
type Message = domain.Message
