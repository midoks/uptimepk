package log

import (
	"fmt"
	"io"
	"os"
	"strings"

	go_logger "github.com/phachon/go-logger"

	"uptimepk/internal/conf"
)

var (
	logFileName = "uppk.log"
	logger      *go_logger.Logger
)

func Init() {
	logger = go_logger.NewLogger()

	jsonFormat := false
	if strings.EqualFold(conf.Log.Format, "json") {
		jsonFormat = true
	}

	logPath := conf.Log.RootPath

	// creare logs dir
	os.MkdirAll(logPath, 0755)

	fileConfig := &go_logger.FileConfig{
		Filename: fmt.Sprintf("%s/%s", logPath, logFileName),
		LevelFileName: map[int]string{
			logger.LoggerLevel("error"): fmt.Sprintf("%s/%s", logPath, "error.log"),
			logger.LoggerLevel("debug"): fmt.Sprintf("%s/%s", logPath, "debug.log"),
		},
		MaxSize:    1024 * 1024,
		MaxLine:    10000,
		DateSlice:  "d",
		JsonFormat: jsonFormat,
		Format:     "",
	}
	logger.Attach("file", go_logger.LOGGER_LEVEL_DEBUG, fileConfig)
}

func ReverseRead(lineNum uint) ([]string, error) {
	logPath := conf.Log.RootPath
	name := fmt.Sprintf("%s/%s", logPath, logFileName)
	// 打开文件
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// 获取文件大小
	fs, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fs.Size()

	var offset int64 = -1   // 偏移量，初始化为-1，若为0则会读到EOF
	char := make([]byte, 1) // 用于读取单个字节
	lineStr := ""           // 存放一行的数据
	buff := make([]string, 0, 100)
	for (-offset) <= fileSize {
		// 通过Seek函数从末尾移动游标然后每次读取一个字节
		file.Seek(offset, io.SeekEnd)
		_, err := file.Read(char)
		if err != nil {
			return buff, err
		}
		if char[0] == '\n' {
			offset--  // windows跳过'\r'
			lineNum-- // 到此读取完一行
			buff = append(buff, lineStr)
			lineStr = ""
			if lineNum == 0 {
				return buff, nil
			}
		} else {
			lineStr = string(char) + lineStr
		}
		offset--
	}
	buff = append(buff, lineStr)
	return buff, nil
}

func GetLogger() *go_logger.Logger {
	return logger
}

func Debug(args string) {
	logger.Debug(args)
}

func Info(args string) {
	logger.Info(args)
}

func Warn(args string) {
	logger.Warning(args)
}

func Error(args string) {
	logger.Error(args)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}
