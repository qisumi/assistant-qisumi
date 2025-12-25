package task

import "assistant-qisumi/internal/domain"

// 类型别名 - 引用 domain 包中的定义，避免循环依赖
type Task = domain.Task
type TaskStep = domain.TaskStep
type TaskDependency = domain.TaskDependency
type UpdateTaskFields = domain.UpdateTaskFields
type UpdateStepFields = domain.UpdateStepFields
type NewStepRecord = domain.NewStepRecord
type DependencyItem = domain.DependencyItem
type FlexibleTime = domain.FlexibleTime

// 导出 domain 包的时间解析函数
var ParseFlexibleTime = domain.ParseFlexibleTime
var ParseRFC3339 = domain.ParseRFC3339
