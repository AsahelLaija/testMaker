	// first obtain "Point" moment
	Ma := 0.0
	Mb1 := 0.0417 * math.Pow(l, 3) * k
	Mb2 := 0.125 * l * math.Pow(k, 3)
	Mb3 := 0.125 * l * math.Pow(j, 3)
	Mb4 := 0.0833 * math.Pow(k, 4)
	Mb5 := 0.1667 * k * math.Pow(j, 3)
	Mb6 := (l * k + l * j + math.Pow(k, 2) + 1.333 * k *j)
	Mb := RoundTo(W *((Mb1 - Mb2 - Mb3 - Mb4 - Mb5) / Mb6), 3)

	// then obtain "Mid" moment
	Mm := RoundTo(moment(j, Ma, Mb, W), 3)

	Mc1 := -2 * Mb * k
	Mc2 := 2 * Mb * j
	Mc3 := 0.25 * math.Pow(k, 3) * W
	Mc4 := 0.25 * math.Pow(j, 3) * W
	Mc := RoundTo((Mc1 - Mc2 - Mc3 - Mc4) / k, 3)

	Mf := RoundTo(moment(k, Mb, Mc, W), 3)

	Md1 := -0.5 * Mc
	Md2 := 0.125 * math.Pow(l, 2) * W
	Md3 := (W * math.Pow(l, 2)) / 12
	Md := RoundTo(((Md1 - Md2) - Md3) / 2, 3)

	Mg := RoundTo(moment(l, Mc, Md, W), 3)

	jo.Tp = make([]Panel, topPanels)
	for i := 0; i < topPanels; i++ {
		pan := &jo.Tp[i]
		pan.Name = "TP " + strconv.Itoa(i +1)
		if i == 0 {
			w2AxiFor	:= jo.WebMem[i].AxiFor
			pan.EndPoint	= RoundTo(j, 3)
			pan.Moment	= W *j /2 *(L -j)
			compTP1		:= CalcTP1Comp(w2AxiFor, A.BethaRelX)
			pan.Compresion	= compTP1
			pan.Mid		= Mm
			pan.Point	= Mb
		} else if i == 1 {
			svForMem	:= jo.WebMem[i].AxiFor
			endPoint	:= jo.Tp[i-1].EndPoint + k
			moment		:= W *endPoint /2 *(L -endPoint)
			prevComp	:= jo.Tp[i-1].Compresion
			pan.EndPoint	= endPoint
			pan.Moment	= moment
			pan.Mid		= Mf
			pan.Point	= Mc
			pan.Compresion	= prevComp -(svForMem *A.GammaRelX)
		} else if i == 2 {
			pCmp := jo.Tp[i-1].Compresion
			pAF := jo.WebMem[i].AxiFor
			cAF := jo.WebMem[i+1].AxiFor
			AF := CalcTPComp(pCmp, cAF, pAF, A.DeltaRelX, A.AlphaRelX)
			endPoint := jo.Tp[i-1].EndPoint + inPanel /2
			moment := W *endPoint /2 *(L -endPoint)
			pan.EndPoint	= endPoint
			pan.Moment	= moment
			pan.Compresion	= AF
			pan.Mid		= Mg
			pan.Point	= Md
		} else if i == 3{
			pAF := jo.Tp[i -2].Compresion
			pFM := jo.WebMem[i -1].AxiFor
			cFM := jo.WebMem[i].AxiFor
			AF := CalcTPComp(pAF, cFM, pFM, A.DeltaRelX, A.AlphaRelX)
			endPoint := jo.Tp[i-1].EndPoint + inPanel /2
			moment := W *endPoint /2 *(L -endPoint)

			pan.EndPoint	= endPoint
			pan.Moment	= moment
			pan.Compresion	= AF
			pan.Mid		= RoundTo(W *math.Pow(l, 2) /24, 3)
			pan.Point	= RoundTo(W *math.Pow(l, 2) /-12, 3)
		} else if i > 3 && i < topPanels -3{
			endPoint := jo.Tp[i-1].EndPoint + inPanel /2
			moment := W *endPoint /2 *(L -endPoint)
			pan.EndPoint	= endPoint
			pan.Moment	= moment
			pan.Mid		= RoundTo(W *math.Pow(l, 2) /24, 3)
			pan.Point	= RoundTo(W *math.Pow(l, 2) /-12, 3)
			if i %2 == 0 {
				pAF := jo.Tp[i -1].Compresion
				pFM := jo.WebMem[i].AxiFor
				cFM := jo.WebMem[i +1].AxiFor
				AF := CalcTPComp(pAF, cFM, pFM, A.AlphaRelX, A.AlphaRelX)
				pan.Compresion = AF
			} else if i %2 != 0{
				pan.Compresion = jo.Tp[i -1].Compresion
			}
		} else if i == topPanels -3 {
			pAF := jo.Tp[i -2].Compresion
			pFM := jo.WebMem[i -1].AxiFor
			cFM := jo.WebMem[i].AxiFor
			AF := CalcTPComp(pAF, cFM, pFM, A.AlphaRelX, A.AlphaRelX)
			endPoint := jo.Tp[i-1].EndPoint +inPanel /2
			moment := W *endPoint /2 *(L -endPoint)

			pan.EndPoint	= endPoint
			pan.Moment	= RoundTo(moment, 3)
			pan.Compresion	= AF
			pan.Mid		= RoundTo(Mg, 3)
			pan.Point	= RoundTo(Mc, 3)
		} else if i == topPanels -2 {
			pAF := jo.Tp[i-1].Compresion
			pFM := jo.WebMem[i].AxiFor
			cFM := jo.WebMem[i +1].AxiFor
			AF := CalcTPComp(pAF, cFM, pFM, A.AlphaRelX, A.AlphaRelX)
			endPoint := jo.Tp[i-1].EndPoint + k
			moment := W *endPoint /2 *(L -endPoint)

			pan.EndPoint	= endPoint
			pan.Moment	= RoundTo(moment, 3)
			pan.Compresion	= AF
			pan.Mid		= RoundTo(Mf, 3)
			pan.Point	= RoundTo(Mb, 3)
		} else if i == topPanels -1 {
			pAF := jo.Tp[i -1].Compresion
			pFM := jo.WebMem[len(jo.WebMem) -2].AxiFor
			moment := W *pan.EndPoint /2 *(L -pan.EndPoint)

			pan.EndPoint = RoundTo(jo.Tp[i-1].EndPoint + j, 3)
			pan.Moment	= RoundTo(moment, 3)
			pan.Compresion	= pAF +(pFM *A.GammaRelX)
			pan.Mid		= RoundTo(Mm, 3)
			pan.Point	= 0
		}
	}
