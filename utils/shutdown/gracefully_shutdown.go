package shutdown

import (
	"context"
	"gocache/utils/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

/*
GracefullyShutdown  优雅关闭服务器

  - 创建一个缓冲通道，用于接受操作系统信号，大小为1
  - 注册信号：
    --os.Interrupt   :Ctrl+C
    --syscall.SIGINT :终端的中断信号
    --syscall.SIGTERM:请求程序的终止信号
  - <-done 等待信号的到来
  - 调用util日志工具，记录日志
  - server.Shutdown(context.Background()) 优雅关闭
    --会等待所有活动连接完成处理再关闭server
*/
func GracefullyShutdown(server *http.Server) {
	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done

	logger.LogrusObj.Println("closing http server gracefully... ")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.LogrusObj.Fatalln("closing http server gracefully failed: ", err)
	}
}
