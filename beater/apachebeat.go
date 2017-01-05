package beater

import (
	"fmt"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/singlehopllc/apachebeat/config"
	"net/url"
	"time"
)

const selector = "apachebeat"
const AUTO_STRING = "?auto"

type Apachebeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
	urls   []*url.URL
	period time.Duration

	auth     bool
	username string
	password string
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Apachebeat{
		done:   make(chan struct{}),
		config: config,
	}

	//define default URL if none provided
	var urlConfig []string
	if config.URLs != nil {
		urlConfig = config.URLs
	} else {
		urlConfig = []string{"http://127.0.0.1/server-status"}
	}

	bt.urls = make([]*url.URL, len(urlConfig))
	for i := 0; i < len(urlConfig); i++ {
		u, err := url.Parse(urlConfig[i])
		if err != nil {
			logp.Err("Invalid Apache HTTPD server status page: %v", err)
			return nil, err
		}
		bt.urls[i] = u
	}

	logp.Info(config.Period.String())
	bt.period = time.Duration(config.Period) * time.Second

	if config.Username == "" || config.Password == "" {
		logp.Info("Username or password is not set.")
		bt.auth = false
	} else {
		bt.username = config.Username
		bt.password = config.Password
		bt.auth = true
	}

	return bt, nil
}

func (bt *Apachebeat) Run(b *beat.Beat) error {
	logp.Info("apachebeat is running! Hit CTRL-C to stop it.")
	for i, u := range bt.urls {
		logp.Debug(selector, "Iteration: %d", i)
		logp.Debug(selector, "URL: %s", u.String())
		go func(u *url.URL) {

			bt.client = b.Publisher.Connect()
			ticker := time.NewTicker(bt.config.Period)
			for {
				select {
				case <-bt.done:
					goto GotoFinish
				case <-ticker.C:
				}
				timerStart := time.Now()
				logp.Info(selector, "Cluster stats for url: %v", u)
				serverStatus, err := bt.GetServerStatus(*u)
				if err != nil {
					logp.Err("Error getting server-status for %s: %v", u.String(), err.Error())
				} else {
					event := common.MapStr{
						"@timestamp": common.Time(time.Now()),
						"type":       "apache_status", //TODO: NAMING??
						"url":        u.String(),
						"apache":     serverStatus, //TODO: NAMING??
					}
					bt.client.PublishEvent(event)
					logp.Info("Event sent")
				}

				timerEnd := time.Now()
				duration := timerEnd.Sub(timerStart)
				if duration.Nanoseconds() > bt.period.Nanoseconds() {
					logp.Warn("Ignoring tick(s) due to processing taking longer than one period")
				}
			}
		GotoFinish:
		}(u)
	}
	<-bt.done
	return nil
}

func (bt *Apachebeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
