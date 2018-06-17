package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/quickfixgo/quickfix"
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

func main() {
	flag.Parse()

	cfgFileName := "client.cfg"
	if flag.NArg() > 0 {
		cfgFileName = flag.Arg(0)
	}

	cfg, err := os.Open(cfgFileName)
	if err != nil {
		fmt.Printf("Error opening %v, %v\n", cfgFileName, err)
		return
	}

	appSettings, err := quickfix.ParseSettings(cfg)
	if err != nil {
		fmt.Println("Error reading cfg,", err)
		return
	}

	app := &TradeClient{}
	fileLogFactory, err := quickfix.NewFileLogFactory(appSettings)

	if err != nil {
		fmt.Println("Error creating file log factory,", err)
		return
	}

	initiator, err := quickfix.NewInitiator(app, quickfix.NewMemoryStoreFactory(), appSettings, fileLogFactory)
	if err != nil {
		fmt.Printf("Unable to create Initiator: %s\n", err)
		return
	}

	initiator.Start()
	for {
	}

	// Loop:
	// 	for {
	// 		action, err := internal.QueryAction()
	// 		if err != nil {
	// 			break
	// 		}

	// 		switch action {
	// 		case "1":
	// 			err = internal.QueryEnterOrder()

	// 		case "2":
	// 			err = internal.QueryCancelOrder()

	// 		case "3":
	// 			err = internal.QueryMarketDataRequest()

	// 		case "4":
	// 			//quit
	// 			break Loop

	// 		default:
	// 			err = fmt.Errorf("unknown action: '%v'", action)
	// 		}

	// 		if err != nil {
	// 			fmt.Printf("%v\n", err)
	// 		}
	// 	}

	// initiator.Stop()
}
