package service

/*
Picker 负责查找密钥的查询请求应发送到哪个节点。（使用一致的哈希算法）
*/
type Picker interface {
	Pick(key string) (Fetcher, bool)
}

/*
Fetcher 负责查询指定组缓存中键的值。
每个分布式kv节点都应该实现这个接口。
*/
type Fetcher interface {
	Fetch(group string, key string) ([]byte, error)
}

/*
Retriever 用于从后端数据库检索数据的检索器接口。
当不能从节点的组高速缓存中查询密钥的值时，
系统需要提供替代选项，即转到后端数据库查询密钥的值。
*/
type Retriever interface {
	retrieve(string) ([]byte, error)
}

/*
通过向 RetrieveFunc 添加一个方法 retrieve，可以将此函数类型的实例用作 Retriever 接口的实现；
这是一种经典的适配器模式，将函数适配为接口。
不需要一个完整的结构来实现接口，只需要一个功能来满足接口的要求，这是go中常见的简化技术，
使代码更加简洁易懂；这种模式允许快速定制以更改或注入数据检索策略，特别是对于需要高度灵活性和动态数据处理的场景。
例如，如果正在构建一个需要从多个数据源检索数据的微服务架构，则可以在不影响其他业务逻辑的情况下轻松切换不同的数据检索策略。
*/
type RetrieveFunc func(key string) ([]byte, error)

/*
RetrieveFunc 实现了 retrieve 方法，即实现 Retriever 接口，使任何匿名函数func通过RetrieverFunc（func）强制类型转换，实现 RetrieverFunc 接口的能力。
这也反映在gin框架内部的HandlerFunc类型封装匿名函数中，http类型的处理程序强制转换可以直接用作gin处理程序。
*/
func (f RetrieveFunc) retrieve(key string) ([]byte, error) {
	return f(key)
}
