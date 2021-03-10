package hub

// 检查通道是否关闭
func isClosedChan(stop chan interface{}) bool {
	if cap(stop) == 0 {
		return true
	}

	select {
	case _, ok := <-stop:
		if !ok {
			return true
		}
		return false
	default:
		return false
	}
}
