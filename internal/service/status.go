package service

type OrderStatus string

const (
	OrderStateReceived  OrderStatus = "received"  // 주문 접수됨
	OrderStatePaid      OrderStatus = "paid"      // 결제 완료
	OrderStatePreparing OrderStatus = "preparing" // 상품 준비중
	OrderStateShipped   OrderStatus = "shipped"   // 배송중
	OrderStateDelivered OrderStatus = "delivered" // 배송완료
	OrderStateCancelled OrderStatus = "cancelled" // 주문 취소됨
	OrderStateRefunding OrderStatus = "refunding" // 환불 처리중
	OrderStateRefunded  OrderStatus = "refunded"  // 환불 완료
)
