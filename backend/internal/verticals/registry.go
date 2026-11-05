package verticals

import (
	"github.com/Langa2007/POS/backend/internal/kernel/module"
	"github.com/Langa2007/POS/backend/internal/verticals/bookshop"
	"github.com/Langa2007/POS/backend/internal/verticals/dealership"
	"github.com/Langa2007/POS/backend/internal/verticals/fuelstation"
	"github.com/Langa2007/POS/backend/internal/verticals/garage"
	"github.com/Langa2007/POS/backend/internal/verticals/hardware"
	"github.com/Langa2007/POS/backend/internal/verticals/hospitality"
	"github.com/Langa2007/POS/backend/internal/verticals/membership"
	"github.com/Langa2007/POS/backend/internal/verticals/salon"
)

func All() []module.Module {
	return []module.Module{
		&bookshop.Module{}, &dealership.Module{}, &fuelstation.Module{},
		&garage.Module{}, &hardware.Module{}, &hospitality.Module{},
		&membership.Module{}, &salon.Module{},
	}
}

func ByID(id string) (module.Module, bool) {
	for _, m := range All() {
		if m.ID() == id {
			return m, true
		}
	}
	return nil, false
}
