package hin

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func Run(r *gin.Engine, addr string) {

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

const (
	RecommendedEnvPrefix = "HIN"

	configFlagName = "configuration"
)

var cfgFile = pflag.StringP(configFlagName, "c", "", "Read configuration from specified `FILE`, "+
	"support JSON, TOML, YAML, HCL, or Java properties formats.")

func LoadConfig(defaultName string) {
	pflag.Parse()

	if cfgFile != nil {
		viper.SetConfigFile(*cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName(defaultName)
	}

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix(RecommendedEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("WARNING: viper failed to discover and load the configuration file: %s\n", err.Error())
	}
}
