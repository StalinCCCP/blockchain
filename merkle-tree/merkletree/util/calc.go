package util

func Log2(x uint32) uint32 {
	var ret uint32 = 0
	for x > 1 {
		x /= 2
		ret++
	}
	return ret
}
func lowbit(x uint32) uint32 {
	return uint32(int32(x) & int32(-x))
}
func Lowcnt(x uint32) uint32 {
	return Log2(lowbit(x))
}
func Lson2Fa(x uint32) uint32 {
	return (x + 1) / 2
}

func Rson2Fa(x uint32) uint32 {
	return (x - 1) / 2
}

func Fa2Lson(x uint32) uint32 {
	return x*2 - 1
}
func Fa2Rson(x uint32) uint32 {
	return x*2 + 1
}

func IsLson(x uint32) bool {
	return (x+1)/2%2 == 1
}
