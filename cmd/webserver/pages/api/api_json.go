package api

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/getkin/kin-openapi/openapi3"
)

func init() {

	err := openapi3.NewSwaggerLoader().ResolveRefsIn(Swagger, nil)
	if err != nil {
		log.Err(err)
	}
}

var (
	// This is here because oapi-codegen will not generate params using $ref
	offsetParam = openapi3.NewQueryParameter("offset").WithSchema(openapi3.NewIntegerSchema().WithDefault(0).WithMin(0))
	limitParam  = openapi3.NewQueryParameter("limit").WithSchema(openapi3.NewIntegerSchema().WithDefault(10).WithMin(1).WithMax(100))
)

var Swagger = &openapi3.Swagger{
	OpenAPI: "3.0.0",
	Servers: []*openapi3.Server{
		{URL: config.Config.GameDBDomain.Get() + "/api"},
	},
	Info: &openapi3.Info{
		Title:   "Steam DB API",
		Version: "1",
		Contact: &openapi3.Contact{
			Name:  "Jleagle",
			URL:   "https://gamedb.online/contact",
			Email: "jimeagle@gmail.com",
		},
	},
	Security: openapi3.SecurityRequirements{
		{
			"key-header": []string{},
			"key-query":  []string{},
			"key-cookie": []string{},
		},
	},
	Components: openapi3.Components{
		SecuritySchemes: map[string]*openapi3.SecuritySchemeRef{
			"key-header": {Value: openapi3.NewSecurityScheme().WithName("key").WithType("apiKey").WithIn("header")},
			"key-query":  {Value: openapi3.NewSecurityScheme().WithName("key").WithType("apiKey").WithIn("query")},
			"key-cookie": {Value: openapi3.NewSecurityScheme().WithName("key").WithType("apiKey").WithIn("cookie")},
		},
		Schemas: map[string]*openapi3.SchemaRef{
			"pagination-schema": {
				Value: &openapi3.Schema{
					Required: []string{"offset", "limit", "total", "pagesTotal", "pagesCurrent"},
					Properties: map[string]*openapi3.SchemaRef{
						"offset":       {Value: openapi3.NewInt64Schema()},
						"limit":        {Value: openapi3.NewInt64Schema()},
						"total":        {Value: openapi3.NewInt64Schema()},
						"pagesTotal":   {Value: openapi3.NewIntegerSchema()},
						"pagesCurrent": {Value: openapi3.NewIntegerSchema()},
					},
				},
			},
			"app-schema": {
				Value: &openapi3.Schema{
					Required: []string{"id", "name"},
					Properties: map[string]*openapi3.SchemaRef{
						"id":   {Value: openapi3.NewIntegerSchema()},
						"name": {Value: openapi3.NewStringSchema()},
					},
				},
			},
			"player-schema": {
				Value: &openapi3.Schema{
					Required: []string{"id", "name"},
					Properties: map[string]*openapi3.SchemaRef{
						"id":   {Value: openapi3.NewInt64Schema()},
						"name": {Value: openapi3.NewStringSchema()},
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
		},
		Responses: map[string]*openapi3.ResponseRef{
			"message-response": {
				Value: &openapi3.Response{
					ExtensionProps: openapi3.ExtensionProps{},
					Description:    "Success",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/message-schema",
					}),
				},
			},
			"pagination-response": {
				Value: &openapi3.Response{
					Description: "Page information",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/pagination-schema",
					}),
				},
			},
			"app-response": {
				Value: &openapi3.Response{
					Description: "An app",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/app-schema",
					}),
				},
			},
			"apps-response": {
				Value: &openapi3.Response{
					Description: "List of apps",
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
					Description: "A player",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/player-schema",
					}),
				},
			},
			"players-response": {
				Value: &openapi3.Response{
					Description: "List of players",
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
		"/apps": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Summary: "List Apps",
				Parameters: openapi3.Parameters{
					{Value: offsetParam},
					{Value: limitParam},
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
				},
			},
		},
		"/apps/{id}": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Summary: "Retrieve App",
				Parameters: openapi3.Parameters{
					{
						Value: openapi3.NewPathParameter("id").WithRequired(true).WithSchema(openapi3.NewInt32Schema().WithMin(1)),
					},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/app-response",
					},
				},
			},
		},
		"/players": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Summary: "List Players",
				Parameters: openapi3.Parameters{
					{Value: offsetParam},
					{Value: limitParam},
					{Value: openapi3.NewQueryParameter("continent").WithSchema(openapi3.NewArraySchema().WithMaxItems(3).WithItems(openapi3.NewStringSchema().WithMaxLength(2)))},
					{Value: openapi3.NewQueryParameter("country").WithSchema(openapi3.NewArraySchema().WithMaxItems(3).WithItems(openapi3.NewStringSchema().WithMaxLength(2)))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/players-response",
					},
				},
			},
		},
		"/players/{id}": &openapi3.PathItem{
			Post: &openapi3.Operation{
				Summary: "Update Player",
				Parameters: openapi3.Parameters{
					{
						Value: openapi3.NewPathParameter("id").WithRequired(true).WithSchema(openapi3.NewInt64Schema().WithMaxLength(2)),
					},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
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
