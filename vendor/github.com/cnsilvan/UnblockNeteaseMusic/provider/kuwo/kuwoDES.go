package kuwo

import (
	"math/big"

	"github.com/cnsilvan/UnblockNeteaseMusic/pool"
)

//
//	Thanks to
//	https://github.com/Levi233/MusicPlayer/blob/master/app/src/main/java/com/chenhao/musicplayer/utils/crypt/KuwoDES.java
//
var (
	SECRET_KEY = []byte("ylzsxkwm")
	gArrayMask []*big.Int
	//gArrayMask = []*big.Int{new(big.Int).SetUint64(0x0000000000000001), new(big.Int).SetUint64(0x0000000000000002),
	//	new(big.Int).SetUint64(0x0000000000000004), new(big.Int).SetUint64(0x0000000000000008), new(big.Int).SetUint64(0x0000000000000010),
	//	new(big.Int).SetUint64(0x0000000000000020), new(big.Int).SetUint64(0x0000000000000040), new(big.Int).SetUint64(0x0000000000000080),
	//	new(big.Int).SetUint64(0x0000000000000100), new(big.Int).SetUint64(0x0000000000000200), new(big.Int).SetUint64(0x0000000000000400),
	//	new(big.Int).SetUint64(0x0000000000000800), new(big.Int).SetUint64(0x0000000000001000), new(big.Int).SetUint64(0x0000000000002000),
	//	new(big.Int).SetUint64(0x0000000000004000), new(big.Int).SetUint64(0x0000000000008000), new(big.Int).SetUint64(0x0000000000010000),
	//	new(big.Int).SetUint64(0x0000000000020000), new(big.Int).SetUint64(0x0000000000040000), new(big.Int).SetUint64(0x0000000000080000),
	//	new(big.Int).SetUint64(0x0000000000100000), new(big.Int).SetUint64(0x0000000000200000), new(big.Int).SetUint64(0x0000000000400000),
	//	new(big.Int).SetUint64(0x0000000000800000), new(big.Int).SetUint64(0x0000000001000000), new(big.Int).SetUint64(0x0000000002000000),
	//	new(big.Int).SetUint64(0x0000000004000000), new(big.Int).SetUint64(0x0000000008000000), new(big.Int).SetUint64(0x0000000010000000),
	//	new(big.Int).SetUint64(0x0000000020000000), new(big.Int).SetUint64(0x0000000040000000), new(big.Int).SetUint64(0x0000000080000000),
	//	new(big.Int).SetUint64(0x0000000100000000), new(big.Int).SetUint64(0x0000000200000000), new(big.Int).SetUint64(0x0000000400000000),
	//	new(big.Int).SetUint64(0x0000000800000000), new(big.Int).SetUint64(0x0000001000000000), new(big.Int).SetUint64(0x0000002000000000),
	//	new(big.Int).SetUint64(0x0000004000000000), new(big.Int).SetUint64(0x0000008000000000), new(big.Int).SetUint64(0x0000010000000000),
	//	new(big.Int).SetUint64(0x0000020000000000), new(big.Int).SetUint64(0x0000040000000000), new(big.Int).SetUint64(0x0000080000000000),
	//	new(big.Int).SetUint64(0x0000100000000000), new(big.Int).SetUint64(0x0000200000000000), new(big.Int).SetUint64(0x0000400000000000),
	//	new(big.Int).SetUint64(0x0000800000000000), new(big.Int).SetUint64(0x0001000000000000), new(big.Int).SetUint64(0x0002000000000000),
	//	new(big.Int).SetUint64(0x0004000000000000), new(big.Int).SetUint64(0x0008000000000000), new(big.Int).SetUint64(0x0010000000000000),
	//	new(big.Int).SetUint64(0x0020000000000000), new(big.Int).SetUint64(0x0040000000000000), new(big.Int).SetUint64(0x0080000000000000),
	//	new(big.Int).SetUint64(0x0100000000000000), new(big.Int).SetUint64(0x0200000000000000), new(big.Int).SetUint64(0x0400000000000000),
	//	new(big.Int).SetUint64(0x0800000000000000), new(big.Int).SetUint64(0x1000000000000000), new(big.Int).SetUint64(0x2000000000000000),
	//	new(big.Int).SetUint64(0x4000000000000000), new(big.Int).Neg(new(big.Int).SetUint64(0x8000000000000000))}
	gArrayIp = []int{57, 49, 41, 33, 25, 17, 9, 1, 59, 51, 43, 35, 27, 19, 11, 3, 61, 53, 45, 37, 29, 21, 13, 5, 63, 55, 47, 39, 31, 23, 15, 7,
		56, 48, 40, 32, 24, 16, 8, 0, 58, 50, 42, 34, 26, 18, 10, 2, 60, 52, 44, 36, 28, 20, 12, 4, 62, 54, 46, 38, 30, 22, 14, 6}
	gArrayE = []int{31, 0, 1, 2, 3, 4, -1, -1, 3, 4, 5, 6, 7, 8, -1, -1, 7, 8, 9, 10, 11, 12, -1, -1, 11, 12, 13, 14, 15, 16, -1, -1, 15, 16, 17,
		18, 19, 20, -1, -1, 19, 20, 21, 22, 23, 24, -1, -1, 23, 24, 25, 26, 27, 28, -1, -1, 27, 28, 29, 30, 31, 30, -1, -1}
	gMatrixnsBox = [][]int{
		{14, 4, 3, 15, 2, 13, 5, 3, 13, 14, 6, 9, 11, 2, 0, 5, 4, 1, 10, 12, 15, 6, 9, 10, 1, 8, 12, 7, 8, 11, 7, 0, 0, 15, 10, 5, 14, 4, 9, 10, 7, 8, 12, 3, 13, 1, 3, 6, 15, 12, 6, 11, 2, 9, 5, 0, 4, 2, 11, 14, 1, 7, 8, 13},
		{15, 0, 9, 5, 6, 10, 12, 9, 8, 7, 2, 12, 3, 13, 5, 2, 1, 14, 7, 8, 11, 4, 0, 3, 14, 11, 13, 6, 4, 1, 10, 15, 3, 13, 12, 11, 15, 3, 6, 0, 4, 10, 1, 7, 8, 4, 11, 14, 13, 8, 0, 6, 2, 15, 9, 5, 7, 1, 10, 12, 14, 2, 5, 9},
		{10, 13, 1, 11, 6, 8, 11, 5, 9, 4, 12, 2, 15, 3, 2, 14, 0, 6, 13, 1, 3, 15, 4, 10, 14, 9, 7, 12, 5, 0, 8, 7, 13, 1, 2, 4, 3, 6, 12, 11, 0, 13, 5, 14, 6, 8, 15, 2, 7, 10, 8, 15, 4, 9, 11, 5, 9, 0, 14, 3, 10, 7, 1, 12},
		{7, 10, 1, 15, 0, 12, 11, 5, 14, 9, 8, 3, 9, 7, 4, 8, 13, 6, 2, 1, 6, 11, 12, 2, 3, 0, 5, 14, 10, 13, 15, 4, 13, 3, 4, 9, 6, 10, 1, 12, 11, 0, 2, 5, 0, 13, 14, 2, 8, 15, 7, 4, 15, 1, 10, 7, 5, 6, 12, 11, 3, 8, 9, 14},
		{2, 4, 8, 15, 7, 10, 13, 6, 4, 1, 3, 12, 11, 7, 14, 0, 12, 2, 5, 9, 10, 13, 0, 3, 1, 11, 15, 5, 6, 8, 9, 14, 14, 11, 5, 6, 4, 1, 3, 10, 2, 12, 15, 0, 13, 2, 8, 5, 11, 8, 0, 15, 7, 14, 9, 4, 12, 7, 10, 9, 1, 13, 6, 3},
		{12, 9, 0, 7, 9, 2, 14, 1, 10, 15, 3, 4, 6, 12, 5, 11, 1, 14, 13, 0, 2, 8, 7, 13, 15, 5, 4, 10, 8, 3, 11, 6, 10, 4, 6, 11, 7, 9, 0, 6, 4, 2, 13, 1, 9, 15, 3, 8, 15, 3, 1, 14, 12, 5, 11, 0, 2, 12, 14, 7, 5, 10, 8, 13},
		{4, 1, 3, 10, 15, 12, 5, 0, 2, 11, 9, 6, 8, 7, 6, 9, 11, 4, 12, 15, 0, 3, 10, 5, 14, 13, 7, 8, 13, 14, 1, 2, 13, 6, 14, 9, 4, 1, 2, 14, 11, 13, 5, 0, 1, 10, 8, 3, 0, 11, 3, 5, 9, 4, 15, 2, 7, 8, 12, 15, 10, 7, 6, 12},
		{13, 7, 10, 0, 6, 9, 5, 15, 8, 4, 3, 10, 11, 14, 12, 5, 2, 11, 9, 6, 15, 12, 0, 3, 4, 1, 14, 13, 1, 2, 7, 8, 1, 2, 12, 15, 10, 4, 0, 3, 13, 14, 6, 9, 7, 8, 9, 6, 15, 1, 5, 12, 3, 10, 14, 5, 8, 7, 11, 0, 4, 13, 2, 11}}
	gArrayP        = []int{15, 6, 19, 20, 28, 11, 27, 16, 0, 14, 22, 25, 4, 17, 30, 9, 1, 7, 23, 13, 31, 26, 2, 8, 18, 12, 29, 5, 21, 10, 3, 24}
	gArrayIP1      = []int{39, 7, 47, 15, 55, 23, 63, 31, 38, 6, 46, 14, 54, 22, 62, 30, 37, 5, 45, 13, 53, 21, 61, 29, 36, 4, 44, 12, 52, 20, 60, 28, 35, 3, 43, 11, 51, 19, 59, 27, 34, 2, 42, 10, 50, 18, 58, 26, 33, 1, 41, 9, 49, 17, 57, 25, 32, 0, 40, 8, 48, 16, 56, 24}
	gArrayPC1      = []int{56, 48, 40, 32, 24, 16, 8, 0, 57, 49, 41, 33, 25, 17, 9, 1, 58, 50, 42, 34, 26, 18, 10, 2, 59, 51, 43, 35, 62, 54, 46, 38, 30, 22, 14, 6, 61, 53, 45, 37, 29, 21, 13, 5, 60, 52, 44, 36, 28, 20, 12, 4, 27, 19, 11, 3}
	gArrayPC2      = []int{13, 16, 10, 23, 0, 4, -1, -1, 2, 27, 14, 5, 20, 9, -1, -1, 22, 18, 11, 3, 25, 7, -1, -1, 15, 6, 26, 19, 12, 1, -1, -1, 40, 51, 30, 36, 46, 54, -1, -1, 29, 39, 50, 44, 32, 47, -1, -1, 43, 48, 38, 55, 33, 52, -1, -1, 45, 41, 49, 35, 28, 31, -1, -1}
	gArrayLs       = []int{1, 1, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 1}
	gArrayLsMask   = []uint64{0x0000000000000000, 0x0000000000100001, 0x0000000000300003}
	DesModeEncrypt = 0
	DesModeDecrypt = 1
)

func init() {
	gArrayMask = make([]*big.Int, 64)
	for i := 0; i < 64; i++ {
		ui := new(big.Int).SetUint64(1 << i)
		if i == 63 {
			ui = ui.Neg(ui)
		}
		gArrayMask[i] = ui
	}
}
func bitTransform(array []int, len int, bts *big.Int) *big.Int {
	var bti int
	var dest = new(big.Int).SetUint64(0)
	for bti = 0; bti < len; bti++ {
		bi := pool.GetBigInt()
		if array[bti] >= 0 && bi.And(bts, gArrayMask[array[bti]]).Uint64() != 0 {
			dest = dest.Or(dest, gArrayMask[bti])
		}
		pool.PutBigInt(bi)

	}
	return dest
}

func desSubKeys(key *big.Int, K []*big.Int, mode int) {
	var j int
	/* PC-1变换 */
	temp := bitTransform(gArrayPC1, 56, key)
	mask := pool.GetBigInt()
	left := pool.GetBigInt()
	right := pool.GetBigInt()
	for j = 0; j < 16; j++ {
		/* 循环左移 */
		{
			source := temp
			mask = mask.SetUint64(gArrayLsMask[gArrayLs[j]])
			left = left.And(source, mask)
			left = left.Lsh(left, uint(28-gArrayLs[j]))
			right = right.And(source, mask.Not(mask))
			right = right.Rsh(right, uint(gArrayLs[j]))
			temp = temp.Or(left, right)
		}
		if j == 15 {
			pool.PutBigInt(mask)
			pool.PutBigInt(left)
			pool.PutBigInt(right)
		}
		/* PC-2变换 */
		// 要初始化k的元素为0
		K[j] = bitTransform(gArrayPC2, 64, temp)
	}
	if mode == DesModeDecrypt { /* 如果解密则反转子密钥顺序 */
		var t *big.Int
		for j = 0; j < 8; j++ {
			t = K[j]
			K[j] = K[15-j]
			K[15-j] = t
		}
	}
}

func des64(subkeys []*big.Int, data *big.Int) *big.Int {
	var SOut = new(big.Int)
	var L *big.Int
	var R *big.Int
	var pSource = make([]*big.Int, 2)
	var pR = make([]*big.Int, 8)
	var sbi int
	// var i int;
	//IP变换
	out := bitTransform(gArrayIp, 64, data)
	temp := pool.GetBigInt()
	defer pool.PutBigInt(temp)
	temp.SetUint64(0x00000000ffffffff)
	pSource[0] = pool.GetBigInt()
	defer pool.PutBigInt(pSource[0])
	pSource[0] = pSource[0].And(out, temp)
	temp.SetInt64(-4294967296)
	pSource[1] = pool.GetBigInt()
	defer pool.PutBigInt(pSource[1])
	pSource[1] = pSource[1].And(out, temp)

	pSource[1] = pSource[1].Rsh(pSource[1], 32)
	/* 主迭代 */
	// source = out;
	for i := 0; i < 16; i++ {
		/* F变换开始 */
		R = temp.Set(pSource[1])
		/* E变换 */
		R = bitTransform(gArrayE, 64, R)
		/* 与子密钥异或 */
		R = R.Xor(R, subkeys[i])
		/* S盒变换 */
		tmpp := pool.GetBigInt()
		for k := 0; k < 8; k++ {
			kk := pool.GetBigInt()
			kk = kk.Rsh(R, uint(k*8))
			pR[k] = kk.And(kk, tmpp.SetUint64(0xff))
		}
		SOut = SOut.SetUint64(0)
		for sbi = 7; sbi >= 0; sbi-- {
			pRTemp := pR[sbi]
			SOut = SOut.Lsh(SOut, 4).Or(SOut, tmpp.SetInt64(int64(gMatrixnsBox[sbi][pRTemp.Uint64()])))
			pool.PutBigInt(pRTemp)
		}
		pool.PutBigInt(tmpp)
		//R = SOut
		/* P变换 */
		R = bitTransform(gArrayP, 32, SOut)

		/* f变换完成 */
		L = pool.GetBigInt()
		L = L.Set(pSource[0])
		pSource[0] = pSource[0].Set(pSource[1])
		pSource[1] = pSource[1].Xor(L, R)
		pool.PutBigInt(L)
	}

	/* 交换高低32位 */
	pSource[1], pSource[0] = pSource[0], pSource[1]
	tmp1 := pool.GetBigInt()
	defer pool.PutBigInt(tmp1)
	/* IP-1变换 */
	out = out.Lsh(pSource[1], 32).And(out, tmp1.SetInt64(-4294967296)).Or(out, pSource[0].And(pSource[0], tmp1.SetInt64(0x00000000ffffffff)))
	out = bitTransform(gArrayIP1, 64, out)
	return new(big.Int).Set(out)
}
func Encrypt(src []byte) []byte {
	by:=encrypt(src, SECRET_KEY)
	return by
}

func encrypt(src []byte, key []byte) []byte {
	// long keyl = Long.valueOf(new String(key));
	var keyl = pool.GetBigInt()
	srcLength := len(src)
	for i := 0; i < 8; i++ {
		kB := pool.GetBigInt().SetBytes([]byte{key[i]})
		b := kB.Lsh(kB, uint(i*8))
		keyl.Set(b.Or(kB, keyl))
		pool.PutBigInt(kB)
	}
	num := srcLength / 8
	// 子密钥（临时数据）
	subKey := make([]*big.Int, 16)
	desSubKeys(keyl, subKey, DesModeEncrypt)
	// 加密
	pSrc := make([]int64, num)
	for i := 0; i < num; i++ {
		for j := 0; j < 8; j++ {
			pSrc[i] |= int64(src[i*8+j]) << (j * 8)
		}
	}
	var pEncyrptLength = ((num+1)*8 + 1) / 8
	//存放密文
	pEncyrpt := make([]*big.Int, pEncyrptLength)
	// 计算前部的数据块(除了最后一部分)
	for i := 0; i < num; i++ {
		pEncyrpt[i] = des64(subKey, new(big.Int).SetInt64(pSrc[i]))
	}

	szTail := make([]byte, srcLength-num*8)
	// 保存多出来的字节
	copy(szTail, src[num*8:])

	var tail64 int64 = 0
	// 处理结尾处不够8个字节的部分
	tailNum := srcLength % 8
	for i := 0; i < tailNum; i++ {
		tail64 = tail64 | (int64(szTail[i]) << (i * 8))
	}
	// 计算多出的那一位(最后一位)
	pEncyrpt[num] = des64(subKey, new(big.Int).SetInt64(tail64))
	result := make([]byte, len(pEncyrpt)*8)
	temp := 0
	ff := pool.GetBigInt().SetInt64(255)
	defer pool.PutBigInt(ff)
	enc := pool.GetBigInt()
	defer pool.PutBigInt(enc)
	// 将密文转为字节型
	for i := 0; i < len(pEncyrpt); i++ {
		for j := 0; j < 8; j++ {
			enc = enc.Set(pEncyrpt[i])
			is := enc.Rsh(enc, uint(j*8))
			r := is.And(enc, ff).Int64()
			result[temp] = byte(r)
			temp++
		}
	}
	return result
}
