package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/common"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

var log = logging.MustGetLogger("log")

// Config holds the configuration parameters.
type Config struct {
	Port          int
	ListenBacklog int
	LoggingLevel  string
}

func InitConfig() (*viper.Viper, error) {

	v := viper.New()
	v.AutomaticEnv()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.BindEnv("default.server_port")
	v.BindEnv("default.server_listen_backlog")
	v.BindEnv("default.logging_level")

	v.SetConfigFile("./config.ini")
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Configuration could not be read from config file. Using env variables instead")
	}

	return v, nil
}

// InitLogger sets up the logging format.
// Note: The standard log package does not support logging levels by default.
func InitLogger(logLevel string) error {
	baseBackend := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05} %{level:.5s}     %{message}`,
	)

	backendFormatter := logging.NewBackendFormatter(baseBackend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	logLevelCode, err := logging.LogLevel(logLevel)
	if err != nil {
		return err
	}
	backendLeveled.SetLevel(logLevelCode, "")
	logging.SetBackend(backendLeveled)
	return nil
}

func PrintConfig(v *viper.Viper) {
	log.Infof("action: config | result: success | port: %d | listen_backlog: %d | logging_level: %s",
		v.GetInt("default.server_port"),
		v.GetInt("default.server_listen_backlog"),
		v.GetString("default.logging_level"),
	)
}

func main() {
	v, err := InitConfig()
	if err != nil {
		log.Criticalf("%s", err)
	}

	if err := InitLogger("INFO"); err != nil {
		log.Criticalf("%s", err)
	}

	PrintConfig(v)

	serverConfig := common.ServerConfig{
		Port: v.GetInt("default.server_port"),
	}

	// Initialize server and start the server loop.
	// Note: The Server implementation is assumed to be defined in the same package.
	server, err := common.NewServer(serverConfig)
	if err != nil {
		log.Fatalf("action: start server | result: success | Failed to start server: %v", err)
	}
	go HandleSignals(server)

	server.Run()
}

func HandleSignals(s *common.Server) {
	sigChannel := make(chan os.Signal, 1) // espera las signals
	//crea un canal (chan) en Go que puede recibir valores del tipo os.Signal
	//el 1 en make(chan os.Signal, 1) significa que es un canal con buffer de tamaño 1
	signal.Notify(sigChannel, syscall.SIGTERM)
	//escuche la señal SIGTERM del sistema operativo.
	//cuando SIGTERM ocurra, se enviará automáticamente al canal sigChannel
	<-sigChannel
	//Bloquea la ejecución hasta que el canal reciba una señal.

	s.GracefulShutdown()
	os.Exit(0)
}
