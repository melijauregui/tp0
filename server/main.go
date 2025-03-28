package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

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

	v.SetEnvPrefix("server")
	v.BindEnv("number_of_agencies")
	v.BindEnv("default.server_port")
	v.BindEnv("default.server_listen_backlog")
	v.BindEnv("default.logging_level")

	v.SetConfigFile("./config.ini")
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Configuration could not be read from config file. Using env variables instead")
	}

	return v, nil
}

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
	log.Infof("action: config | result: success | port: %d | listen_backlog: %d | logging_level: %s | number_of_agencies: %d",
		v.GetInt("default.server_port"),
		v.GetInt("default.server_listen_backlog"),
		v.GetString("default.logging_level"),
		v.GetInt("number_of_agencies"),
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
		Port:             v.GetInt("default.server_port"),
		NumberOfAgencies: v.GetInt("number_of_agencies"),
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	finishChan := make(chan bool)
	server, err := common.NewServer(serverConfig)
	if err != nil {
		log.Fatalf("action: start server | result: success | Failed to start server: %v", err)
	}
	go HandleSignals(server, &wg, finishChan)
	server.Run()
	if server.IsRunning() {
		finishChan <- true
	}
	close(finishChan)
	log.Infof("action: finish server | result: success ")
	wg.Wait()
	time.Sleep(1000 * time.Millisecond)
}

func HandleSignals(s *common.Server, wg *sync.WaitGroup, finishChan chan bool) {

	defer wg.Done()

	sigChannel := make(chan os.Signal, 1) // espera las signals
	//crea un canal (chan) en Go que puede recibir valores del tipo os.Signal
	//el 1 en make(chan os.Signal, 1) significa que es un canal con buffer de tamaño 1
	signal.Notify(sigChannel, syscall.SIGTERM)
	//escuche la señal SIGTERM del sistema operativo.
	//cuando SIGTERM ocurra, se enviará automáticamente al canal sigChannel
	select {
	case <-finishChan:
		log.Infof("action: signal | result: success | signal: finish")
	case <-sigChannel:
		log.Infof("action: signal | result: success | signal: SIGTERM")
		//Bloquea la ejecución hasta que el canal reciba una señal.
		s.GracefulShutdown()
	}
}
