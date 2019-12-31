package engine

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"kunkka-match/enum"
	"kunkka-match/log"
	"time"
)

type Trade struct {
	MarkerId  string          `json:"markerId"`
	TakerId   string          `json:"takerId"`
	TakerSide string          `json:"takerSide"`
	Amount    decimal.Decimal `json:"amount"`
	Price     decimal.Decimal `json:"price"`
	Timestamp time.Time       `json:"timestamp"`
}

//成交对象结果转换为json字符串
func (this Trade) toJson() string {
	bytes, _ := json.Marshal(&this)
	return string(bytes)
}

//成交撮合
//做数量减法即可
func matchTrade(headOrder *Order, order *Order, book *OrderBook, lastTradePrice decimal.Decimal) *Order {
	var trade Trade
	trade.MarkerId = headOrder.OrderId
	trade.TakerId = order.OrderId
	trade.TakerSide = order.Side.String()
	trade.Timestamp = time.Now()
	result := order.Amount.Sub(headOrder.Amount)
	order.Amount = result
	if result.Cmp(decimal.Zero) >= 0 {
		if headOrder.Side == enum.SideBuy {

			book.removeBuyOrder(headOrder)
		} else if headOrder.Side == enum.SideSell {
			book.removeSellOrder(headOrder)
		}
	} else {

	}

	//TODO 发送成交数据给行情系统
	// 发送订单成交信息给业务系统
	return order
}

// 撤单
func dealCancel(order *Order, book *OrderBook) {
	var ok bool
	switch order.Side {
	case enum.SideBuy:
		book.removeBuyOrder(order)
	case enum.SideSell:
		book.removeSellOrder(order)
	}

	//TODO 移除缓存

	//TODO 发送到消息队列
	log.Info("engine: [%s],orderId: [%s] cancelResult: %v\n", order.Symbol, order.OrderId, ok)
}

//撤单逻辑处理
func cancelOrder(order *Order, book *OrderBook) {

}

// 创建订单
func dealCreate(order *Order, book *OrderBook, lastTradePrice decimal.Decimal) {
	switch order.OrderType {
	case enum.Limit:
		dealLimit(order, book, lastTradePrice)
	case enum.LimitIoc:
		//dealLimitIoc(order, book, lastTradePrice)
	case enum.Market:
		//dealMarket(order, book, lastTradePrice)
	case enum.MarketTop5:
		//dealMarketTop5(order, book, lastTradePrice)
	case enum.MarketTop10:
		//dealMarketTop10(order, book, lastTradePrice)
	case enum.MarketOpponent:
		//dealMarketOpponent(order, book, lastTradePrice)

	}
}

//限价挂单
func dealLimit(order *Order, book *OrderBook, lastTradePrice decimal.Decimal) {
	switch order.Side {
	case enum.SideBuy:
		dealBuyLimit(order, book, lastTradePrice)
	case enum.SideSell:
		dealSellLimit(order, book, lastTradePrice)
	}
}

//限价挂单  -- 买单
func dealBuyLimit(order *Order, book *OrderBook, lastTradePrice decimal.Decimal) {
LOOP:
	headOrder := book.getHeadSellOrder()
	if headOrder == (Order{}) || order.Price.LessThan(headOrder.Price) {
		book.addBuyOrder(*order)
		log.Info("engine %s, a order has added to the orderBook: %s\n", order.Symbol, order.toJson())
	} else {
		matchTrade(&headOrder, order, book, lastTradePrice)
		if order.Amount.IsPositive() {
			goto LOOP
		}
	}
}

//限价挂单 -- 卖单
func dealSellLimit(order *Order, book *OrderBook, lastTradePrice decimal.Decimal) {
	log.Info("receive sell limit order: %s", order.toJson())
LOOP:
	headOrder := book.getHeadBuyOrder()
	if headOrder == (Order{}) || order.Price.GreaterThan(headOrder.Price) {
		book.addSellOrder(*order)
		log.Info("engine %s, a order added to the orderBook: %s\n", order.Symbol, order.toJson())
	} else {
		matchTrade(&headOrder, order, book, lastTradePrice)
		if order.Amount.IsPositive() {
			goto LOOP
		}
	}
}
