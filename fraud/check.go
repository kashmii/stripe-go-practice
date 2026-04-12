package fraud

import (
	"fmt"

	"github.com/stripe/stripe-go/v85"
)

// 許容する最大金額（円）
const maxAmount int64 = 50000

// Check は PaymentIntent の内容を検証し、不正と判断した場合にエラーを返す
func Check(pi *stripe.PaymentIntent) error {
	// NOTE: 形式的な処理だけに留める。この処理を本格的にすることは現時点でプロジェクトの主目的ではないため
	if pi.Amount > maxAmount {
		return fmt.Errorf("amount %d exceeds the maximum allowed amount %d", pi.Amount, maxAmount)
	}
	return nil
}
