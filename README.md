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
* An attempt to login is made. A logon message will be received, *if and only if* you successfully login, otherwise a *logout* occurs.

```sh
INFO[0000] TradeClient:OnCreate                          sessionID="..."
INFO[0000] TradeClient:ToAdmin                           msg="..." sessionID="..."
INFO[0000] TradeClient:FromAdmin                         msg="..." sessionID="..."
INFO[0000] TradeClient:OnLogon                           sessionID="..."
^Csignal: interrupt

```

**TODO:**
* Place single market order.

Thanks to: <https://github.com/quickfixgo/quickfix>.
