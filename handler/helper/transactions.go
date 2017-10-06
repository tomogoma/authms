package helper

func TransactionSerializer(tIDCh chan int) {

	tID := 0
	for {
		tID++
		tIDCh <- tID
	}
}
