package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/adshao/go-binance/v2"
)

var config Config = LoadConfiguration("config.json")
var client = binance.NewClient(config.Binance.BinanceAPIKey, config.Binance.BinanceSecretKey)

func main() {
	anaconda.SetConsumerKey(config.Twitter.TwitterConsumerKey)
	anaconda.SetConsumerSecret(config.Twitter.TwitterConsumerSecret)
	api := anaconda.NewTwitterApi(config.Twitter.TwitterAccessToken, config.Twitter.TwitterAccessSecret)
	v := url.Values{}
	s := api.PublicStreamFilter(v)
	v.Add("follow", config.Twitter.Elonid)
	//fmt.Println(v.Encode())
	fmt.Println("Signal Started")
	for t := range s.C {
		switch v := t.(type) {
		case anaconda.Tweet:
			if strconv.Itoa(int(v.User.Id)) == config.Twitter.Elonid {
				fmt.Println(v.User.ScreenName, v.Text)
				fmt.Println("Signal Alert!")
				Buy()
			}
		}
	}
}

func Buy() {
	order, err := client.NewCreateOrderService().Symbol(config.BaseAsset + config.QuoteAsset).
		Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).QuoteOrderQty(fmt.Sprintf("%f", config.Quantity)).Do(context.Background())
	if err != nil {
		fmt.Println(err)
	} else {
		totalQuantity := order.ExecutedQuantity
		if s, err := strconv.ParseFloat(totalQuantity, 32); err == nil {
			BuyPrice := config.Quantity / s
			AmountOfProfit := (BuyPrice / float64(100)) * float64(config.ProfitRate)
			var SellPrice float64 = BuyPrice + AmountOfProfit
			fmt.Printf("%v Buy Order Completed! Amount: %v Buy Price: %v\n", config.BaseAsset, totalQuantity, BuyPrice)
			Sell(totalQuantity, fmt.Sprintf("%f", SellPrice))
		}
	}
}

func Sell(totalQuantity string, SellPrice string) {
	fmt.Printf("%v Sell Order Completed! Amount: %v Sell Price: %v\n", config.BaseAsset, totalQuantity, SellPrice)
	order, err := client.NewCreateOrderService().Symbol(config.BaseAsset + config.QuoteAsset).
		Side(binance.SideTypeSell).Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).Quantity(totalQuantity).
		Price(SellPrice).Do(context.Background())
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(order.Status)
	}
}

func LoadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

type Config struct {
	Binance struct {
		BinanceAPIKey    string `json:"BinanceAPIKey"`
		BinanceSecretKey string `json:"BinanceSecretKey"`
	} `json:"Binance"`
	Twitter struct {
		TwitterConsumerKey    string `json:"TwitterConsumerKey"`
		TwitterConsumerSecret string `json:"TwitterConsumerSecret"`
		TwitterAccessToken    string `json:"TwitterAccessToken"`
		TwitterAccessSecret   string `json:"TwitterAccessSecret"`
		Elonid                string `json:"Elonid"`
	} `json:"Twitter"`
	BaseAsset  string  `json:"BaseAsset"`
	QuoteAsset string  `json:"QuoteAsset"`
	Quantity   float64 `json:"Quantity"`
	ProfitRate int     `json:"ProfitRate"`
	Fee        float32 `json:"Fee"`
}
