package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/hubaxis/trader/protocol"
	"github.com/leshachaplin/OpenPriceTool/model"
	"github.com/leshachaplin/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	s := make(chan os.Signal)
	done, cnsl := context.WithCancel(context.Background())

	username := "lesha"
	symbol := "USDUAH"
	stoppLoss := float64(10000)

	opts := grpc.WithInsecure()
	clientConnInterface, err := grpc.Dial("0.0.0.0:50051", opts)
	if err != nil {
		log.Error(err)
	}
	defer clientConnInterface.Close()

	client := protocol.NewTraderServiceClient(clientConnInterface)

	// Connect to aws-ssm
	awsConf, err := config.NewForAws("us-west-2")
	if err != nil {
		log.Error("Can't connect to aws: ", err)
	}

	// get redis connection url from aws parameter store
	redisConfig, err := awsConf.GetRedis("aws-ssm-redis://Redis/")
	if err != nil {
		log.Error("Can't get redis config from aws: ", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisConfig.Host + ":" + strconv.Itoa(redisConfig.Port),
		DB:   0,
	})

	price := &model.Price{}
	key := fmt.Sprintf("%s_last", symbol)
	result, err := redisClient.Get(key).Bytes()
	if err != nil {
		log.Error("error in get from redis last id ", err)
	}

	price, err = price.UnmarshalBinary(result)
	if err != nil {
		log.Error("error in unmarshalling binary in get last id's ", err)
	}

	req := &protocol.ReadPriceRequest{
		Symbols: []string{
			symbol,
			price.ID},
	}

	stream, err := client.ReadPrice(context.Background(), req)
	if err != nil {
		log.Error(err)
	}

	repo := make(map[string]model.Price, 0)

	go func(ctx context.Context, r map[string]model.Price, str protocol.TraderService_ReadPriceClient) {
		for {
			select {
			case <-ctx.Done():
				{
					err := stream.CloseSend()
					if err != nil {
						log.Error(err)
					}
				}
			default:
				{
					in, err := str.Recv()
					if err == io.EOF {
						continue
					}
					if err != nil {
						log.Error(err)
						//TODO Reconnect
					}

					res := model.Price{
						Bid:      in.Price.Bid,
						Ask:      in.Price.Ack,
						Date:     time.Unix(in.Price.Date, 0),
						Symbol:   in.Price.Symbol,
						Currency: in.Price.Currency,
					}

					r[res.Symbol] = res
				}
			}
		}
	}(done, repo, stream)

	val := repo[symbol]

	stopLoss := &protocol.StopLossValue{
		Value:    stoppLoss,
		IsEnable: true,
	}

	posReq := &protocol.OpenPositionRequest{
		Username: username,
		Symbol:   symbol,
		Short:    false,
		Amount:   10,
		PriceId:  val.ID,
		Value:    stopLoss,
	}

	_, err = client.OpenPosition(context.Background(), posReq)
	if err != nil {
		log.Error(err)
	}

	getOpenReq := &protocol.GetOpenPositionsRequest{
		Username: posReq.Username,
	}
	_, err = client.GetOpenPositions(done, getOpenReq)
	if err != nil {
		log.Error("Can't get data from postgres", err)
	}

	c := make(chan os.Signal, 0)
	signal.Notify(c, os.Interrupt)

	<-s
	close(s)
	cnsl()

	<-c
	cnsl()

	if err := redisClient.Close(); err != nil {
		log.Errorf("redis not closed %s", err)
	}

	log.Info("Cancel is successful")
	close(c)
	return
}
