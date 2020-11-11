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

func ReadPrice(ctx context.Context, price *model.Price, c chan model.Price,
	client protocol.TraderServiceClient) {
	go func(ctx context.Context, pr *model.Price,
		priceChannel chan model.Price, cli protocol.TraderServiceClient) {

		req := &protocol.ReadPriceRequest{
			Symbols: []string{
				price.Symbol,
				price.ID},
		}

		stream, err := cli.ReadPrice(context.Background(), req)
		if err != nil {
			log.Error(err)
		}

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
					in, err := stream.Recv()
					if err == io.EOF {
						continue
					}
					if err != nil {
						log.Error(err)
						//TODO Reconnect
					}

					res := model.Price{
						ID:       in.Price.PriceId,
						Bid:      in.Price.Bid,
						Ask:      in.Price.Ack,
						Date:     time.Unix(in.Price.Date, 0),
						Symbol:   in.Price.Symbol,
						Currency: in.Price.Currency,
					}

					priceChannel <- res
				}
			}
		}

	}(ctx, price, c, client)
}

func OpenPosition(ctx context.Context,
	redisCli redis.UniversalClient,
	client protocol.TraderServiceClient,
	isShort bool,
	symbol, username string) {
	log.Info("OPEN SYMBOL: ", symbol)
	price := &model.Price{}
	key := fmt.Sprintf("%s_last", symbol)
	result, err := redisCli.Get(key).Bytes()
	if err != nil {
		log.Error("error in get from redis last id ", err)
	}

	price, err = price.UnmarshalBinary(result)
	if err != nil {
		log.Error("error in unmarshalling binary in get last id's ", err)
	}

	ch := make(chan model.Price)

	ReadPrice(ctx, price, ch, client)

	val := <-ch

	takeProfit := &protocol.TakeProfitValue{
		Value:    0,
		IsEnable: false,
	}

	stopLoss := &protocol.StopLossValue{
		Value:    0,
		IsEnable: false,
	}

	posReq := &protocol.OpenPositionRequest{
		Username:   username,
		Symbol:     symbol,
		Short:      isShort,
		Amount:     10,
		PriceId:    val.ID,
		Value:      stopLoss,
		TakeProfit: takeProfit,
	}

	_, err = client.OpenPosition(context.Background(), posReq)
	if err != nil {
		log.Error(err)
	}
}

func main() {
	done, cnsl := context.WithCancel(context.Background())

	username := "lesha"

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

	c := make(chan os.Signal, 0)
	signal.Notify(c, os.Interrupt)

	symbol := fmt.Sprintf("EURUSD")
	OpenPosition(done, redisClient, client, false, symbol, username)

	symbol1 := fmt.Sprintf("EURUSD0")
	OpenPosition(done, redisClient, client, false, symbol1, username)

	symbol2 := fmt.Sprintf("EURUSD1")
	OpenPosition(done, redisClient, client, false, symbol2, username)

	symbol3 := fmt.Sprintf("EURUSD2")
	OpenPosition(done, redisClient, client, true, symbol3, username)

	symbol4 := fmt.Sprintf("EURUSD3")
	OpenPosition(done, redisClient, client, true, symbol4, username)

	symbol5 := fmt.Sprintf("EURUSD4")
	OpenPosition(done, redisClient, client, true, symbol5, username)

	<-c
	cnsl()

	if err := redisClient.Close(); err != nil {
		log.Errorf("redis not closed %s", err)
	}

	log.Info("Cancel is successful")
	close(c)
	return
}
