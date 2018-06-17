# `go_gdax_fix`
Go (golang) FIX Client for the Coinbase GDAX API https://www.gdax.com.

**Why:**

https://docs.gdax.com/#rate-limits

> FINANCIAL INFORMATION EXCHANGE API
> The FIX API throttles the number of incoming messages to 50 commands per second.

vs

> PRIVATE ENDPOINTS
> We throttle private endpoints by user ID: 5 requests per second, up to 10 requests per second in bursts.


## Run

* Replace `<GDAX_KEY>` in `client.template.cfg` with your GDAX key, then rename `client.template.cfg` to `client.cfg`.
* Set `GDAX_SECRET`, `GDAX_KEY` and `GDAX_PASSPHRASE` in `.env.template`, then rename `.env.template` to `.env`.
* `make run`.
* An attempt to login is made. A *logon* message will be received, *if and only if* you successfully login, otherwise a *logout* occurs.
* After *logon* is received a single sell order is placed, see the function `sendNewOrder`. You should see a new message in the log with `35=8`, where `35` is type and `8` is the value, in this case it is the *Execution Report* (https://docs.gdax.com/#execution-report-8).

```sh
INFO[0000] TradeClient:OnCreate                          sessionID="..."
INFO[0000] TradeClient:ToAdmin                           msg="..." sessionID="..."
INFO[0000] TradeClient:FromAdmin                         msg="..." sessionID="..."
INFO[0000] TradeClient:OnLogon                           sessionID="..."
^Csignal: interrupt

```

Thanks to: <https://github.com/quickfixgo/quickfix>.
