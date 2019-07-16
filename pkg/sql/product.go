package sql

import (
	"errors"
	"strconv"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
)

type ProductInterface interface {
	GetID() int
	GetProductType() helpers.ProductType
	GetName() string
	GetIcon() string
	GetPrice(code steam.ProductCC) (price ProductPrice, err error)
	GetPrices() (prices ProductPrices, err error)
	GetPath() string
	GetType() string
}

var ErrMissingCountryCode = errors.New("invalid code")

//
type ProductPrices map[steam.ProductCC]ProductPrice

func (p *ProductPrices) AddPriceFromPackage(code steam.ProductCC, prices steam.PackageDetailsBody) {

	if prices.Data.Price.Currency == "" {
		prices.Data.Price.Currency = helpers.GetProdCC(code).CurrencyCode
	}

	(*p)[code] = ProductPrice{
		Currency:        prices.Data.Price.Currency,
		Initial:         prices.Data.Price.Initial,
		Final:           prices.Data.Price.Final,
		DiscountPercent: prices.Data.Price.DiscountPercent,
		Individual:      prices.Data.Price.Individual,
	}
}

func (p *ProductPrices) AddPriceFromApp(code steam.ProductCC, prices steam.AppDetailsBody) {

	if prices.Data.PriceOverview.Currency == "" {
		prices.Data.PriceOverview.Currency = helpers.GetProdCC(code).CurrencyCode
	}

	(*p)[code] = ProductPrice{
		Currency:        prices.Data.PriceOverview.Currency,
		Initial:         prices.Data.PriceOverview.Initial,
		Final:           prices.Data.PriceOverview.Final,
		DiscountPercent: prices.Data.PriceOverview.DiscountPercent,
	}
}

func (p ProductPrices) Get(code steam.ProductCC) (price ProductPrice, err error) {
	if val, ok := p[code]; ok {
		return val, err
	}
	return price, ErrMissingCountryCode
}

//
type ProductPrice struct {
	Currency        steam.CurrencyCode `json:"currency"`
	Initial         int                `json:"initial"`
	Final           int                `json:"final"`
	DiscountPercent int                `json:"discount_percent"`
	Individual      int                `json:"individual"`
}

func (p ProductPrice) GetDiscountPercent() string {
	return strconv.Itoa(p.DiscountPercent) + "%"
}

func (p ProductPrice) GetCountryName(code steam.ProductCC) string {
	return helpers.GetProdCC(code).Name
}

func (p ProductPrice) GetFlag(code steam.ProductCC) string {
	return "/assets/img/flags/" + helpers.GetProdCC(code).GetFlag() + ".png"
}

func (p ProductPrice) GetInitial() string {
	return p.format(p.Initial)
}

func (p ProductPrice) GetFinal() string {
	return p.format(p.Final)
}

func (p ProductPrice) GetIndividual() string {
	return p.format(p.Individual)
}

func (p ProductPrice) format(value int) string {
	if p.Currency == "" {
		return "-"
	}
	return helpers.FormatPrice(p.Currency, value)
}
