package main
import (
	"math"
)

type Geom struct {
	TrussType	string
	Span,	// span
	Depth,	// Total depth
	Tfep,	// First end Panel Top
	Tsep,	// Second end Panel Top
	Ip,	// Interior Panel Length
	Bfep,	// First end Panel Bot
	Bsep,	// Second end Panel Bot 
	NumBip,
	LenBip,
	L,	// Design Length span - 4in
	Tip,	// Total interior Panel
	Diag,	// Total diagonals
	Struts,	// Total verticals
	Ed	float64	// Efective depth
	WebAng	WebAngles
	TopAng	AngleProperties
	BotAng	AngleProperties
}

type WebAngles struct {
	Betha, BethaRelX, BethaRelY,
	Gamma, GammaRelX, GammaRelY,
	Delta, DeltaRelX, DeltaRelY,
	Alpha, AlphaRelX, AlphaRelY	float64
}

type AngleProperties struct {
	Section		string
	Mark		int
	MaxKlr, 
	Rx, Ry, Rz,
	Ix, Iy, Y, 
	B, T, Q, Area	float64
}

type Load struct {
	WuPlf,	// Ultimate design load
	WuKip,	// Ultimate design load
	WllPlf,	// Live Load
	WllKip	float64	
}

type Analysis struct {
	Stab	bool
	WebMem	Wms
	Tp	Panels	// Top Panels
	Bp	Panels	// Bot Panels
	TKnots	TKnots
	BKnots	BKnots
}

type Panels	= []Panel
type Wms	= []Mem
type TKnots	= []Knot
type BKnots	= []Knot

type Mem struct {
	Name	string
	Start,
	End,
	Cort,
	Moment,
	AxiFor	float64
}

type Panel struct {
	Name	string
	EndPoint,
	Suct,		// Suction
	Compresion,	// Axial Force
	AxiFor,
	Moment,
	Mid,		// Mid Point local bend
	Point,		// End Point local bend
	ULLoad,	// I Dont kow if this variable should be here 
	ULAxiFor,
	ULTens,
	ULBndMid,
	ULBndPnt,
	ULFbuMid,
	ULFbuPnt float64	
}

type Knot struct {
	Dist,
	Cort,
	ULVx,
	Compresion,
	ULTens,
	Mmnt	float64
}

var Cmp float64 = .9
var Bnd float64 = .9
var E float64 = 29000.0
var Fy float64 = 50.0
var EPkx float64 = 1.0
var IPkx float64 = 0.75
var ky float64 = 0.94
var EPkz float64 = 1.0
var IPkz float64 = 0.75
var EPksz float64 = 1.0
var IPksz float64 = 1.0
type Design struct {
	Pass	bool
	MaxLrxEP,	//
	MaxLrxIP,	//
	Lry,		//
	LrzEP,		//
	LrzIP,		//
	MaxkLrxEP,	//
	MaxkLrxIP,	//
	KLryEP,		//
	KLryIP,		//
	KLrzkLsrzEP,	//
	KLrzkLsrzIP,	//
	MaxkLrEP,
	MaxkLrIP	float64
	TopChord	Chords
	BotChord	Chords
	BotSlend	Slender
}

type Chords []Chord
type Chord struct {
	MaxkLr,
	Kl_rx, Kl_ry, Kl_rz,
	Fex, Fez,
	Fcr_x, Fcr_z,
	Fau, Fbu, Fcr,
	Cm, FauFc, ratio,
	Unity, Stress, FauPuA,
	BndUnity float64
}


type Slender struct {
	MaxLrxEP,
	LryEP,
	LrzLsrzEP,
	MaxkLrxEP,
	KLryEP,
	KLrzkLsrzEP,
	MaxkLrEP,  

	MaxLrxIP,
	LryIP,
	LrzLsrzIP,
	MaxkLrxIP,
	KLryIP,
	KLrzkLsrzIP,
	MaxkLrIP float64
}

func moment(a, b, c, w float64) float64 {
	M1 := math.Pow(a, 4) * math.Pow(w, 2)
	M2 := 4 * math.Pow(a, 2) * b * w
	M3 := 4 * math.Pow(a, 2) * c * w
	M4 := 4 * math.Pow(b, 2)
	M5 := 8 * b * c
	M6 := 4 * math.Pow(c, 2)
	M7 := 8 * math.Pow(a, 2) * w
	M := (M1 + M2 + M3 + M4 - M5 + M6) / M7
	return M
}
