package gocache

/*ByteView 只有一个数据成员，b []byte，b 将会存储真实的缓存值。
  选择 byte 类型是为了能够支持任意的数据类型的存储，例如字符串、图片等。

  实现 Len() int 方法，我们在 lru.Cache 的实现中，要求被缓存对象必须实现 Value 接口，
  即 Len() int 方法，返回其所占的内存大小。

  b 是只读的，使用 ByteSlice() 方法返回一个拷贝，防止缓存值被外部程序修改。
*/
// ByteView 表示一个不可变的字节视图
type ByteView struct {
	b []byte
}

// Len 返回view的大小
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 切片类型返回数据的copy
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String 视图转为string返回
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
