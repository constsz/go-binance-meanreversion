package system

import "directionalMaker/ta"

type StrategyModes map[string]StrategyMode

var strategyModes = StrategyModes{
	"DEFAULT": ModeVolatileAggressive,
}

type StrategyMode struct {
	Name              string
	TP                float64
	SL                float64
	BetterBandsParams ta.BetterBandsParams
	TrendMoveParams   ta.TrendMoveParams
	Timeframe         int
	QtyUsd            float64
}

// ------------------------------------------------------
// STRATEGY MODES

var ModeVolatileAggressive = StrategyMode{
	Name: "ModeVolatileAggressive",
	TP:   0.15,
	SL:   0.2,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        50,
		ATRMultiplier:      4,
		BandMult2:          2, // 3
		BandMult3:          2, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 2, // 2
		MidlineBoostParam: 2, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    300, // 120
		MovePeriod:     50,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    6,
}

var ModeHftWide = StrategyMode{
	Name: "ModeHftWide",
	TP:   0.23,
	SL:   0.4,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      10,
		BandMult2:          1.7, // 3
		BandMult3:          2.2, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 2,   // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    160, // 120
		MovePeriod:     40,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftModerateTightTpSl = StrategyMode{
	Name: "ModeHftModerateTightTpSl",
	TP:   0.18,
	SL:   0.35,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      8,
		BandMult2:          2,   // 3
		BandMult3:          2.6, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.8, // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    160, // 120
		MovePeriod:     40,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftModerateWiderTpTighterSl = StrategyMode{
	Name: "ModeHftModerateWiderTpTighterSl",
	TP:   0.30,
	SL:   0.4,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      8,
		BandMult2:          2,   // 3
		BandMult3:          2.6, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.8, // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    160, // 120
		MovePeriod:     40,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftModerateWiderTpSl = StrategyMode{
	Name: "ModeHftModerateWiderTpSl",
	TP:   0.30,
	SL:   0.5,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      8,
		BandMult2:          2,   // 3
		BandMult3:          2.6, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.8, // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    160, // 120
		MovePeriod:     40,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftModerateWiderTpSlFast = StrategyMode{
	Name: "ModeHftModerateWiderTpSlFast",
	TP:   0.30,
	SL:   0.5,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      8,
		BandMult2:          2,   // 3
		BandMult3:          2.6, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.8, // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    120, // 120
		MovePeriod:     30,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftModerateWiderTpSlSlower = StrategyMode{
	Name: "ModeHftModerateWiderTpSlSlower",
	TP:   0.30,
	SL:   0.5,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      8,
		BandMult2:          2,   // 3
		BandMult3:          2.6, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.8, // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    200, // 120
		MovePeriod:     46,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftModerate_Narrow = StrategyMode{
	Name: "ModeHftModerate_MoreNarrow",
	TP:   0.3,
	SL:   0.45,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      4,
		BandMult2:          1.5, // 3
		BandMult3:          2,   // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.8, // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    160, // 120
		MovePeriod:     40,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftModerate_NarrowAndFaster = StrategyMode{
	Name: "ModeHftModerate_NarrowAndFaster",
	TP:   0.3,
	SL:   0.45,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      4,
		BandMult2:          1.5, // 3
		BandMult3:          2,   // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.8, // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    120, // 120
		MovePeriod:     30,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftModerate_NarrowAndSlower = StrategyMode{
	Name: "ModeHftModerate_NarrowAndFaster",
	TP:   0.3,
	SL:   0.45,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        10,
		ATRMultiplier:      4,
		BandMult2:          1.5, // 3
		BandMult3:          2,   // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.8, // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    240, // 120
		MovePeriod:     50,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftModerate_TwiceNarrow = StrategyMode{
	Name: "ModeHftModerate_MoreNarrow",
	TP:   0.3,
	SL:   0.45,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      2,
		BandMult2:          1, // 3
		BandMult3:          1, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.8, // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    160, // 120
		MovePeriod:     40,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftModerate_TwiceNarrowAndFaster = StrategyMode{
	Name: "ModeHftModerate_TwiceNarrowAndFaster",
	TP:   0.3,
	SL:   0.45,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      2,
		BandMult2:          1, // 3
		BandMult3:          1, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.8, // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    120, // 120
		MovePeriod:     30,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftModerate_TwiceNarrowAndSlower = StrategyMode{
	Name: "ModeHftModerate_NarrowAndFaster",
	TP:   0.3,
	SL:   0.45,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        10,
		ATRMultiplier:      2,
		BandMult2:          1, // 3
		BandMult3:          1, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.8, // 2
		MidlineBoostParam: 1.5, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    240, // 120
		MovePeriod:     50,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftNarrow = StrategyMode{
	Name: "ModeHftNarrow",
	TP:   0.21,
	SL:   0.35,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      1,
		BandMult2:          1,   // 3
		BandMult3:          1.5, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.2, // 2
		MidlineBoostParam: 1.1, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    120, // 120
		MovePeriod:     30,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

var ModeHftCalm = StrategyMode{
	Name: "ModeHftCalm",
	TP:   0.2,
	SL:   0.35,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        7,
		ATRMultiplier:      4,
		BandMult2:          1,   // 3
		BandMult3:          1.5, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.7, // 2
		MidlineBoostParam: 1.4, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    160, // 120
		MovePeriod:     40,  // 60?
		TrendSmoothing: 21,  // 21
		MoveSmoothing:  14,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

//------------------------------------------------------
//SCALPING

var ModeSlowCalm = StrategyMode{
	Name: "ModeSlowCalm",
	TP:   0.36,
	SL:   0.5,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        10,
		ATRMultiplier:      1.5,
		BandMult2:          1, // 3
		BandMult3:          1, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.7, // 2
		MidlineBoostParam: 1.3, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    400, // 120
		MovePeriod:     60,  // 60?
		TrendSmoothing: 42,  // 21
		MoveSmoothing:  28,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}

// ------------------------------------------------------
// DEBUG

var ModeDebug = StrategyMode{
	TP: 0.3,
	SL: 0.4,
	BetterBandsParams: ta.BetterBandsParams{
		BandsPeriod:        10,
		ATRMultiplier:      1,
		BandMult2:          1, // 3
		BandMult3:          1, // 3
		MoveBoostParameter: 0.0005,
		// lower = stronger
		MidlineTrendParam: 1.7, // 2
		MidlineBoostParam: 1.3, // 2
	},
	TrendMoveParams: ta.TrendMoveParams{
		TrendPeriod:    400, // 120
		MovePeriod:     60,  // 60?
		TrendSmoothing: 42,  // 21
		MoveSmoothing:  28,  // 14
	},
	Timeframe: 20,
	QtyUsd:    100,
}
