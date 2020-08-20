package api

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/getkin/kin-openapi/openapi3"
)

const (
	tagGame   = "Games"
	tagPlayer = "Players"
)

func stringPointer(s string) *string {
	return &s
}

var (
	apiKeySchema = openapi3.NewStringSchema().WithPattern("^[0-9A-Z]{20}$")

	keyGetParam  = openapi3.NewQueryParameter("key").WithSchema(apiKeySchema).WithRequired(true)
	keyPostParam = openapi3.NewHeaderParameter("key").WithSchema(apiKeySchema).WithRequired(true)

	// Schemas
	priceSchema = &openapi3.Schema{
		Required: []string{"currency", "initial", "final", "discountPercent", "individual", "free"},
		Properties: map[string]*openapi3.SchemaRef{
			"currency":        {Value: openapi3.NewStringSchema()},
			"initial":         {Value: openapi3.NewInt32Schema()},
			"final":           {Value: openapi3.NewInt32Schema()},
			"discountPercent": {Value: openapi3.NewInt32Schema()},
			"individual":      {Value: openapi3.NewInt32Schema()},
			"free":            {Value: openapi3.NewBoolSchema()},
		},
	}
)

var SwaggerGameDB = &openapi3.Swagger{
	OpenAPI: "3.0.0",
	Servers: []*openapi3.Server{
		{URL: "https://api.gamedb.online"},
	},
	ExternalDocs: &openapi3.ExternalDocs{
		URL: config.Config.GameDBDomain.Get() + "/api/gamedb",
	},
	Info: &openapi3.Info{
		Title:          "Game DB API",
		Version:        "1.0.0",
		TermsOfService: config.Config.GameDBDomain.Get() + "/terms",
		Contact: &openapi3.Contact{
			Name: "Jleagle",
			URL:  config.Config.GameDBDomain.Get() + "/contact",
		},
		ExtensionProps: openapi3.ExtensionProps{Extensions: map[string]interface{}{
			"x-logo": config.Config.GameDBDomain.Get() + "/assets/img/sa-bg-192x192.png",
		}},
	},
	Tags: openapi3.Tags{
		&openapi3.Tag{Name: tagGame},
		&openapi3.Tag{Name: tagPlayer},
	},
	Security: openapi3.SecurityRequirements{
		openapi3.NewSecurityRequirement().Authenticate("key-header"),
		openapi3.NewSecurityRequirement().Authenticate("key-query"),
	},
	Components: openapi3.Components{
		SecuritySchemes: map[string]*openapi3.SecuritySchemeRef{
			"key-header": {Value: openapi3.NewSecurityScheme().WithName("key").WithType("apiKey").WithIn("header")},
			"key-query":  {Value: openapi3.NewSecurityScheme().WithName("key").WithType("apiKey").WithIn("query")},
		},
		Parameters: map[string]*openapi3.ParameterRef{
			"limit-param": {
				Value: openapi3.NewQueryParameter("limit").WithSchema(openapi3.NewIntegerSchema().WithDefault(10).WithMin(1).WithMax(1000)),
			},
			"offset-param": {
				Value: openapi3.NewQueryParameter("offset").WithSchema(openapi3.NewIntegerSchema().WithDefault(0).WithMin(0)),
			},
			"order-param-asc": {
				Value: openapi3.NewQueryParameter("order").WithSchema(openapi3.NewStringSchema().WithEnum("asc", "desc").WithDefault("asc")),
			},
			"order-param-desc": {
				Value: openapi3.NewQueryParameter("order").WithSchema(openapi3.NewStringSchema().WithEnum("asc", "desc").WithDefault("desc")),
			},
		},
		Schemas: map[string]*openapi3.SchemaRef{
			"pagination-schema": {
				Value: &openapi3.Schema{
					Required: []string{"offset", "limit", "total", "pagesTotal", "pagesCurrent"},
					Properties: map[string]*openapi3.SchemaRef{
						"offset":       {Value: openapi3.NewInt64Schema()},
						"limit":        {Value: openapi3.NewInt64Schema()},
						"total":        {Value: openapi3.NewInt64Schema()},
						"pagesTotal":   {Value: openapi3.NewInt64Schema()},
						"pagesCurrent": {Value: openapi3.NewInt64Schema()},
					},
				},
			},
			"app-schema": {
				Value: &openapi3.Schema{
					Required: []string{"id", "name", "tags", "genres", "categories", "developers", "publishers", "prices", "players_max", "players_week_max", "players_week_avg", "release_date", "reviews_positive", "reviews_negative", "reviews_score", "metacritic_score"},
					Properties: map[string]*openapi3.SchemaRef{
						"id":               {Value: openapi3.NewIntegerSchema()},
						"name":             {Value: openapi3.NewStringSchema()},
						"tags":             {Value: openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema())},
						"genres":           {Value: openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema())},
						"categories":       {Value: openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema())},
						"developers":       {Value: openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema())},
						"publishers":       {Value: openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema())},
						"prices":           {Value: openapi3.NewArraySchema().WithItems(priceSchema)},
						"players_max":      {Value: openapi3.NewIntegerSchema()},
						"players_week_max": {Value: openapi3.NewIntegerSchema()},
						"players_week_avg": {Value: openapi3.NewFloat64Schema().WithFormat("double")},
						"release_date":     {Value: openapi3.NewInt64Schema()},
						"reviews_positive": {Value: openapi3.NewIntegerSchema()},
						"reviews_negative": {Value: openapi3.NewIntegerSchema()},
						"reviews_score":    {Value: openapi3.NewFloat64Schema().WithFormat("double")},
						"metacritic_score": {Value: openapi3.NewInt32Schema()},
					},
				},
			},
			"player-schema": {
				Value: &openapi3.Schema{
					Required: []string{"id", "name", "avatar", "badges", "comments", "friends", "games", "groups", "level", "playtime", "country", "continent", "state", "vanity_url"},
					Properties: map[string]*openapi3.SchemaRef{
						"id":         {Value: openapi3.NewStringSchema()}, // Too big for int in JS
						"name":       {Value: openapi3.NewStringSchema()},
						"avatar":     {Value: openapi3.NewStringSchema()},
						"badges":     {Value: openapi3.NewIntegerSchema()},
						"comments":   {Value: openapi3.NewIntegerSchema()},
						"friends":    {Value: openapi3.NewIntegerSchema()},
						"games":      {Value: openapi3.NewIntegerSchema()},
						"groups":     {Value: openapi3.NewIntegerSchema()},
						"level":      {Value: openapi3.NewIntegerSchema()},
						"playtime":   {Value: openapi3.NewIntegerSchema()},
						"country":    {Value: openapi3.NewStringSchema()},
						"continent":  {Value: openapi3.NewStringSchema()},
						"state":      {Value: openapi3.NewStringSchema()},
						"vanity_url": {Value: openapi3.NewStringSchema()},
					},
				},
			},
			"message-schema": {
				Value: &openapi3.Schema{
					Required: []string{"message"},
					Properties: map[string]*openapi3.SchemaRef{
						"message": {Value: openapi3.NewStringSchema()},
					},
				},
			},
			"price-schema": {
				Value: priceSchema,
			},
		},
		Responses: map[string]*openapi3.ResponseRef{
			"message-response": {
				Value: &openapi3.Response{
					ExtensionProps: openapi3.ExtensionProps{},
					Description:    stringPointer("Message"),
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/message-schema",
					}),
				},
			},
			"pagination-response": {
				Value: &openapi3.Response{
					Description: stringPointer("Page information"),
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/pagination-schema",
					}),
				},
			},
			"app-response": {
				Value: &openapi3.Response{
					Description: stringPointer("An app"),
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/app-schema",
					}),
				},
			},
			"apps-response": {
				Value: &openapi3.Response{
					Description: stringPointer("List of apps"),
					Content: openapi3.NewContentWithJSONSchema(&openapi3.Schema{
						Description: "List of apps, with pagination",
						Required:    []string{"pagination", "apps"},
						Properties: map[string]*openapi3.SchemaRef{
							"pagination": {
								Ref: "#/components/schemas/pagination-schema",
							},
							"apps": {
								Value: &openapi3.Schema{
									Type: "array",
									Items: &openapi3.SchemaRef{
										Ref: "#/components/schemas/app-schema",
									},
								},
							},
						},
					}),
				},
			},
			"player-response": {
				Value: &openapi3.Response{
					Description: stringPointer("A player"),
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/player-schema",
					}),
				},
			},
			"players-response": {
				Value: &openapi3.Response{
					Description: stringPointer("List of players"),
					Content: openapi3.NewContentWithJSONSchema(&openapi3.Schema{
						Required: []string{"pagination", "players"},
						Properties: map[string]*openapi3.SchemaRef{
							"pagination": {
								Ref: "#/components/schemas/pagination-schema",
							},
							"players": {
								Value: &openapi3.Schema{
									Type: "array",
									Items: &openapi3.SchemaRef{
										Ref: "#/components/schemas/player-schema",
									},
								},
							},
						},
					}),
				},
			},
		},
	},
	Paths: openapi3.Paths{
		"/games": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{tagGame},
				Summary: "List Apps",
				Parameters: openapi3.Parameters{
					{Value: keyGetParam},
					{Ref: "#/components/parameters/offset-param"},
					{Ref: "#/components/parameters/limit-param"},
					{Ref: "#/components/parameters/order-param-desc"},
					{Value: openapi3.NewQueryParameter("sort").WithSchema(openapi3.NewStringSchema())},
					{Value: openapi3.NewQueryParameter("ids").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(100))},
					{Value: openapi3.NewQueryParameter("tags").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("genres").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("categories").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("developers").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("publishers").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("platforms").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema()).WithMaxItems(3))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/apps-response",
					},
					"401": {
						Ref: "#/components/responses/message-response",
					},
					"500": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		"/games/{id}": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{tagGame},
				Summary: "Retrieve App",
				Parameters: openapi3.Parameters{
					{Value: keyGetParam},
					{Value: openapi3.NewPathParameter("id").WithRequired(true).WithSchema(openapi3.NewInt32Schema().WithMin(1))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/app-response",
					},
					"400": {
						Ref: "#/components/responses/message-response",
					},
					"401": {
						Ref: "#/components/responses/message-response",
					},
					"404": {
						Ref: "#/components/responses/message-response",
					},
					"500": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		"/players": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{tagPlayer},
				Summary: "List Players",
				Parameters: openapi3.Parameters{
					{Value: keyGetParam},
					{Ref: "#/components/parameters/offset-param"},
					{Ref: "#/components/parameters/limit-param"},
					{Ref: "#/components/parameters/order-param-desc"},
					{Value: openapi3.NewQueryParameter("sort").WithSchema(openapi3.NewStringSchema().WithEnum("id", "level", "badges", "games", "time", "friends", "comments").WithDefault("id"))},
					{Value: openapi3.NewQueryParameter("continent").WithSchema(openapi3.NewArraySchema().WithMaxItems(3).WithItems(openapi3.NewStringSchema().WithMaxLength(2)))},
					{Value: openapi3.NewQueryParameter("country").WithSchema(openapi3.NewArraySchema().WithMaxItems(3).WithItems(openapi3.NewStringSchema().WithMaxLength(2)))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/players-response",
					},
					"400": {
						Ref: "#/components/responses/message-response",
					},
					"401": {
						Ref: "#/components/responses/message-response",
					},
					"404": {
						Ref: "#/components/responses/message-response",
					},
					"500": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		"/players/{id}": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{tagPlayer},
				Summary: "Retrieve Player",
				Parameters: openapi3.Parameters{
					{Value: keyGetParam},
					{Value: openapi3.NewPathParameter("id").WithRequired(true).WithSchema(openapi3.NewInt64Schema().WithMin(1))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/player-response",
					},
				},
			},
			Post: &openapi3.Operation{
				Tags:    []string{tagPlayer},
				Summary: "Update Player",
				Parameters: openapi3.Parameters{
					{Value: keyPostParam},
					{Value: openapi3.NewPathParameter("id").WithRequired(true).WithSchema(openapi3.NewInt64Schema().WithMaxLength(2))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/message-response",
					},
					"401": {
						Ref: "#/components/responses/message-response",
					},
					"500": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		// "/app - players",
		// "/app - price changes",
		// "/articles",
		// "/bundles",
		// "/bundles",
		// "/bundles/{id}",
		// "/changes",
		// "/groups"
		// "/packages"
		// "/players/{id}/update"
		// "/players/{id}/badges"
		// "/players/{id}/games"
		// "/players/{id}/history"
		// "/stats/Categories"
		// "/stats/Genres"
		// "/stats/Publishers"
		// "/stats/Steam"
		// "/stats/Tags"
	},
}
