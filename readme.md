1) The app uses binance API for exchange rates - binance
2) The app uses the external library github.com/go-telegram-bot-api/telegram-bot-api/v5 to create a Telegram bot. The library connects to a bot created with BotFather

Bot has commands:
ADD - Example: "ADD BTC 2" - adds 2 bitcoins to the wallet
SUB - Example: "SUB BTC 1" - removes 1 bitcoin from the wallet
DEL - Example: "DEL BTC" - completely removes the currency from the wallet
SHOW - Example: "SHOW" - shows the balance for all the coins and their total value in USD and in RUB

TODO:
Add database connectivity 
Add commands using /