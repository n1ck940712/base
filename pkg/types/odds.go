package types

import (
	"errors"
	"fmt"
	"math"
)

const OddsUSEven = Odds(100)

func (odds Odds) Ptr() *float64 {
	return (*float64)(&odds)
}

func (odds Odds) String(precision ...int) string {
	if len(precision) > 0 {
		return fmt.Sprintf("%.*f", precision[0], odds)
	}
	return fmt.Sprint(odds)
}

func (odds Odds) EuroToUK(precision int) Odds {
	if err := odds.validateEuro(); err != nil {
		return oddsNaN()
	}
	return (odds - 1).Precision(precision)
}

func (odds Odds) UKToEuro(precision int) Odds {
	if err := odds.validateUK(); err != nil {
		return oddsNaN()
	}
	return (odds + 1).Precision(precision)
}

func (odds Odds) EuroToUS(precision int) Odds {
	if err := odds.validateEuro(); err != nil {
		return oddsNaN()
	}
	if odds == 2 {
		return OddsUSEven
	}
	if odds < 2 {
		return (100 / (1 - odds)).Precision(precision)
	}
	return (100 * (odds - 1)).Precision(precision)
}

func (odds Odds) USToEuro(precision int) Odds {
	if err := odds.validateUS(); err != nil {
		return oddsNaN()
	}
	if odds == OddsUSEven {
		return 2
	}
	if odds < 0 {
		return (1 - (100 / odds)).Precision(precision)
	} else if odds > 0 {
		return (1 + (odds / 100)).Precision(precision)
	}
	return Odds(math.NaN())
}

func (odds Odds) EuroToMalay(precision int) Odds {
	if err := odds.validateEuro(); err != nil {
		return oddsNaN()
	}
	if odds > 2 {
		return (1 / (1 - odds)).Precision(precision)
	}
	return (odds - 1).Precision(precision)
}

func (odds Odds) MalayToEuro(precision int) Odds {
	if err := odds.validateMalay(); err != nil {
		return oddsNaN()
	}
	if odds < 0 {
		return (1 - (1 / odds)).Precision(precision)
	}
	return (1 + odds).Precision(precision)
}

func (odds Odds) EuroToHK(precision int) Odds {
	if err := odds.validateEuro(); err != nil {
		return oddsNaN()
	}
	return (odds - 1).Precision(precision)
}

func (odds Odds) HKToEuro(precision int) Odds {
	if err := odds.validateHK(); err != nil {
		return oddsNaN()
	}
	return (odds + 1).Precision(precision)
}

func (odds Odds) EuroToIndo(precision int) Odds {
	if err := odds.validateEuro(); err != nil {
		return oddsNaN()
	}
	if odds < 2 {
		return (1 / (1 - odds)).Precision(precision)
	}
	return (odds - 1).Precision(precision)
}

func (odds Odds) IndoToEuro(precision int) Odds {
	if err := odds.validateIndo(); err != nil {
		return oddsNaN()
	}
	if odds < 0 {
		return (1 - (1 / odds)).Precision(precision)
	}
	return (1 + odds).Precision(precision)
}

func (odds Odds) validateEuro() error {
	if odds == oddsNaN() {
		return errors.New("euro odds must not be NaN")
	}
	if odds <= 1 {
		return errors.New("euro odds must not be <= 1")
	}
	return nil
}

func (odds Odds) validateMalay() error {
	if odds == oddsNaN() {
		return errors.New("malay odds must not be NaN")
	}
	if odds > 1 {
		return errors.New("malay odds must not be > 1")
	}
	if odds == 0 {
		return errors.New("malay odds must not be 0")
	}
	return nil
}

func (odds Odds) validateUK() error {
	if odds <= 0 {
		return errors.New("uk odds must not be <= 0")
	}
	return nil
}

func (odds Odds) validateHK() error {
	if odds <= 0 {
		return errors.New("hk odds must not be <= 0")
	}
	return nil
}

func (odds Odds) validateIndo() error {
	if odds < 1 && odds > -1 {
		return errors.New("indo odds must not be < 1 and > -1")
	}
	return nil
}

func (odds Odds) validateUS() error {
	if odds < OddsUSEven {
		return errors.New("us odds must not be < 100")
	}
	return nil
}

func (odds Odds) Precision(precision int) Odds {
	return odds.Fixed(precision)
}

func (odds Odds) Round(precision int) Odds {
	return Odds(Float(odds).Round(precision))
}

func (odds Odds) Fixed(precision int) Odds {
	return Odds(Float(odds).Fixed(precision))
}

func oddsNaN() Odds {
	return Odds(math.NaN())
}
