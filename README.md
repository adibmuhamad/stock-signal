## Stock Trading Signals

This application provides real-time stock trading signals based on the 50-day and 200-day Simple Moving Averages (SMA) crossover strategy combined with Fibonacci retracement levels. The application fetches stock data from the Yahoo Finance API and sends buy, sell, or hold signals for multiple stocks via WebSocket. It uses a very basic and unrealistic stock price prediction method. For accurate predictions, consider using advanced machine learning techniques or time series analysis methods.

## Features

1. Real-time stock trading signals
2. 50-day and 200-day SMA crossover strategy
3. Fibonacci retracement levels
4. WebSocket communication

## Usage

1. Ensure you have Go installed on your system.
2. Clone the repository and navigate to the project directory.
3. Run the server with `go run main.go`.
4. Connect to the websocket server with a client using a URL in the following format: `ws://localhost:8080/stock?symbols=<SYMBOLS>&ticker=<TICKER_INTERVAL>`.

- Replace `<SYMBOLS>` with a comma-separated list of stock symbols (e.g., AAPL,BBRI.JK).
- Replace `<TICKER_INTERVAL>` with the desired interval in seconds between signals.

## Contributing

Feel free to contribute! Here's how you can contribute:

- [Open an issue](https://github.com/adibmuhamad/stock-signal/issues) if you believe you've encountered a bug.
- Make a [pull request](https://github.com/adibmuhamad/stock-signal/pull) to add new features/make quality-of-life improvements/fix bugs.

## Author

- Muhammad Adib Yusrul Muna

## License
Copyright © 2023 Muhammad Adib Yusrul Muna

This software is distributed under the MIT license. See the [LICENSE](https://github.com/adibmuhamad/stock-signal/blob/main/LICENSE) file for the full license text.