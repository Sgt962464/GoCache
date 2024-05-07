package logger

import (
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

var (
	LogrusObj *logrus.Logger
	once      sync.Once
)

func init() {
	if LogrusObj != nil {
		outputFile, _ := setOutputFile()
		LogrusObj.Out = outputFile
		return
	}

	logger := logrus.New()
	// outputFile, _ := setOutputFile()
	// logger.Out = outputFile
	logger.Out = os.Stdout
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	LogrusObj = logger
}

func setOutputFile() (*os.File, error) {
	//获取时间用于生产日志文件名
	now := time.Now()
	logFilePath := ""
	//获取当前文件目录
	if dir, err := os.Getwd(); err == nil {
		logFilePath = dir + "/logs/"
	}
	_, err := os.Stat(logFilePath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(logFilePath, 0777); err != nil {
			log.Println(err.Error())
			return nil, err
		}
	}
	//格式化日志文件名
	logFileName := now.Format("2006-01-02") + ".log"
	fileName := path.Join(logFilePath, logFileName)
	//检查是否存在日志文件
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			log.Println(err.Error())
			return nil, err
		}
	}
	// 使用 os.OpenFile 打开日志文件，以追加模式（os.O_APPEND）和只写模式（os.O_WRONLY）打开它。
	// os.ModeAppend 只能追加信息
	outputFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return outputFile, nil

}
