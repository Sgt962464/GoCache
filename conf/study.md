## Viper库       
+ 参考资料：https://github.com/spf13/viper?tab=readme-ov-file

```go get github.com/spf13/viper ```导入      
- 支持JSON/TOML/YAML/HCL/ENV/Java properties 等多种格式的配置文件       
- 可以设置监听配置文件的修改，修改时自动加载新配置
- 从环境变量、命令行和io.Reader中读取
- 从远程配置系统中读取和监听修改
- 代码中显式设置键值