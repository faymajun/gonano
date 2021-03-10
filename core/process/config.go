package process

import (
	"fmt"
	"github.com/faymajun/gonano/config"
	"github.com/faymajun/gonano/tags"
	"github.com/faymajun/gonano/version"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"github.com/urfave/cli"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func initConfig(ctx *cli.Context) {
	viper.SetConfigType("toml")
	c := ctx.String("config")
	if c == "" {
		panic("configuration file argument missing")
	}

	viper.SetConfigFile(c)

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func loadLogConf() {
	logrus.Infof("当前服务器打包版本=%v", config.Content.String("core_version"))

	// 设置日志相关配置
	lvl, err := logrus.ParseLevel(config.Content.String("log_level"))
	if err != nil {
		lvl = logrus.DebugLevel
	}

	logrus.Infof("日志等级: %s", lvl)
	logrus.SetLevel(lvl)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.00",
	})

	// 通过编译参数控制日志输出
	var outputs []io.Writer
	if !tags.DisableConsole {
		outputs = append(outputs, os.Stderr)
	}
	if directory := config.Content.String("log_directory"); directory != "" {
		abs, err := filepath.Abs(directory)
		if err != nil {
			panic(fmt.Errorf("日志目录配置错误: Error=%s", err.Error()))
		}
		os.MkdirAll(abs, os.ModePerm)
		appname := filepath.Base(os.Args[0])
		//filename := fmt.Sprintf("%s-%s.log", appname, time.Now().Format("2006-01-02.15.04.05"))
		//file, err := os.OpenFile(filepath.Join(abs, filename), os.O_CREATE|os.O_WRONLY, os.ModePerm)
		//if err != nil {
		//	panic(fmt.Errorf("日志文件打开错误, Error=%s", err.Error()))
		//}

		baseLogPath := filepath.Join(abs, appname) + ".logs"
		writer, err := rotatelogs.New(
			baseLogPath+"-%Y-%m-%d.%H.%M",
			rotatelogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
			rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
			rotatelogs.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文件
		)
		if err != nil {
			logrus.Errorf("config local file system logger error. %v", errors.WithStack(err))
		}
		outputs = append(outputs, writer)
	}
	if count := len(outputs); count == 1 {
		logrus.SetOutput(io.MultiWriter(outputs[0]))
	} else if count > 1 {
		logrus.SetOutput(io.MultiWriter(outputs...))
	}

	logrus.Infof("当前服务器打包版本=%v", version.VERSION)

}
