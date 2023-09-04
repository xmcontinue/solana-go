package sol

import (
	"math/big"
)

func GetSqrtPriceAtTick(tick int32) *big.Int {
	if tick >= 0 {
		return GetSqrtPriceAtPositiveTick(tick)
	} else {
		return GetSqrtPriceAtNegativeTick(tick)
	}
}

func GetSqrtPriceAtPositiveTick(tick int32) *big.Int {

	ratio := &big.Int{}
	if (tick & 1) != 0 {
		ratio = newBigInt("79232123823359799118286999567")
	} else {
		ratio = newBigInt("79228162514264337593543950336")
	}

	if (tick & 2) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("79236085330515764027303304731")),
			96,
		)
	}
	if (tick & 4) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("79244008939048815603706035061")),
			96,
		)
	}
	if (tick & 8) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("79259858533276714757314932305")),
			96,
		)
	}
	if (tick & 16) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("79291567232598584799939703904")),
			96,
		)
	}
	if (tick & 32) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("79355022692464371645785046466")),
			96,
		)
	}
	if (tick & 64) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("79482085999252804386437311141")),
			96,
		)
	}
	if (tick & 128) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("79736823300114093921829183326")),
			96,
		)
	}
	if (tick & 256) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("80248749790819932309965073892")),
			96,
		)
	}
	if (tick & 512) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("81282483887344747381513967011")),
			96,
		)
	}
	if (tick & 1024) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("83390072131320151908154831281")),
			96,
		)
	}
	if (tick & 2048) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("87770609709833776024991924138")),
			96,
		)
	}
	if (tick & 4096) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("97234110755111693312479820773")),
			96,
		)
	}
	if (tick & 8192) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("119332217159966728226237229890")),
			96,
		)
	}
	if (tick & 16384) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("179736315981702064433883588727")),
			96,
		)
	}
	if (tick & 32768) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("407748233172238350107850275304")),
			96,
		)
	}
	if (tick & 65536) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("2098478828474011932436660412517")),
			96,
		)
	}
	if (tick & 131072) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("55581415166113811149459800483533")),
			96,
		)
	}
	if (tick & 262144) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("38992368544603139932233054999993551")),
			96,
		)
	}

	return signedShiftRight(ratio, 32)
}

func GetSqrtPriceAtNegativeTick(tick int32) *big.Int {
	if tick < 0 {
		tick = -tick
	}

	ratio := &big.Int{}
	if (tick & 1) != 0 {
		ratio = newBigInt("18445821805675392311")
	} else {
		ratio = newBigInt("18446744073709551616")
	}

	if (tick & 2) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("18444899583751176498")),
			64,
		)
	}
	if (tick & 4) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("18443055278223354162")),
			64,
		)
	}
	if (tick & 8) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("18439367220385604838")),
			64,
		)
	}
	if (tick & 16) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("18431993317065449817")),
			64,
		)
	}
	if (tick & 32) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("18417254355718160513")),
			64,
		)
	}
	if (tick & 64) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("18387811781193591352")),
			64,
		)
	}
	if (tick & 128) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("18329067761203520168")),
			64,
		)
	}
	if (tick & 256) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("18212142134806087854")),
			64,
		)
	}
	if (tick & 512) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("17980523815641551639")),
			64,
		)
	}
	if (tick & 1024) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("17526086738831147013")),
			64,
		)
	}
	if (tick & 2048) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("16651378430235024244")),
			64,
		)
	}
	if (tick & 4096) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("15030750278693429944")),
			64,
		)
	}
	if (tick & 8192) != 0 {
		ratio = signedShiftRight(
			ratio.Mul(ratio, newBigInt("12247334978882834399")),
			64,
		)
	}
	if (tick & 16384) != 0 {
		ratio = signedShiftRight(ratio.Mul(ratio, newBigInt("8131365268884726200")), 64)
	}
	if (tick & 32768) != 0 {
		ratio = signedShiftRight(ratio.Mul(ratio, newBigInt("3584323654723342297")), 64)
	}
	if (tick & 65536) != 0 {
		ratio = signedShiftRight(ratio.Mul(ratio, newBigInt("696457651847595233")), 64)
	}
	if (tick & 131072) != 0 {
		ratio = signedShiftRight(ratio.Mul(ratio, newBigInt("26294789957452057")), 64)
	}
	if (tick & 262144) != 0 {
		ratio = signedShiftRight(ratio.Mul(ratio, newBigInt("37481735321082")), 64)
	}

	return ratio
}

func signedShiftRight(ratio *big.Int, n int) *big.Int {
	return ratio.Rsh(ratio, uint(n))
}

func newBigInt(i string) *big.Int {
	count := new(big.Int)
	b, _ := count.SetString(i, 10)
	return b
}
