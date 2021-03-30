package config

import (
	"raccoon/config/util"

	"github.com/spf13/viper"
)

var EventDistribution eventDistribution

type eventDistribution struct {
	PublisherPattern string
}

func eventDistributionConfigLoader() {
	viper.SetDefault("EVENT_DISTRIBUTION_PUBLISHER_PATTERN", "clickstream-%s-log")
	EventDistribution = eventDistribution{
		PublisherPattern: util.MustGetString("EVENT_DISTRIBUTION_PUBLISHER_PATTERN"),
	}
}
