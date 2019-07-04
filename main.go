package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/mintance/nginx-clickhouse/clickhouse"
	configParser "github.com/mintance/nginx-clickhouse/config"
	"github.com/mintance/nginx-clickhouse/nginx"
	"github.com/papertrail/go-tail/follower"
	"io"
	"sync"
	"time"
)

var (
	locker sync.Mutex
	logs   []string
)

func main() {


	// Read config & incoming flags
	config := configParser.Read()

	// Update config with environment variables if exist
	config.SetEnvVariables()

	nginxParser, err := nginx.GetParser(config)

	if err != nil {
		logrus.Fatal("Can`t parse nginx log format: ", err)
	}

	logs = []string{}

	t, err := follower.New(config.Settings.LogPath, follower.Config{
		Whence: io.SeekEnd,
		Offset: 0,
		Reopen: true,
	})

	if err != nil {
		logrus.Fatal("Can`t tail logfile: ", err)
	}

	logrus.Info("Opening logfile: " + config.Settings.LogPath)

	go func() {
		for {
			time.Sleep(time.Second * time.Duration(config.Settings.Interval))

			if len(logs) > 0 {

				logrus.Info("Preparing to save ", len(logs), " new log entries.")
				locker.Lock()
				err := clickhouse.Save(config, nginx.ParseLogs(nginxParser, logs))

				if err != nil {
					logrus.Error("Can`t save logs: ", err)
				} else {
					logrus.Info("Saved ", len(logs), " new logs.")
					logs = []string{}
				}

				locker.Unlock()
			}
		}
	}()

	// Push new log entries to array
	for line := range t.Lines() {
		locker.Lock()
		logs = append(logs, line.String())
		locker.Unlock()
	}
}
