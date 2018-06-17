package go_gdax_fix

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	fix42nos "github.com/quickfixgo/fix42/newordersingle"
	"github.com/quickfixgo/quickfix"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

//TradeClient implements the quickfix.Application interface
type TradeClient struct {
}

//OnCreate implemented as part of Application interface
func (e *TradeClient) OnCreate(sessionID quickfix.SessionID) {
	logrus.WithFields(logrus.Fields{
		"sessionID": fmt.Sprintf("%#v", sessionID),
	}).Info("TradeClient:OnCreate")
	return
}

//OnLogon implemented as part of Application interface
func (e *TradeClient) OnLogon(sessionID quickfix.SessionID) {
	logrus.WithFields(logrus.Fields{
		"sessionID": fmt.Sprintf("%#v", sessionID),
	}).Info("TradeClient:OnLogon")
	sendNewOrder()

	return
}

//OnLogout implemented as part of Application interface
func (e *TradeClient) OnLogout(sessionID quickfix.SessionID) {
	logrus.WithFields(logrus.Fields{
		"sessionID": fmt.Sprintf("%#v", sessionID),
	}).Info("TradeClient:OnLogout")
	return
}

//FromAdmin implemented as part of Application interface
func (e *TradeClient) FromAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	logrus.WithFields(logrus.Fields{
		"msg":       fmt.Sprintf("%+v", msg),
		"sessionID": fmt.Sprintf("%#v", sessionID),
	}).Info("TradeClient:FromAdmin")
	return
}

//ToAdmin implemented as part of Application interface
func (e *TradeClient) ToAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) {
	isLogonMsg := msg.IsMsgTypeOf("A")
	if !isLogonMsg {
		return
	}

	// logrus.WithFields(logrus.Fields{
	// 	"msg":       fmt.Sprintf("%+v", msg),
	// 	"sessionID": fmt.Sprintf("%+v", sessionID),
	// }).Info("TradeClient:ToAdmin")
	initLogonMessage(msg)
	logrus.WithFields(logrus.Fields{
		"msg":       fmt.Sprintf("%+v", msg),
		"sessionID": fmt.Sprintf("%#v", sessionID),
	}).Info("TradeClient:ToAdmin")
	return
}

//ToApp implemented as part of Application interface
func (e *TradeClient) ToApp(msg *quickfix.Message, sessionID quickfix.SessionID) (err error) {
	logrus.WithFields(logrus.Fields{
		"msg":       fmt.Sprintf("%+v", msg),
		"sessionID": fmt.Sprintf("%#v", sessionID),
	}).Info("TradeClient:ToApp")
	// fmt.Printf("Sending %s\n", msg)
	return
}

//FromApp implemented as part of Application interface. This is the callback for all Application level messages from the counter party.
func (e *TradeClient) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	logrus.WithFields(logrus.Fields{
		"msg":       fmt.Sprintf("%+v", msg),
		"sessionID": fmt.Sprintf("%#v", sessionID),
	}).Info("TradeClient:FromApp")
	// fmt.Printf("FromApp: %s\n", msg.String())
	return
}

func sendNewOrder() {
	id, _ := uuid.NewUUID()
	clordid := id.String()

	order := fix42nos.New(
		field.NewClOrdID(clordid),
		field.NewHandlInst("1"),
		field.NewSymbol("BTC-EUR"),
		field.NewSide(enum.Side_SELL),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT),
	)

	qty, _ := decimal.NewFromString("0.01")
	order.Set(field.NewOrderQty(qty, 2))

	price, _ := decimal.NewFromString("9000.00")
	order.Set(field.NewPrice(price, 2))

	order.Set(field.NewTimeInForce(enum.TimeInForce_GOOD_TILL_CANCEL))

	msg := order.ToMessage()

	msg.Header.Set(field.NewSenderCompID(os.Getenv("GDAX_KEY")))
	msg.Header.Set(field.NewTargetCompID("Coinbase"))

	logrus.WithFields(logrus.Fields{
		"msg": fmt.Sprintf("%+v", msg),
	}).Info("TradeClient:sendNewOrder")

	quickfix.Send(msg)
}

func initLogonMessage(msg *quickfix.Message) {
	// 98	EncryptMethod	Must be 0 (None)
	msg.Body.SetInt(98, 0)
	// 108	HeartBtInt	Must be 30 (seconds)
	msg.Body.SetInt(108, 30)
	// 554	Password	Client API passphrase
	msg.Body.SetString(554, os.Getenv("GDAX_PASSPHRASE"))
	// 96	RawData	Client message signature (see below)
	msg.Body.SetString(96, rawData())
	// 8013	CancelOrdersOnDisconnect	Y: Cancel all open orders for the current profile; S: Cancel open orders placed during session
	msg.Body.SetString(8013, "S")
	// 9406	DropCopyFlag	If set to Y, execution reports will be generated for all user orders (defaults to Y)
	msg.Body.SetString(9406, "Y")

	// Override the time
	msg.Header.SetString(52, strconv.FormatInt(time.Now().Unix(), 10))
}

func rawData() string {
	// sendingTime := time.Now().UTC().Format(time.RFC3339Nano)
	// sendingTime := time.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00")
	sendingTime := strconv.FormatInt(time.Now().Unix(), 10)
	msgType := "A" // https://docs.gdax.com/#connectivity
	msgSeqNum := "1"
	senderCompID := os.Getenv("GDAX_KEY")
	targetCompID := "Coinbase"
	password := os.Getenv("GDAX_PASSPHRASE")

	prehash := strings.Join([]string{
		sendingTime,
		msgType,
		msgSeqNum,
		senderCompID,
		targetCompID,
		password,
	}, string("\x01"))
	return sign(prehash)
}

func sign(prehash string) string {
	secret := os.Getenv("GDAX_SECRET")
	if secret == "" {
		logrus.WithFields(logrus.Fields{
			"err": errors.New("GDAX_SECRET not set"),
		}).Fatal("sign")
	}

	key, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("sign")
	}

	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write([]byte(prehash)); err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("sign")
	}

	encoded := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return encoded
}
