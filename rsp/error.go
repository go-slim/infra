package rsp

type Fundamental interface {
	// Status 返回一个 HTTP 状态码
	//
	// 如果针对业务逻辑，一般情况下，HTTP 状态码请使用下面给出的：
	//
	//   - 400：用于参数校验失败或查询出错；
	//   - 500：操作数据失败，比如创建、删除、更新等操作。
	//
	// 若遇到特殊情况，可使用其它表示错误的 HTTP 状态码。
	Status() int

	// Code 返回错误代码
	//
	// 参考 Google 错误码规范实现：
	// https://cloud.google.com/resource-manager/docs/core_errors?hl=zh-cn
	//
	// 错误码的命名规则：
	//
	//   - 使用驼峰命名法；
	//   - 必须由有意义的单词组成；
	//   - 单词首字母大写，其余字母小写；
	//   - 错误码必须是唯一的。
	//
	// 例如：RecordNotFound 表示查询的数据不存在。
	Code() string

	// Text 返回错误提示
	//
	// 描述错误的原因和解决方法，必须是人类友好的文本，对于非技术架构错误，切勿包含任何技术信息：
	//
	//   - 错误示例："数据库连接失败"，这是一个技术信息，不适合作为响应提示消息。
	//   - 正确示例："查询的数据不存在"，这是一个人类可读的文本，适合作为响应提示消息。
	Text() string

	// Data 返回携带的响应数据
	Data() any

	// Cause 返回原始错误对象
	Cause() error
}
