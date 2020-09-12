package apis

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestEventAPI_checkWithdrawLimit(t *testing.T) {
	t.Skip()
	go resetWithdrawTimeLimit(context.Background(), 1)
	for i := 0; i < 30; i++ {
		time.Sleep(9 * time.Second)
		b := isWithdrawLimitExceeded("0x255eEcd17E11C5d2FFD5818da31d04B5c1721D7C")
		fmt.Println(b)
	}
	time.Sleep(5 * time.Second)
}
