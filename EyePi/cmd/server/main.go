package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"google.golang.org/grpc"

	gometrics "github.com/armon/go-metrics"

	"github.com/deciphernow/gm-fabric-go/metrics/gmfabricsink"
	"github.com/deciphernow/gm-fabric-go/metrics/gometricsobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/grpcmetrics"
	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
	ms "github.com/deciphernow/gm-fabric-go/metrics/metricsserver"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"

	"github.com/dborncamp/EyePi/EyePi/cmd/server/config"
	"github.com/dborncamp/EyePi/EyePi/cmd/server/methods"
	pb "github.com/dborncamp/EyePi/EyePi/protobuf"

	// we don't use this directly, but need it in vendor for gateway grpc plugin
	_ "github.com/golang/glog"
	_ "github.com/grpc-ecosystem/grpc-gateway/runtime"
)

func main() {
	var tlsMetricsConf *tls.Config
	var tlsServerConf *tls.Config
	var err error
	var zkCancels []zkCancelFunc

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	logger.Info().Str("service", "EyePi").Msg("starting")

	ctx, cancelFunc := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		for _, f := range zkCancels {
			f()
		}
	}()

	logger.Debug().Str("service", "EyePi").Msg("initializing config")
	if err = config.Initialize(); err != nil {
		logger.Fatal().AnErr("config.Initialize()", err).Msg("")
	}

	logger.Debug().Str("service", "EyePi").Msg("creating server")
	server, err := methods.NewEyePiServer()
	if err != nil {
		logger.Fatal().AnErr("NewEyePiServer())", err).Msg("")
	}

	if tlsMetricsConf, err = buildMetricsTLSConfigIfNeeded(logger); err != nil {
		logger.Fatal().AnErr("buildMetricsTLSConfigIfNeeded", err).Msg("")
	}

	if tlsServerConf, err = buildServerTLSConfigIfNeeded(logger); err != nil {
		logger.Fatal().AnErr("buildServerTLSConfigIfNeeded", err).Msg("")
	}

	ctx = putOauthInCtxIfNeeded(ctx)

	logger.Debug().Str("service", "EyePi").
		Str("host", viper.GetString("grpc_server_host")).
		Int("port", viper.GetInt("grpc_server_port")).
		Msg("creating listener")

	lis, err := net.Listen(
		"tcp",
		fmt.Sprintf(
			"%s:%d",
			viper.GetString("grpc_server_host"),
			viper.GetInt("grpc_server_port"),
		),
	)
	if err != nil {
		logger.Fatal().AnErr("net.Listen", err).Msg("")
	}

	grpcObserver := grpcobserver.New(viper.GetInt("metrics_cache_size"))
	goMetObserver := gometricsobserver.New()
	observers := []subject.Observer{grpcObserver, goMetObserver}

	statsdObserver, err := getStatsdObserverIfNeeded(logger)
	if err != nil {
		logger.Fatal().AnErr("getStatsdObserverIfNeeded", err).Msg("")
	}
	observers = append(observers, statsdObserver...)

	logger.Debug().Str("service", "EyePi").
		Str("host", viper.GetString("metrics_server_host")).
		Int("port", viper.GetInt("metrics_server_port")).
		Msg("starting metrics server")
	err = ms.Start(
		fmt.Sprintf(
			"%s:%d",
			viper.GetString("metrics_server_host"),
			viper.GetInt("metrics_server_port"),
		),
		tlsMetricsConf,
		grpcObserver.Report,
		goMetObserver.Report,
	)
	if err != nil {
		logger.Fatal().AnErr("start metrics server", err).Msg("")
	}

	zkCancels = append(
		zkCancels,
		notifyZkOfMetricsIfNeeded(logger)...,
	)

	metricsChan := subject.New(ctx, observers...)

	sink := gmfabricsink.New(metricsChan)
	gometrics.NewGlobal(gometrics.DefaultConfig("EyePi"), sink)

	opts := []grpc.ServerOption{
		grpc.StatsHandler(grpcmetrics.NewStatsHandler(metricsChan)),
	}

	opts = append(opts, getTLSOptsIfNeeded(tlsServerConf)...)

	oauthOpts, err := getOauthOptsIfNeeded(logger)
	if err != nil {
		logger.Fatal().AnErr("getOauthOptsIfNeeded", err).Msg("")
	}
	opts = append(opts, oauthOpts...)

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterEyePiServer(grpcServer, server)

	logger.Debug().Str("service", "EyePi").
		Msg("starting grpc server")
	go grpcServer.Serve(lis)

	zkCancels = append(
		zkCancels,
		notifyZkOfRPCServerIfNeeded(logger)...,
	)

	if viper.GetBool("use_gateway_proxy") {
		logger.Debug().Str("service", "EyePi").
			Msg("starting gateway proxy")
		if err = startGatewayProxy(ctx, logger); err != nil {
			logger.Fatal().AnErr("startGatewayProxy", err).Msg("")
		}
	}

	zkCancels = append(
		zkCancels,
		notifyZkOfGatewayEndpointIfNeeded(logger)...,
	)

	s := <-sigChan
	logger.Info().Str("service", "EyePi").
		Str("signal", s.String()).
		Msg("shutting down")
	cancelFunc()
	grpcServer.Stop()
}
