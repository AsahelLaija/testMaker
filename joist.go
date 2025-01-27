package main 
import (
	"encoding/csv"
	"os"
	"fmt"
	"math"
	"strconv"
)

type JoistBuilder interface {
	setJoistType()
	getJoist()	Joist
}

func getBuilder(builderType string) JoistBuilder {
	if builderType == "normal" {
		return newJoist()
	}
	return nil
}

type Joist struct {
	JoistType	string
	JoistName	string
	Geom		Geom
	Load		Load
	Analysis	Analysis
	Design		Design
}

func newJoist() *Joist {
	return &Joist{}
}

// b means builder
func (b *Joist) setJoistType() {
	b.JoistType = "Normal Joist"
}

func (joist *Joist) setJoistName(name string) {
	joist.JoistName = name
}
//	nGeom is the New Geometry method
//	sGeom is the Set Geometry method
//	sGeom initialize the variables given by the user in the Geom Struct
func(joist *Joist) sGeom(tt string, s, d, tf, np, ts, i, bf float64) {
	j := &joist.Geom
	j.TrussType = tt
	j.Span = s
	j.Depth = d
	j.Tfep = RoundTo(tf, 3)
	j.Tip = np
	j.Tsep = RoundTo(ts, 3)
	j.Bfep = bf
	j.Ip = i
}

func(joist *Joist) nGeom() {
	j := &joist.Geom
	j.Bsep = (j.Tfep +j.Tsep +(j.Ip /2)) -j.Bfep
	j.NumBip = (j.Tip+4-6)/2
	j.LenBip = j.Ip
	j.L = RoundTo((j.Span * 12) -4, 3)
	topPanL := RoundTo((j.L -(j.Tfep +j.Tsep) *2), 3)
	if j.TrussType == "warren" {
		j.Tip = topPanL / j.Ip
		j.Diag = (j.Tip *2) +4
	} else {
		jip := j.Ip /2
		j.Tip = topPanL / jip
		j.Diag = j.Tip +4
	}
	j.TopAng = setAngleChordProps(1)
	j.BotAng = setAngleChordProps(1)
	j.Ed = calculteED(j.Depth, j.TopAng.Y, j.BotAng.Y)
	j.Struts = j.Tip /2

	// calculate the angles
	j.WebAng.Betha = math.Atan(j.Ed / j.Bfep)
	bethaAdy := j.Bfep
	bethaHyp := math.Sqrt(bethaAdy *bethaAdy + j.Ed *j.Ed)
	j.WebAng.BethaRelX = bethaAdy /bethaHyp

	j.WebAng.BethaRelY = j.Ed /bethaHyp
	j.WebAng.Gamma = math.Atan(j.Ed / (j.Bfep - j.Tfep))
	gammaAdy := j.Bfep -j.Tfep
	gammaHyp := math.Sqrt(gammaAdy *gammaAdy + j.Ed *j.Ed)
	j.WebAng.GammaRelX = gammaAdy /gammaHyp
	j.WebAng.GammaRelY = j.Ed /gammaHyp

	j.WebAng.Delta = math.Atan(j.Ed / (j.Tfep + j.Tsep -j.Bfep))
	deltaAdy := (j.Tfep +j.Tsep) -j.Bfep
	deltaHyp := math.Sqrt(deltaAdy *deltaAdy + j.Ed *j.Ed)
	j.WebAng.DeltaRelX = deltaAdy /deltaHyp
	j.WebAng.DeltaRelY = j.Ed /deltaHyp

	//j.WebAng.Alpha = math.Atan(j.Ed / j.Ip /2)
	alphaAdy := j.Ip /2
	alphaHyp := math.Sqrt(alphaAdy *alphaAdy + j.Ed *j.Ed)
	j.WebAng.AlphaRelX = alphaAdy /alphaHyp
	j.WebAng.AlphaRelY = j.Ed /alphaHyp
}

func(joist *Joist) nLoad(wu, wll, ulL float64, ULsRaw, ULs []float64, s string) {
	j := &joist.Load
	j.WuPlf = wu
	j.WllPlf = wll
	j.WuKip = wu / 12000
	j.WllKip = wll / 12000
	j.UpLiftLoad = ulL
	j.UpLift = ULs
	j.UpLiftRaw = ULsRaw

	if s == "Y" {
		joist.Analysis.Stab = true
	} else {
		joist.Analysis.Stab = false
	}
}
func CalcCort(w, dsgnL, dist float64) float64 {
	return RoundTo(w *((dsgnL /2) -dist), 4)
}
func CalcMmnt(w, dsgnL, dist float64) float64 { return w*(dist /2) *(dsgnL -dist) }

func isTopPan1(i int) bool{ return i == 1}
func isTopPan2(i int) bool{ return i == 2}
func isIntPan(i, np int) bool{ return i > 2 && i < np -2}
func isLeftTopPan2(i, np int) bool { return i == np -2}
func isLeftTopPan1(i, np int) bool { return i == np -1}

func calcKnots(g *Geom, w float64) []Knot {
	nTP	:= int(g.Tip) +5
	ps	:= make([]Knot, nTP)
	np	:= len(ps)
	for i := 0; i < np; i++ {
		if isTopPan1(i) { ps[i].Dist = g.Tfep
		} else if isTopPan2(i) { ps[i].Dist = g.Tfep +g.Tsep
		} else if isIntPan(i, np) { ps[i].Dist = ps[i -1].Dist + g.Ip /2
		} else if isLeftTopPan2(i, np) { ps[i].Dist = ps[i -1].Dist + g.Tsep
		} else if isLeftTopPan1(i, np) { ps[i].Dist = ps[i -1].Dist + g.Tfep }
		ps[i].Cort = CalcCort(w, g.L, ps[i].Dist)
		ps[i].Mmnt = CalcMmnt(w, g.L, ps[i].Dist)
	}

	var nxtULV, prvULV, L, R	float64
	var Ls []float64
	var Rs []float64
	k := 0
	for j := np -1; j >= 0; j-- {
		if j == np -1 && k == 0{ 
			L = CalcULVxL(0, 0, 0, g.L, 0)
			nxtULV = L
			R = CalcULVxR(0, 0, 0, g.L, 0)
			prvULV = R
		} else {
			L = CalcULVxL(-0.01275, ps[j+1].Dist, ps[j].Dist, g.L, nxtULV)
			nxtULV = L
			R = CalcULVxR(-0.01275, ps[k].Dist, ps[k-1].Dist, g.L, prvULV)
			prvULV = R
		}
		Ls = append(Ls, L)
		Rs = append(Rs, R)
		k++
	}

	j := 0
	for i := len(Ls) -1; i >= 0; i--{
		ps[j].ULVx = Ls[i] +Rs[j]
		j++
	}
	return ps
}

func CalcULPanels(ap AngleProperties, ip, tsep, tfep float64, ulLoad []float64, pan Panels, knots TKnots, a WebAngles){
	svAxiFor := CalcULsvAxiFor(tsep, tfep, ulLoad[0], a.GammaRelY)
	for j := 1; j < len(knots); j++ {
		i := j -1
		if j == 1 {
			pan[i].ULAxiFor = CalcULAxiFor(knots[j].ULVx, knots[j -1].ULVx, a.BethaRelY)
		} else if j == 2 {
			pan[i].ULAxiFor = CalcULAxiFor(knots[j].ULVx, knots[j -1].ULVx, a.DeltaRelY)
		} else if j > 2 && j < len(knots) -2{
			pan[i].ULAxiFor = CalcULAxiFor(knots[j].ULVx, knots[j -1].ULVx, a.AlphaRelY)
		} else if j == len(knots) -2{
			pan[i].ULAxiFor = CalcULAxiFor(knots[j].ULVx, knots[j -1].ULVx, a.DeltaRelY)
		} else {
			pan[i].ULAxiFor = CalcULAxiFor(knots[j].ULVx, knots[j -1].ULVx, a.BethaRelY)
		}

		if i == 0 {
			pan[i].ULTens = CalcTensULTP1(pan[i].ULAxiFor, a.BethaRelX)
		} else if i == 1 {
			pan[i].ULTens = CalcTensULTP2(svAxiFor, a.GammaRelX, pan[i -1].ULTens)
		} else if i == 2 {
			pan[i].ULTens = CalcTensUL(pan[i-1].ULTens, pan[i -1].ULAxiFor, a.DeltaRelX, pan[i].ULAxiFor, a.AlphaRelX)
		} else if i > 2 && i < len(pan) -2{
			if i % 2 == 0 {
				pan[i].ULTens = CalcTensUL(pan[i-1].ULTens, pan[i -1].ULAxiFor, a.AlphaRelX, pan[i].ULAxiFor, a.AlphaRelX)
			} else {
				pan[i].ULTens = pan[i -1].ULTens
			}
		} else if i == len(pan) -2{
			pan[i].ULTens = CalcTensUL(pan[i-1].ULTens, pan[i -1].ULAxiFor, a.AlphaRelX, pan[i].ULAxiFor, a.DeltaRelX)
		} else {

			pan[i].ULTens = CalcTensULTPlast(svAxiFor, a.GammaRelX, pan[i -1].ULTens)

		}

		if i == 0 {
			pan[i].ULBndPnt = CalcTP1ULBndPnt(ip/2, tsep, tfep, ulLoad[i])
			pan[i].ULBndMid = CalcTP1ULBndMid(tfep, ulLoad[i], pan[i].ULBndPnt)
		} else if i == 1 {
			pan[i].ULBndPnt = CalcTP2ULBndPnt(ulLoad[i], pan[i-1].ULBndPnt, tsep, tfep)
			pan[i].ULBndMid = CalcTP2ULBndMid(ulLoad[i], pan[i-1].ULBndPnt, pan[i].ULBndPnt, tsep)
		} else if i == 2 {
			pan[i].ULBndPnt = CalcTP3ULBndPnt(ulLoad[i], pan[i].EndPoint, pan[i-1].EndPoint)
			pan[i].ULBndMid = CalcTP3ULBndMid(ip/2, ulLoad[i], pan[i-1].ULBndPnt, pan[i].ULBndPnt)
		} else if i > 2 && i < len(pan) -3{
			pan[i].ULBndPnt = CalcTP3ULBndPnt(ulLoad[i], pan[i].EndPoint, pan[i-1].EndPoint)
			pan[i].ULBndMid = CalcTP4ULBndMid(ulLoad[i], pan[i].EndPoint, pan[i-1].EndPoint)
		} else if i == len(pan) -3 {
			pan[i].ULBndPnt = CalcTP2ULBndPnt(ulLoad[i], pan[0].ULBndPnt, tsep, tfep)
			pan[i].ULBndMid = CalcTP3ULBndMid(ip/2, ulLoad[i], pan[1].ULBndPnt, pan[2].ULBndPnt)
		} else if i == len(pan) -2 {
			pan[i].ULBndPnt = CalcTP1ULBndPnt(ip/2, tsep, tfep, ulLoad[i])
			pan[i].ULBndMid = CalcTP2ULBndMid(ulLoad[i], pan[i-1].ULBndPnt, pan[i].ULBndPnt, tsep)
		} else {
			pan[i].ULBndMid = CalcTP1ULBndMid(tfep, ulLoad[i], pan[i -1].ULBndPnt)
		}

		pan[i].ULFbuMid = CalcFbuUL(pan[i].ULBndMid, ap.Y, ap.Ix, ap.B)
		pan[i].ULFbuPnt = CalcFbuUL(pan[i].ULBndPnt, ap.Y, ap.Ix, ap.B)
	}
}
func CalcULDesign(pan Panels) {

}

func(joist *Joist) nAnalysis() {
	uLLoads := &joist.Load.UpLift
	jtt := &joist.Geom.TrussType
	geom := &joist.Geom
	p := &joist.Analysis.TKnots
	jo := &joist.Analysis
	diag := &joist.Geom.Diag
	ttip := &joist.Geom.Tip		// num of interior Panels
	topPanels := int(*ttip +4)	// num of top interior panels

	A := joist.Geom.WebAng
	L := joist.Geom.L
	W := joist.Load.WuKip
	j := joist.Geom.Tfep
	k := joist.Geom.Tsep
	inPanel := joist.Geom.Ip
	bfep := joist.Geom.Bfep
	
	*p = calcKnots(geom, W)

	var l float64
	var bip float64
	if *jtt == "warren" {
		l = joist.Geom.Ip /2
		bip = *ttip +3	// num of bot interior panels
	} else if *jtt == "warrenModified"{
		l = joist.Geom.Ip /2
		bip = ((*ttip +2) /2) +2 // num of bot interior panels
	}

	svStart := j
	jo.WebMem = make([]Mem, int(*diag) +2)
	svV := W *(L /2 - svStart)
	for m := 0; m < len(jo.WebMem); m++ {
		mem := &jo.WebMem[m]
		if m == 0 {
			mem.Name = "w2"
			mem.Start = 0
			mem.End = bfep
			M := W *(bfep /2) *(L -bfep)
			mem.Moment = M
			V := RoundTo(W *L /2, 8)
			cortAtZero := V
			mem.Cort = V
			axiFor := ((cortAtZero +svV) /2) /A.BethaRelY
			mem.AxiFor = axiFor
		} else if m == 1 {
			mem.Name = "sv"
			mem.Start = RoundTo(svStart, 3)
			mem.End = bfep
			M := W *(svStart /2) *(L -svStart)
			mem.Moment = M
			mem.Cort = svV
			mem.AxiFor = ((j/2 +k/2) *W) /A.GammaRelY
		} else if m == 2 {
			mem.Name = "w3"
			mem.End = RoundTo(j +k, 3)
			M := W *(mem.End /2) *(L -mem.End)
			mem.Moment = M
			V := W *(L /2 -mem.End)
			mem.Cort = V
			mem.AxiFor = ((svV + V) /2) /A.DeltaRelY
		} else if m > 1 &&  m < len(jo.WebMem) -3{
			mem.Name = "w" +strconv.Itoa(m +1)
			mem.Start = jo.WebMem[m-1].End
			mem.End = jo.WebMem[m-1].End +l 
			M := W *(mem.End /2) *(L -mem.End)
			mem.Moment = M
			V := W *(L /2 -mem.End)
			mem.Cort = V
			pV := jo.WebMem[m-1].Cort
			mem.AxiFor = ((pV + V) /2) /A.AlphaRelY
		} else if m == len(jo.WebMem) -3 {
			mem.Name = "w" +strconv.Itoa(m +1)
			mem.Start = jo.WebMem[m-1].End
			mem.End = mem.Start +k
			M := W *(mem.End /2) *(L -mem.End)
			mem.Moment = M
			V := W *(L /2 -mem.End)
			mem.Cort = V
			pV := jo.WebMem[m-1].Cort
			mem.AxiFor = ((pV + V) /2) /A.DeltaRelY
		} else if m == len(jo.WebMem) -2 {
			mem.Name = "sv"
			mem.Start = jo.WebMem[m-1].End
			mem.End = mem.Start +l
			mem.AxiFor = ((j/2 +k/2) *W) /A.GammaRelY
		} else if m == len(jo.WebMem) -1 {
			mem.Name = "w" +strconv.Itoa(m)
			mem.Start = jo.WebMem[m-2].End
			mem.End = mem.Start +j
			M := W *(mem.End /2) *(L -mem.End)
			mem.Moment = M
			V := W *(L /2 -mem.End)
			mem.Cort = V
			pV := jo.WebMem[m-2].Cort
			mem.AxiFor = ((pV +V) /2) /A.BethaRelY
		}
	}

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
	CalcULPanels(geom.TopAng, geom.Ip, geom.Tfep, geom.Tsep, *uLLoads, jo.Tp, *p, geom.WebAng)
	BotPanel := make([]Panel, int(bip))
	jo.Bp = BotPanel
	w2AF := jo.WebMem[0].AxiFor
	w3AF := jo.WebMem[2].AxiFor
	svAF := jo.WebMem[1].AxiFor
	var bb int
	for b := 0; b < int(bip); b++ {
		jo.Bp[b].Name = "BP " + strconv.Itoa(b +1)
		if b == 0{
			ten1 := w2AF *A.BethaRelX
			ten2 := w3AF *A.DeltaRelY
			ten3 := svAF *A.GammaRelX
			ten := ten1 + ten2 - ten3
			jo.Bp[b].Compresion = ten
			bb = b + 3
		} else {
			pTen := jo.Bp[b-1].Compresion
			w4AF := jo.WebMem[bb].AxiFor
			w5AF := jo.WebMem[bb].AxiFor
			ten := pTen +w4AF * A.AlphaRelX +w5AF * A.AlphaRelX
			jo.Bp[b].Compresion = ten
			bb += 2
		}
	}
}

func CalcTP1Comp(w2AxiFor, betaRelx float64) float64{ return w2AxiFor *betaRelx; }
func CalcTP2Comp(pComp, svAF, gammaRelX float64) float64 {
	return pComp -svAF *gammaRelX
}
func CalcTPComp(pComp, cAF, pAF, cRelX, pRelX float64) float64 {
	return  pComp +cAF *cRelX +pAF *pRelX
}

func(joist *Joist) nTopDesign() {
	span	:= &joist.Geom.Span
	tfep	:= &joist.Geom.Tfep
	ipl	:= &joist.Geom.Ip
	dsgnL	:= &joist.Geom.L
	angl	:= &joist.Geom.TopAng
	an	:= joist.Analysis.Tp
	Knots	:= joist.Analysis.TKnots
	joist.Design.TopChord = make([]Chord, len(an))
	dsgn	:= &joist.Design
	chrd	:= joist.Design.TopChord
	rx	:= angl.Rx
	ry	:= angl.Ry
	rz	:= angl.Rz
	dsgn.MaxLrxEP = CalcTPMaxLrx(*tfep, rx)
	dsgn.MaxLrxIP = CalcTPMaxLrx(*ipl /2, rx)
	dsgn.LrzEP = CalcTPLrz(false, *tfep, rz, )
	dsgn.LrzIP = CalcTPLrz(false, *ipl /2, rz, )
	dsgn.KLrzkLsrzEP = CalcTPkLrzkLsrz(false, EPkz, IPksz, dsgn.LrzEP)
	dsgn.KLrzkLsrzIP = CalcTPkLrzkLsrz(false, IPkz, IPksz, dsgn.LrzIP)
	klrzs	:= [...]float64{0, dsgn.KLrzkLsrzEP, dsgn.KLrzkLsrzIP}
	brid	:= CalcTopBrid(klrzs, *dsgnL, ry)
	spacing := CalcSpacing(*span, brid)
	dsgn.Lry = CalcTPLry(false, spacing, ry)
	dsgn.MaxkLrxEP = CalcTPMaxkLrx(EPkx, dsgn.MaxLrxEP)
	dsgn.MaxkLrxIP = CalcTPMaxkLrx(IPkx, dsgn.MaxLrxIP)
	dsgn.KLryEP = CalcTPkLry(ky, dsgn.Lry)
	dsgn.KLryIP = CalcTPkLry(ky, dsgn.Lry)
	dsgn.MaxkLrEP = CalcMaxkLr(dsgn.MaxkLrxEP, dsgn.KLryEP, dsgn.KLrzkLsrzEP)
	dsgn.MaxkLrIP = CalcMaxkLr(dsgn.MaxkLrxIP, dsgn.KLryIP, dsgn.KLrzkLsrzIP)
	Cond := 4.71 *math.Sqrt(E /(angl.Q *Fy))
	var ratio float64
	var Mark int = 1
	for !dsgn.Pass {
		for j := 0; j < len(Knots); j++ {
			Knots[j].Fcr = Fy *0.9
			if j != 0 {
				pnt := an[j -1].Point
				Knots[j].Fbu = CalcTPFbu(pnt, angl.Y, angl.Ix, angl.B)
			}
		}
		for i := 0; i < len(chrd); i++ {
			panEP	:= an[i].EndPoint
			comp	:= an[i].Compresion
			mid	:= an[i].Mid
			var panSP float64 = 0
			if i != 0 {panSP = an[i -1].EndPoint;}
			if i < 2 {
				chrd[i].Kl_rx = CalcTPKl_rx(EPkx, panEP, panSP, rx)
				chrd[i].MaxkLr = dsgn.MaxkLrEP
				chrd[i].Fex = CalcTPFex(E, chrd[i].Kl_rx)
				chrd[i].Fez = CalcTPFez(E, chrd[i].MaxkLr)
				chrd[i].Fcr_x = CalcTPFcr(chrd[i].Kl_rx, Cond, angl.Q, Fy, chrd[i].Fex)
				chrd[i].Fcr_z = CalcTPFcr(chrd[i].MaxkLr, Cond, angl.Q, Fy, chrd[i].Fez)
				chrd[i].Fcr = CalcTPFc(chrd[i].Fcr_z, 0.9)
				chrd[i].Fau = CalcTPFau(comp, angl.Area)
				chrd[i].Cm = CalcCm(0.3, chrd[i].Fau, comp, chrd[i].Fez)
			} else if i < len(chrd) -2{
				chrd[i].Kl_rx = CalcTPKl_rx(IPkx, panEP, panSP, rx)
				chrd[i].MaxkLr = dsgn.MaxkLrIP
				chrd[i].Fex = CalcTPFex(E, chrd[i].Kl_rx)
				chrd[i].Fez = CalcTPFez(E, chrd[i].MaxkLr)
				chrd[i].Fcr_x = CalcTPFcr(chrd[i].Kl_rx, Cond, angl.Q, Fy, chrd[i].Fex)
				chrd[i].Fcr_z = CalcTPFcr(chrd[i].MaxkLr, Cond, angl.Q, Fy, chrd[i].Fez)
				chrd[i].Fcr = CalcTPFc(chrd[i].Fcr_z, 0.9)
				chrd[i].Fau = CalcTPFau(comp, angl.Area)
				chrd[i].Cm = CalcCm(0.4, chrd[i].Fau, comp, chrd[i].Fez)
			} else {
				chrd[i].Kl_rx = CalcTPKl_rx(EPkx, panEP, panSP, rx)
				chrd[i].MaxkLr = dsgn.MaxkLrEP
				chrd[i].Fex = CalcTPFex(E, chrd[i].Kl_rx)
				chrd[i].Fez = CalcTPFez(E, chrd[i].MaxkLr)
				chrd[i].Fcr_x = CalcTPFcr(chrd[i].Kl_rx, Cond, angl.Q, Fy, chrd[i].Fex)
				chrd[i].Fcr_z = CalcTPFcr(chrd[i].MaxkLr, Cond, angl.Q, Fy, chrd[i].Fez)
				chrd[i].Fcr = CalcTPFc(chrd[i].Fcr_z, 0.9)
				chrd[i].Fau = CalcTPFau(comp, angl.Area)
				chrd[i].Cm = CalcCm(0.3, chrd[i].Fau, comp, chrd[i].Fez)
			}
			chrd[i].Fbu = CalcTPFbu(mid, angl.Y, angl.Ix, angl.B)
			chrd[i].FauFc = CalcFauFc(chrd[i].Fau, chrd[i].Fcr)

			ratio = Calcratio(chrd[i].FauFc, chrd[i].Cm, chrd[i].Fbu, chrd[i].Fau, chrd[i].Fex, Cmp, angl.Q, Bnd, Fy)
			chrd[i].ratio = ratio

			chrd[i].Unity = unityPanTP(
				chrd[i].FauFc,
				chrd[i].Cm,
				chrd[i].Fbu,
				chrd[i].Fau,
				0.9,
				chrd[i].Fex,
				angl.Q,
				0.9,
				Fy,
			)

			Knots[i+1].Unity = unityPntTP(
				chrd[i].FauFc,
				0.9,
				Fy,
				Knots[i+1].Fbu,
			)
			chrd[i].Stress = an[i].ULTens / angl.Area
			chrd[i].FauPuA = chrd[i].Stress / angl.Area
			chrd[i].BndUnity = calcBndUnity(
				chrd[i].FauPuA,
				0.9,
				Fy,
				chrd[i].Fbu,
				0.9,
			)
			if ratio > 1.0 {
				dsgn.Pass = false
//				fmt.Println("NOT PASS", ratio)
				continue
			} else { 
				dsgn.Pass = true
//				fmt.Println("PASS", ratio)
			}
		}
		if dsgn.Pass { break }
		*angl = setAngleChordProps(Mark)
		Mark++
	}
	//fmt.Println(Mark -1)
	//fmt.Println(dsgn.Pass)
	//fmt.Println(joist.Geom.TopAng.Mark)
}

func calcBndUnity (R111, I9, Q91, Q111, M9 float64) float64 {
	return math.Abs(R111/(I9*Q91))+Q111/(M9*Q91)
}

func unityPntTP(I112, K9, Q91, K112 float64) float64 {
	return I112/(K9*Q91)+K112/(K9*Q91)
}

func unityPanTP(FauFc, Cm, Fbu, Fau, K9, E111, S90, M9, Q91 float64) float64{
	if FauFc>=0.2 {
		return FauFc+8/9*Cm*Fbu/((1-Fau/(K9*E111))*S90*M9*Q91)
	} else {
		return FauFc/2+Cm*Fbu/(((1-Fau/(K9*E111))*S90*M9*Q91))
	}
}


func(joist *Joist) nBotSlend() {
	geom := &joist.Geom
	aPrp := &joist.Geom.BotAng
	bSlnd := &joist.Design.BotSlend
	// create spacing on bridging place
	spacing := 296.5
	EPFiller := 0
	IPFiller := 0

	bSlnd.MaxLrxEP = geom.Bsep /aPrp.Rx
	bSlnd.MaxLrxIP = geom.LenBip /aPrp.Rx

	// LryEP and LryIP should be same variable
	bSlnd.LryEP = spacing /aPrp.Ry
	bSlnd.LryIP = spacing /aPrp.Ry

	if EPFiller == 0{
		bSlnd.LrzLsrzEP = geom.Bsep /aPrp.Rz
	} else {
		bSlnd.LrzLsrzEP = geom.Bsep /(2*aPrp.Rz)
	}

	if IPFiller == 0{
		bSlnd.LrzLsrzIP = geom.LenBip /aPrp.Rz
	} else {
		bSlnd.LrzLsrzIP = geom.LenBip /(2*aPrp.Rz)
	}

	bSlnd.MaxkLrxEP = bSlnd.MaxLrxEP * .9
	bSlnd.MaxkLrxIP = bSlnd.MaxLrxIP * .9

	// KLryEP and KLryEP should be same variable
	bSlnd.KLryEP = bSlnd.LryEP *0.94
	bSlnd.KLryIP = bSlnd.LryIP *0.94

	if EPFiller == 0 {
		bSlnd.KLrzkLsrzEP = bSlnd.LrzLsrzEP *0.9
	} else {
		bSlnd.KLrzkLsrzEP = bSlnd.LrzLsrzEP *1
	}

	if IPFiller == 0 {
		bSlnd.KLrzkLsrzIP = bSlnd.LrzLsrzIP *0.9
	} else {
		bSlnd.KLrzkLsrzIP = bSlnd.LrzLsrzIP *1
	}
	bSlnd.MaxkLrEP = larger(
		bSlnd.MaxkLrxEP,
		bSlnd.KLryEP,
		bSlnd.KLrzkLsrzEP,
	)
	bSlnd.MaxkLrIP = larger(
		bSlnd.MaxkLrxIP,
		bSlnd.KLryIP,
		bSlnd.KLrzkLsrzIP,
	)
}

func(joist *Joist) nBotDesing(){
	g := &joist.Geom
	s := &joist.Design.BotSlend
	d := &joist.Design.BotChord
	a := &joist.Analysis.BKnots
	ta := joist.Analysis.Tp
	aPrp := &joist.Geom.BotAng
	
	numBP := int(g.NumBip +4)

	// -0.01275 ulLoad
	svAxiFor := CalcULsvAxiFor(g.Tsep, g.Tfep, -0.01275, g.WebAng.GammaRelY)

	joist.Design.BotChord = make([]Chord, numBP)
	joist.Analysis.BKnots = make([]Knot, numBP -1)

	FcrxCond := CalcFcrxCond(aPrp.Q, Fy)

	bKnots := joist.Analysis.BKnots
	for i := 0; i < len(*a); i++{
		tpInd := i *2
		if i == 0 {
			bKnots[i].Dist = g.Bfep
			bKnots[i].Compresion = CalcULCompEP(
				ta[0].ULAxiFor,
				g.WebAng.BethaRelX,
				svAxiFor,
				g.WebAng.GammaRelX,
				ta[1].ULAxiFor,
				g.WebAng.DeltaRelX,
			)
		} else if i == 1 {
			bKnots[i].Dist = bKnots[i -1].Dist + g.Bsep
			bKnots[i].Compresion = CalcULCompIP(
				bKnots[i -1].Compresion,
				g.WebAng.AlphaRelX,
				ta[tpInd +1].ULAxiFor,
				ta[tpInd].ULAxiFor,
			)
		} else if i > 1 && i < numBP -2{
			bKnots[i].Dist = bKnots[i -1].Dist + g.LenBip
			bKnots[i].Compresion = CalcULCompIP(
				bKnots[i -1].Compresion,
				g.WebAng.AlphaRelX,
				ta[tpInd +1].ULAxiFor,
				ta[tpInd].ULAxiFor,
			)
		} else {
			bKnots[i].Dist = bKnots[i -1].Dist + g.Bsep
		}
		bKnots[i].FauFc = bKnots[i].Compresion /aPrp.Area
		// 0.9 is the compresion when is LRFD
		bKnots[i].Fcr = Fy *0.9
		bKnots[i].Fb = Fy *0.9
		bKnots[i].Unity = unityK(
			bKnots[i].FauFc,
			0.9,
			Fy,
		)
	}
	bPanels := joist.Design.BotChord
	j := 1
	for i := 0; i < len(*d) -2; i++{
		bPanels[i].Kl_rx = CalcKlrx(
			.9,
			bKnots[i +1].Dist,
			bKnots[i].Dist,
			aPrp.Rx,
		)
		bPanels[i].Fex = CalcFex(E, bPanels[i].Kl_rx)
		if i == 0 {
			bPanels[i].MaxkLr = s.MaxkLrEP
		} else if i > 0 && i < numBP -2 {
			bPanels[i].MaxkLr = s.MaxkLrIP
			bPanels[i].Comp = 0
		} else {
			bPanels[i].MaxkLr = s.MaxkLrEP
			bPanels[i].Comp = 0
		}
		bPanels[i].Fez = CalcFex(E, bPanels[i].MaxkLr)

		bKnots[j].FcrX = CalcFcrX(
			bPanels[i].Kl_rx, 
			FcrxCond,
			aPrp.Q,
			Fy,
			bPanels[i].Fex,
		)
		bPanels[i].Fcr_z = CalcFcrX(
			bPanels[i].MaxkLr, 
			FcrxCond,
			aPrp.Q,
			Fy,
			bPanels[i].Fez,
		)
		bPanels[i].Fau = bKnots[j].Dist /aPrp.Area
		bPanels[i].FauFc = bKnots[i].Compresion /aPrp.Area
		// 0.9 is the compresion when is LRFD
		bPanels[i].Fcr = bPanels[i].Fcr_z *0.9
		bPanels[i].Fb = Fy *0.9
		bPanels[i].FaFc = bPanels[i].FauFc /bPanels[i].Fcr
		if i == 0 {
			bPanels[i].Unity = unityEP(bPanels[i].FaFc)
		} else {
			bPanels[i].Unity = unityIP (
				bPanels[i].FaFc,
				0, 0,
				bPanels[i].FauFc,
				0.9,
				bPanels[i].Fex,
				aPrp.Q,
				0.9,
				Fy,
			)
		}
		j++
	}
}
func unityK (I196, K9, Q176  float64 ) float64 {
	return math.Abs(I196)/(K9*Q176)
}
func unityIP (FaFc, L195, K195, I195, K9, E195, S90, M9, Q91 float64) float64 {
	if FaFc>=0.2 {
		return FaFc+8/9*L195*K195/((1-I195/(K9*E195))*S90*M9*Q91) 
	} else { 
		return FaFc/2+L195*K195/(((1-I195/(K9*E195))*S90*M9*Q91)) 
	}
}

func unityEP( N193 float64 ) float64 {
	if N193>=0.2 {
		return N193
	} else {
		return N193/2
	}
}

// TP1ULAF -> TP1 Uplift Axial Force
// svULAF -> sv member Uplift Axial Force
// TP2ULAF -> TP1 Uplift Axial Force
func CalcULCompEP(TP1ULAF, bethaRelX, svULAF, gammaRelX, TP2ULAF, DeltaRelX float64 ) float64 {
        return TP1ULAF *bethaRelX -svULAF *gammaRelX +TP2ULAF *DeltaRelX
}

//nxtTPAF -> next TP Axial Force
//prvTPAF -> prev TP Axial Force
func CalcULCompIP(prvComp, alphaRelX, nxtTPAF, prvTPAF float64 ) float64 {
	return prvComp +prvTPAF *alphaRelX +nxtTPAF *alphaRelX
}

func CalcFcrxCond(q, fy float64) float64 {
        return 4.71*math.Sqrt(29000/(q*fy))
}

func CalcFcrX( C193, S182, S175, Q176, E193 float64 ) float64 {
	if C193 <= S182 {
		return S175*(math.Pow(0.658, (S175*Q176/E193)))*Q176
	} else {
		return 0.877 *E193
	}
}

func CalcFex(Q175, C193 float64) float64 {
        return (math.Pi *math.Pi) *Q175 /(C193 *C193)
}

func CalcKlrx(E179, Q32, Q29, K175 float64 ) float64 {
	return E179*(Q32-Q29)/K175
}

func CalcFbuUL( ulBnd, O24, K24, I23 float64 ) float64 {
	if ulBnd>0.01 {
		return ulBnd*O24/K24
	} else {
		return math.Abs(ulBnd*((I23-O24)/K24))
	}
}

func CalcTP4ULBndMid( ulLoad, crrDist, prvDist float64 ) float64 {
	//	also
	//		TP5ULBndMid
        return ulLoad *math.Pow((crrDist-prvDist), 2) /24
}

func CalcTP3ULBndPnt( ulLoad, crrDist, prvDist float64 ) float64 {
	//	also 
	//		TP4ULBndPnt
        return ulLoad *math.Pow((crrDist-prvDist), 2) /-12
}

func CalcTP3ULBndMid( intPanL, ulLoad, prvBndPnt, bndPnt float64 ) float64{
	intPanLPow4 := math.Pow(intPanL, 4)
	ulLoadPow2 := math.Pow(ulLoad, 2)
	intPanLPow2 := math.Pow(intPanL, 2)
	prvBndPntPow2 := math.Pow(prvBndPnt, 2)
	bndPntPow2 := math.Pow(bndPnt, 2)
        return (intPanLPow4* ulLoadPow2 +4 *intPanLPow2 *prvBndPnt *ulLoad +4 * intPanLPow2 *bndPnt*ulLoad+4* prvBndPntPow2 -8*prvBndPnt*bndPnt+4* bndPntPow2)/(8*intPanLPow2 *ulLoad)
}

func CalcTP2ULBndPnt( ulLoad, prvBndPnt, tsep, tfep float64 ) float64{
	tsepPow3 := math.Pow(tsep, 3)
	tfepPow3 := math.Pow(tfep, 3)
        return (-2*prvBndPnt*tsep - 2*prvBndPnt*tfep - 0.25*tsepPow3*ulLoad - 0.25*tfepPow3 *ulLoad) / tsep
}

func CalcTP2ULBndMid( ulLoad, prvBndPnt, bndPnt, tsep float64 ) float64{
	tsepPow4 := math.Pow(tsep, 4)
	ulLoadPow4 := math.Pow(ulLoad, 2)
	tsepPow2 := math.Pow(tsep, 2)
	prvBndPntPow2 := math.Pow(prvBndPnt, 2)
	bndPntPow2 := math.Pow(bndPnt, 2)

	dividend	:= tsepPow4 *ulLoadPow4 +4 *tsepPow2 *prvBndPnt *ulLoad +4 *tsepPow2 *bndPnt *ulLoad +4 *prvBndPntPow2 -8*prvBndPnt*bndPnt+4*bndPntPow2
	divisor		:= 8 *tsepPow2 *ulLoad
        return dividend /divisor
}

func CalcTP1ULBndPnt( intPanL, tsep, tfep, ulLoad float64 ) float64{
	intPanLPow3 := math.Pow(intPanL, 3)
	tsepPow3 := math.Pow(tsep, 3)
	tfepPow3 := math.Pow(tfep, 3)
	tsepPow4 := math.Pow(tsep, 4)
	tsepPow2 := math.Pow(tsep, 2)

	dividend := ulLoad *(0.04165 *intPanLPow3 *tsep -0.125 *intPanL *tsepPow3 -0.125 *intPanL *tfepPow3 -0.0833 *tsepPow4 -0.1667 *tsep *tfepPow3)
	divisor := intPanL *tsep +intPanL *tfep +tsepPow2 +1.3333 *tsep *tfep

        return  dividend /divisor
}

func CalcTP1ULBndMid(tfep, ulLoad, bndPnt float64) float64{
	cuarta := math.Pow(tfep, 4)
	return (cuarta* (ulLoad*ulLoad)+4*(tfep *tfep)*bndPnt*ulLoad+4*(bndPnt *bndPnt))/(8*(tfep *tfep)*ulLoad)
}

func CalcTensUL(prvULTns, prvAF, prvRelX, crrAF, crrRelX float64)float64{ 
	return prvULTns+prvAF*prvRelX+crrAF*crrRelX
}

func CalcTensULTP2( svAxiFor, gammaRelX, prvTns float64 ) float64 {
        return prvTns-svAxiFor*gammaRelX
}

func CalcTensULTPlast( svAxiFor, gammaRelX, prvTns float64 ) float64 {
        return prvTns+svAxiFor*gammaRelX
}

func CalcTensULTP1(O43, R15 float64 ) float64 {
	return O43 * R15
}

func CalcULsvAxiFor( tfep, tsep, ulLoad, gammaRelY float64 ) float64 {
        return (tfep+tsep)/2*ulLoad/gammaRelY
}

func CalcULAxiFor(crrULVx, prvULVx, RelY float64) float64{
	return (crrULVx+prvULVx)/2/RelY
}

func CalcCortUL(L42, M42 float64) float64 {
	return L42 +M42
}
func CalcULVxR(ulLoad, crrDist, prvDist, dsgnL, prvULV float64) float64 {
	return (ulLoad*(crrDist-prvDist)/(2*dsgnL))*(2*(dsgnL-crrDist)+(crrDist-prvDist))-(ulLoad*(crrDist-prvDist))+prvULV
}

func CalcULVxL(ulLoad, nxtDist, crrDist, dsgnL, nxtULVx float64) float64 {
	return (ulLoad*(nxtDist-crrDist)/(2*dsgnL))*(2*(dsgnL-nxtDist)+(nxtDist-crrDist))+nxtULVx
}

func CalcUplift(load1, load2 float64) float64 {
	return (load1 +load2) /12000
}

func Calcratio(fauFc, cm, fbu, fau, fex, cmp, q, bnd, fy float64)float64 {
	if fauFc >= 0.2{
		ratio := fauFc+ ((8.0/9.0*(cm*fbu)) /((1.0-(fau/(cmp*fex)))*q*bnd*fy))
		return ratio
	} else {
		return fauFc/2.0+cm*fbu/(((1.0-fau/(cmp*fex))*q*bnd*fy))
	}
}

func CalcFauFc(fau, fc float64) float64{ return fau /fc }

func CalcCm(dMCond, fau, cmp, fex float64) float64 {
	return 1 -dMCond *fau /(cmp *fex)
}

func CalcTPFc(fcrz, bend float64) float64{ return fcrz *bend }

func CalcTPFbu(mid, y, ix, b float64) float64{
	if mid > 0.01 { return mid *y /ix
	} else { return math.Abs(mid *((b -y) /ix)) }
}

func CalcTPFau(comp, area float64) float64 { return comp /area }

func CalcTPFcr(klrx, cond, Q, Fy, fex float64) float64 {
	if klrx <= cond { return Q *(math.Pow(0.658 ,(Q *Fy /fex))) *Fy
	} else { return 0.877 *fex }
}

func CalcTPFez(e, maxkLr float64) float64 {
	return (math.Pi *math.Pi) *e /(maxkLr *maxkLr)
}

func CalcCheck(maxkLr, maxLsr float64) bool{ return maxkLr < maxLsr; }

func CalcMaxkLr(maxkLrx, kLry, kLrzkLsrz float64) float64{
	return larger(maxkLrx, kLry, kLrzkLsrz)
}

func CalcTopBrid(klrz [3]float64, dsgnL, ry float64) float64{
	return math.Ceil(dsgnL /maximum(klrz) *ry/0.94) -1
}

func CalcSpacing(span, bridRows float64) float64{ return span /(bridRows +1) *12 } 

func CalcTPkLrzkLsrz(filler bool, kz, ksz, Lrz float64) float64{
	if filler { return Lrz *ksz
	} else { return Lrz *kz }
}

func CalcTPkLry(ky, Lry float64) float64{ return ky *Lry; }

func CalcTPMaxkLrx(kx, maxLrx float64) float64{ return kx *maxLrx; }

func CalcTPLrz(filler bool, panLength, rz float64) float64{
	if filler { return panLength /(2 *rz)
	} else { return panLength /rz }
}

func CalcTPLry(stability bool, spacing, ry float64) float64 {
	if stability { return 36.0 /ry
	} else { return spacing /ry }
}

func CalcTPMaxLrx(tp, rx float64) float64 { return tp /rx; }

func CalcTPKl_rx(kx, panEP, panSP, rx float64) float64{ return kx *(panEP -panSP) /rx }

func CalcTPFex(E, currKlrx float64) float64 {
	return (math.Pi *math.Pi) *E /(currKlrx *currKlrx)
}


func setAngleChordProps(angleMark int) AngleProperties{
	var ap AngleProperties
	angle := angleProp(angleMark)
	 
	tmpMark,_	:= strconv.ParseFloat(angle[0], 64)
	ap.Mark		= int(tmpMark)
	ap.Section	= angle[1]
	ap.Area,_	= strconv.ParseFloat(angle[13], 64)
	ap.Rx,_	= strconv.ParseFloat(angle[14], 64)
	ap.Ry,_	= strconv.ParseFloat(angle[15], 64)
	ap.Rz,_	= strconv.ParseFloat(angle[16], 64)
	ap.Ix,_	= strconv.ParseFloat(angle[17], 64)
	ap.Iy,_	= strconv.ParseFloat(angle[18], 64)
	ap.Y, _	= strconv.ParseFloat(angle[9], 64)
	ap.B, _	= strconv.ParseFloat(angle[4], 64)
	ap.T, _	= strconv.ParseFloat(angle[3], 64)
	ap.Q, _	= strconv.ParseFloat(angle[11], 64)
	return ap
}

func (b *Joist) getJoist() Joist {
	return Joist{
		JoistType: b.JoistType,
		JoistName: b.JoistName,
	}
}

type Director struct {
	builder JoistBuilder
}
func newDirector(b JoistBuilder) *Director {
	return &Director{
		builder:	b,
	}
}

func (d *Director) buildJoist() Joist {
	d.builder.setJoistType()
	return d.builder.getJoist()
}

// logic Code
func RoundTo(number float64, decimals uint32) float64 {
	ratio := math.Pow(10, float64(decimals))
	return math.Round(number *ratio) / ratio
}

//	doubleAngle lives
func doubleAngle(area, ix, y string, sbca float64) (string, string, string) {
	areaStr,_  := strconv.ParseFloat(area, 64)
	areaFlt := 2*areaStr
	areaDAngle := strconv.FormatFloat(areaFlt, 'g', -1, 64)
	ixStr,_ := strconv.ParseFloat(ix, 64)
	yStr,_ := strconv.ParseFloat(y, 32)
	ixFlt := 2*ixStr
	ixDAngle := strconv.FormatFloat(ixFlt, 'g', -1, 64)
	result := (sbca/2) + yStr

	// res is egual to ((sbca/2)+Ytc)^2
	res := math.Pow(result, 2)
	iyFlt := 2*(ixStr + areaStr*res)
	iyDAngle := strconv.FormatFloat(iyFlt, 'g', -1, 64)
	return areaDAngle, ixDAngle, iyDAngle
}
// leave for the posterity
func maximum (array[3]float64) float64{
	var holder float64
	for _, num := range array{
		if num > holder {
			holder = num
		}
	}
	return holder
}
func larger (array ...float64) float64{
	var holder float64
	for _, num := range array{
		if num > holder {
			holder = num
		}
	}
	return holder
}

func angleProp(row int) []string {
	file, err := os.Open("Propiedades.csv")
	if err != nil {
		fmt.Println("Error: ", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	record, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	return record[row]
}

func calculteED(depth, Ytc, Ybc float64) float64{ return (depth -Ytc) -Ybc }
func Atof(a string) float64{
	A,_ := strconv.ParseFloat(a, 64)
	return A
}
