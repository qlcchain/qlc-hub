package apis

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestEventAPI_checkWithdrawLimit(t *testing.T) {
	t.Skip()
	go resetWithdrawTimeLimit(context.Background(), 10)
	for i := 0; i < 300; i++ {
		time.Sleep(9 * time.Minute)
		addr := "0x255eEcd17E11C5d2FFD5818da31d04B5c1721D7C"
		b := isWithdrawLimitExceeded(addr)
		fmt.Println(b)
		if !b {
			setWithdrawLimitExceeded(addr)
		}
	}
	time.Sleep(300 * time.Minute)
}
