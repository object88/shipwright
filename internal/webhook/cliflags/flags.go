package cliflags

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	secretKey string = "github-webhook-secret"
)

type FlagManager struct {
	secret string
}

func New() *FlagManager {
	return &FlagManager{}
}

func (fl *FlagManager) ConfigureSecret(flags *pflag.FlagSet) {
	flags.StringVar(&fl.secret, secretKey, "", "used to confirm incoming webhook requests'")
	viper.BindEnv(secretKey)
	viper.BindPFlag(secretKey, flags.Lookup(secretKey))
}

func (fl *FlagManager) Secret() string {
	return viper.GetString(secretKey)
}
