package logx

import (
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"path"
	"time"
)

const LogPath = "./logs"
const FileSuffix = ".log"

func Init() {
	if err := os.MkdirAll(LogPath, os.ModePerm); err != nil {
		panic(err)
	}

	logrus.SetLevel(logrus.TraceLevel)
	//logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02.15:04:05.000000",
		ForceColors:     true,
		DisableColors:   false,
	})
	logrus.SetOutput(os.Stdout)
	// set logx file
	writeFile(LogPath, "console", 8)

	// std err redirect to file
	stdErrPath := path.Join(LogPath, "error") + FileSuffix
	f, err := os.OpenFile(stdErrPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	redirectStderr(f)
}

func writeFile(logPath string, name string, save uint) {
	var cstSh, _ = time.LoadLocation("Asia/Shanghai") //上海
	fileSuffix := time.Now().In(cstSh).Format("2006-01-02") + FileSuffix
	logFullPath := path.Join(logPath, name)
	logFullName := logFullPath + fileSuffix
	fileWriter, err := rotatelogs.New(
		logFullName,
		//rotatelogs.WithLinkName(logFullPath+FileSuffix), // 生成软链，指向最新日志文件
		rotatelogs.WithRotationCount(int(save)),   // 文件最大保存份数
		rotatelogs.WithRotationTime(time.Hour*24), // 日志切割时间间隔
	)

	if err != nil {
		panic(err)
	}

	fileFormatter := &prefixed.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02.15:04:05.000000",
		ForceFormatting: true,
		ForceColors:     false,
		DisableColors:   true,
	}
	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.PanicLevel: fileWriter,
		logrus.FatalLevel: fileWriter,
		logrus.ErrorLevel: fileWriter,
		logrus.WarnLevel:  fileWriter,
		logrus.InfoLevel:  fileWriter,
	}, fileFormatter)

	logrus.AddHook(lfHook)
}
