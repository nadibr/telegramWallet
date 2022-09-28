package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/spf13/viper"
)

type binanceResp struct {
	Price float64 `json:"price,string"`
	Code  int64   `json:"code"`
}

type wallet map[string]float64 //currency data with their amounts

var db = map[int64]wallet{} //chats with their wallets

func main() {

	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	token := viper.Get("tgtoken")

	bot, err := tgbotapi.NewBotAPI(token.(string))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates { //channel
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		//log.Println(update.Message.Text)

		msgArr := strings.Split(update.Message.Text, " ") //array of user message's elements

		switch msgArr[0] {
		case "ADD":
			summ, err := strconv.ParseFloat(msgArr[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Impossible to convert the amount"))
				continue
			}

			if _, ok := db[update.Message.Chat.ID]; !ok { //check if chat already exists, if not, then add it
				db[update.Message.Chat.ID] = wallet{}
			}

			db[update.Message.Chat.ID][msgArr[1]] += summ //add currency to the existing balance

			msg := fmt.Sprintf("Balance: %s %f", msgArr[1], db[update.Message.Chat.ID][msgArr[1]]) //
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		case "SUB":
			var msg string
			summ, err := strconv.ParseFloat(msgArr[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Impossible to convert the amount"))
				continue
			}

			if _, ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID] = wallet{}
			}

			if summ <= db[update.Message.Chat.ID][msgArr[1]] {

				db[update.Message.Chat.ID][msgArr[1]] -= summ
				msg = fmt.Sprintf("Balance: %s %f", msgArr[1], db[update.Message.Chat.ID][msgArr[1]])
			} else {
				msg = fmt.Sprintf("Not enough funds to substract. Balance: %s %f", msgArr[1], db[update.Message.Chat.ID][msgArr[1]])
			}

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		case "DEL":
			delete(db[update.Message.Chat.ID], msgArr[1])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Currency deleted"))
		case "SHOW":
			msg := "Balance: \n"
			var usdSumm, rubSumm float64
			for key, value := range db[update.Message.Chat.ID] {
				coinPriceUSD, err := getPrice(key, "USDT")
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				}
				usdSumm += value * coinPriceUSD
				msg += fmt.Sprintf("%s: %f [%.2f]\n", key, value, value*coinPriceUSD)

				coinPriceRUB, err := getPrice(key, "RUB")
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				}
				rubSumm += value * coinPriceRUB
				//msg += fmt.Sprintf("%s: %f [%.2f]\n", key, value, value*coinPriceRUB)
			}
			msg += fmt.Sprintf("Total in USD: %.2f\n", usdSumm)
			msg += fmt.Sprintf("Total in RUB: %.2f\n", rubSumm)

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command"))
		}

	}
}

//common function for getting a price for required currency
func getPrice(coin, curr string) (price float64, err error) { //getting currency price from binance
	resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s%s", coin, curr))
	if err != nil {
		return
	}

	defer resp.Body.Close()

	var jsonResp binanceResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}

	if jsonResp.Code != 0 {
		err = errors.New("Unknown currency")
		return
	}

	price = jsonResp.Price

	return
}
