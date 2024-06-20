package pipe

func SendRecords[R any](records []R, outCh chan<- R, stopped <-chan struct{}) {
	for _, record := range records {
		select {
		case <-stopped:
			return
		default:
			select {
			case <-stopped:
				return
			case outCh <- record:
			}
		}
	}
}
